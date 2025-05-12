import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as eks from "@pulumi/eks";

export interface NodeGroupConfig {
  name: string;
  instanceTypes: string[];
  desiredSize: number;
  minSize: number;
  maxSize: number;
  diskSize?: number;
  amiType?: string;
  capacityType?: "ON_DEMAND" | "SPOT";
  labels?: { [key: string]: string };
  taints?: {
    key: string;
    value: string;
    effect: "NO_SCHEDULE" | "PREFER_NO_SCHEDULE" | "NO_EXECUTE";
  }[];
  maxUnavailable?: number;
}

export interface EksClusterArgs {
  vpcId: pulumi.Input<string>;
  subnetIds: pulumi.Input<pulumi.Input<string>[]>;
  version?: string;
  nodeGroups?: NodeGroupConfig[];
  tags?: Record<string, string>;
}

export class EksCluster extends pulumi.ComponentResource {
  public readonly cluster: eks.Cluster;
  public readonly kubeconfig: pulumi.Output<string>;
  public readonly albControllerRole: aws.iam.Role;
  public readonly nodeInstanceProfile: aws.iam.InstanceProfile;

  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:EksCluster", name, {}, opts);

    this.cluster = this.createCluster(name, args);
    this.nodeInstanceProfile = this.createNodeRole(name);
    if (args.nodeGroups && args.nodeGroups.length > 0) {
      this.createNodeGroups(name, args.nodeGroups, args.tags);
    }
    this.albControllerRole = this.createAlbControllerRole(name);
    this.kubeconfig = this.cluster.kubeconfig.apply(
      (kubeconfig) => kubeconfig as string
    );

