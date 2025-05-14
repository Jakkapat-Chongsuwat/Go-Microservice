// pulumi/src/components/eks-cluster.ts

import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as eks from "@pulumi/eks";
import { NodeGroupConfig } from "../config/types";
import { NodeGroupFactory } from "@infrastructure/compute/node-groups/factory";

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
  public readonly nodeRole: aws.iam.Role;
  public readonly nodeInstanceProfile: aws.iam.InstanceProfile;
  public readonly oidcProviderUrl: pulumi.Output<string>;
  public readonly oidcProviderArn: pulumi.Output<string>;
  public readonly nodeGroups: eks.ManagedNodeGroup[] = [];

  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:EksCluster", name, {}, opts);

    // Create node IAM role
    this.nodeRole = this.createNodeRole(name);

    // Create instance profile from the role
    this.nodeInstanceProfile = new aws.iam.InstanceProfile(
      `${name}-node-profile`,
      { role: this.nodeRole.name },
      { parent: this }
    );

    // Create the EKS cluster with OIDC provider
    this.cluster = new eks.Cluster(
      name,
      {
        vpcId: args.vpcId,
        subnetIds: args.subnetIds,
        version: args.version || "1.28",
        skipDefaultNodeGroup: true,
        createOidcProvider: true,
        instanceRoles: [this.nodeRole],
        tags: args.tags,
      },
      { parent: this }
    );

    // Setup essential member variables
    this.oidcProviderArn = this.cluster
      .oidcProviderArn as pulumi.Output<string>;
    this.oidcProviderUrl = this.getOidcProviderUrl();
    this.kubeconfig = pulumi
      .output(this.cluster.kubeconfig)
      .apply((config) => config as string);

    // Create node groups if specified using our factory
    if (args.nodeGroups && args.nodeGroups.length > 0) {
      const nodeGroupFactory = new NodeGroupFactory({
        cluster: this.cluster,
        nodeRole: this.nodeRole.arn,
        baseName: name,
        tags: args.tags,
      });

      this.nodeGroups = nodeGroupFactory.createNodeGroups(args.nodeGroups);
    }

    // Create ALB controller role
    this.albControllerRole = this.createAlbControllerRole(name);

    // Register outputs
    this.registerOutputs({
      kubeconfig: this.kubeconfig,
      albControllerRole: this.albControllerRole,
      clusterName: this.cluster.eksCluster.name,
      oidcProviderUrl: this.oidcProviderUrl,
      oidcProviderArn: this.oidcProviderArn,
      nodeRole: this.nodeRole,
      nodeGroups: this.nodeGroups,
    });
  }

  /**
   * Extract the OIDC provider URL from the ARN
   */
  private getOidcProviderUrl(): pulumi.Output<string> {
    return this.oidcProviderArn.apply((arn) => {
      if (!arn) {
        throw new Error("OIDC Provider ARN is undefined");
      }

      const match = arn.match(/oidc-provider\/(.+)$/);
      if (!match || !match[1]) {
        throw new Error(`Failed to extract OIDC provider URL from ARN: ${arn}`);
      }
      return match[1];
    });
  }

  /**
   * Create IAM role for worker nodes
   */
  private createNodeRole(name: string): aws.iam.Role {
    // Create the IAM role for node groups
    const nodeRole = new aws.iam.Role(
      `${name}-node-role`,
      {
        assumeRolePolicy: aws.iam.assumeRolePolicyForPrincipal({
          Service: "ec2.amazonaws.com",
        }),
      },
      { parent: this }
    );

    // Attach required policies for EKS worker nodes
    const requiredPolicies = [
      "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
      "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
      "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
    ];

    requiredPolicies.forEach((policyArn, index) => {
      new aws.iam.RolePolicyAttachment(
        `${name}-node-policy-${index}`,
        {
          role: nodeRole.name,
          policyArn,
        },
        { parent: this }
      );
    });

    return nodeRole;
  }

  /**
   * Create the IAM role for AWS Load Balancer Controller
   */
  private createAlbControllerRole(name: string): aws.iam.Role {
    const accountId = aws
      .getCallerIdentity()
      .then((identity) => identity.accountId);

    // Create IAM role with OIDC trust relationship
    const role = new aws.iam.Role(
      `${name}-alb-controller`,
      {
        assumeRolePolicy: pulumi
          .all([this.oidcProviderUrl, accountId])
          .apply(([url, account]) =>
            JSON.stringify({
              Version: "2012-10-17",
              Statement: [
                {
                  Effect: "Allow",
                  Principal: {
                    Federated: `arn:aws:iam::${account}:oidc-provider/${url}`,
                  },
                  Action: "sts:AssumeRoleWithWebIdentity",
                  Condition: {
                    StringEquals: {
                      [`${url}:sub`]:
                        "system:serviceaccount:kube-system:aws-load-balancer-controller",
                      [`${url}:aud`]: "sts.amazonaws.com",
                    },
                  },
                },
              ],
            })
          ),
      },
      { parent: this }
    );

    // Attach the ALB controller policy
    this.attachAlbControllerPolicy(name, role);

    return role;
  }

  /**
   * Attach the AWS Load Balancer Controller policy to the role
   */
  private attachAlbControllerPolicy(name: string, role: aws.iam.Role): void {
    // Create policy for ALB controller
    new aws.iam.RolePolicy(
      `${name}-alb-policy`,
      {
        role: role.id,
        policy: JSON.stringify({
          Version: "2012-10-17",
          Statement: [
            {
              Effect: "Allow",
              Action: [
                "iam:CreateServiceLinkedRole",
                "ec2:DescribeAccountAttributes",
                "ec2:DescribeAddresses",
                "ec2:DescribeAvailabilityZones",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeVpcs",
                "ec2:DescribeSubnets",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeInstances",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeTags",
                "ec2:GetCoipPoolUsage",
                "ec2:DescribeCoipPools",
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
                "wafv2:GetWebACL",
                "wafv2:GetWebACLForResource",
                "shield:GetSubscriptionState",
                "shield:DescribeProtection",
                "shield:CreateProtection",
                "shield:DeleteProtection",
              ],
              Resource: "*",
            },
            {
              Effect: "Allow",
              Action: [
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:RevokeSecurityGroupIngress",
                "ec2:CreateSecurityGroup",
                "ec2:CreateTags",
                "ec2:DeleteTags",
                "ec2:ModifyVpcAttribute",
                "ec2:CreateNetworkInterface",
                "ec2:DeleteNetworkInterface",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:CreateTargetGroup",
                "elasticloadbalancing:CreateRule",
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:DeleteListener",
                "elasticloadbalancing:DeleteRule",
                "elasticloadbalancing:DeleteTargetGroup",
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:SetIpAddressType",
                "elasticloadbalancing:SetSecurityGroups",
                "elasticloadbalancing:SetSubnets",
                "elasticloadbalancing:AddTags",
                "elasticloadbalancing:RemoveTags",
                "elasticloadbalancing:ModifyTargetGroup",
                "elasticloadbalancing:ModifyTargetGroupAttributes",
                "elasticloadbalancing:RegisterTargets",
                "elasticloadbalancing:DeregisterTargets",
                "elasticloadbalancing:AddListenerCertificates",
                "elasticloadbalancing:RemoveListenerCertificates",
                "elasticloadbalancing:ModifyRule",
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