    this.registerOutputs({
      kubeconfig: this.kubeconfig,
      albControllerRole: this.albControllerRole,
      clusterName: this.cluster.eksCluster.name,
    });
  }

  private createNodeRole(name: string): aws.iam.InstanceProfile {
    const nodeRole = new aws.iam.Role(
      `${name}-node-role`,
      {
        assumeRolePolicy: JSON.stringify({
          Version: "2012-10-17",
          Statement: [
            {
              Action: "sts:AssumeRole",
              Principal: { Service: "ec2.amazonaws.com" },
              Effect: "Allow",
            },
          ],
        }),
      },
      { parent: this }
    );

    const policies = [
      { name: "AmazonEKSWorkerNodePolicy", suffix: "WorkerNodePolicy" },
      { name: "AmazonEKS_CNI_Policy", suffix: "CNIPolicy" },
      { name: "AmazonEC2ContainerRegistryReadOnly", suffix: "ECRReadOnly" },
    ];

    policies.forEach((policy) => {
      new aws.iam.RolePolicyAttachment(
        `${name}-node-${policy.suffix}`,
        {
          policyArn: `arn:aws:iam::aws:policy/${policy.name}`,
          role: nodeRole.name,
        },
        { parent: this }
      );
    });

    return new aws.iam.InstanceProfile(
      `${name}-node-profile`,
      {
        role: nodeRole.name,
      },
      { parent: this }
    );
  }

  private createCluster(name: string, args: EksClusterArgs): eks.Cluster {
    return new eks.Cluster(
      name,
      {
        vpcId: args.vpcId,
        subnetIds: args.subnetIds,
        version: args.version || "1.28",
        instanceType: "t3.medium",
        desiredCapacity: 2,
        minSize: 1,
        maxSize: 4,
        nodePublicKey:
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com",
        tags: {
          ...(args.tags || {}),
          "kubernetes.io/cluster/name": name,
        },
        createOidcProvider: true,
        skipDefaultNodeGroup: true,
      },
      { parent: this }
    );
  }

  private createNodeGroups(
    clusterName: string,
    nodeGroups: NodeGroupConfig[],
    tags?: { [key: string]: string }
  ): void {
    nodeGroups.forEach((ng) => {
      const nodeGroupName = `${clusterName}-${ng.name}`;
      const formattedTaints = this.formatTaints(ng.taints);
      const updateConfig: any =
        ng.maxUnavailable !== undefined
          ? { maxUnavailable: ng.maxUnavailable }
          : { maxUnavailable: 1 };

      new eks.NodeGroupV2(
        nodeGroupName,
        {
          cluster: this.cluster,
          instanceTypes: ng.instanceTypes,
          desiredCapacity: ng.desiredSize,
          minSize: ng.minSize,
          maxSize: ng.maxSize,
          labels: ng.labels || {},
          taints: formattedTaints,
          nodeAssociatePublicIpAddress: false,
          nodeRootVolumeSize: ng.diskSize || 20,
          amiType: ng.amiType || "AL2_x86_64",
          capacityType: ng.capacityType || "ON_DEMAND",
          updateConfig: updateConfig,
          instanceProfile: { name: this.nodeInstanceProfile.name },
          tags: {
            ...(tags || {}),
            Name: nodeGroupName,
            ManagedBy: "pulumi",
          },
          ...this.getSpotConfiguration(ng),
        },
        { parent: this }
      );
    });
  }

  private formatTaints(
    taints?: { key: string; value: string; effect: string }[]
  ): { [key: string]: { value: string; effect: string } } | undefined {
    if (!taints) return undefined;

    return taints.reduce((acc, taint) => {
      acc[taint.key] = {
        value: taint.value,
        effect: taint.effect,
      };
      return acc;
    }, {} as { [key: string]: { value: string; effect: string } });
  }

  private getSpotConfiguration(nodeGroup: NodeGroupConfig): any {
    if (nodeGroup.capacityType === "SPOT") {
      return {
        kubeletExtraArgs: "--node-labels=node.kubernetes.io/lifecycle=spot",
      };
    }
    return {};
  }

  private createAlbControllerRole(name: string): aws.iam.Role {
    const oidcProviderUrl = this.getOidcProviderUrl();
    const callerIdentity = aws.getCallerIdentity();

    const albControllerRole = new aws.iam.Role(
      `${name}-alb-controller`,
      {
        assumeRolePolicy: pulumi
          .all([oidcProviderUrl, callerIdentity])
          .apply(([providerUrl, identity]) => {
            if (!providerUrl) {
              throw new Error("OIDC Provider URL is undefined or empty");
            }

            return JSON.stringify({
              Version: "2012-10-17",
              Statement: [
                {
                  Effect: "Allow",
                  Principal: {
                    Federated: `arn:aws:iam::${identity.accountId}:oidc-provider/${providerUrl}`,
                  },
                  Action: "sts:AssumeRoleWithWebIdentity",
                  Condition: {
                    StringEquals: {
                      [`${providerUrl}:sub`]:
                        "system:serviceaccount:kube-system:aws-load-balancer-controller",
                    },
                  },
                },
              ],
            });
          }),
      },
      { parent: this }
    );

    this.attachAlbControllerPolicy(name, albControllerRole);

    return albControllerRole;
  }

  private getOidcProviderUrl(): pulumi.Output<string> {
    return this.cluster.oidcIssuer.apply((issuer) => {
      if (!issuer) {
        return "";
      }

      return issuer.replace("https://", "");
    });
  }

  private attachAlbControllerPolicy(
    baseName: string,
    role: aws.iam.Role
  ): void {
    new aws.iam.RolePolicy(
      `${baseName}-alb-controller-policy`,
      {
        role: role.id,
        policy: JSON.stringify({
          Version: "2012-10-17",
          Statement: [
            {
              Effect: "Allow",
              Action: [
                "ec2:DescribeVpcs",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeInstances",
                "ec2:DescribeSubnets",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeNatGateways",
                "ec2:CreateSecurityGroup",
                "ec2:CreateTags",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:RevokeSecurityGroupIngress",
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:DescribeLoadBalancerAttributes",
                "elasticloadbalancing:DescribeListeners",
                "elasticloadbalancing:DescribeListenerCertificates",
                "elasticloadbalancing:DescribeSSLPolicies",
                "elasticloadbalancing:DescribeRules",
                "elasticloadbalancing:DescribeTargetGroups",
                "elasticloadbalancing:DescribeTargetGroupAttributes",
                "elasticloadbalancing:DescribeTargetHealth",
                "elasticloadbalancing:DescribeTags",
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:CreateRule",
                "elasticloadbalancing:CreateTargetGroup",
                "elasticloadbalancing:ModifyListener",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:ModifyRule",
                "elasticloadbalancing:ModifyTargetGroup",
                "elasticloadbalancing:ModifyTargetGroupAttributes",
                "elasticloadbalancing:RegisterTargets",
                "elasticloadbalancing:DeregisterTargets",
                "elasticloadbalancing:SetIpAddressType",
                "elasticloadbalancing:SetSecurityGroups",
                "elasticloadbalancing:SetSubnets",
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:DeleteTargetGroup",
                "elasticloadbalancing:DeleteListener",
                "elasticloadbalancing:DeleteRule",
                "elasticloadbalancing:AddTags",
                "elasticloadbalancing:RemoveTags",
                "wafv2:GetWebACL",
                "wafv2:GetWebACLForResource",
                "wafv2:AssociateWebACL",
                "wafv2:DisassociateWebACL",
              ],
              Resource: "*",
            },
          ],
        }),
      },
      { parent: this }
    );
  }
}
