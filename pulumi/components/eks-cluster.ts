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
}

export interface EksClusterArgs {
    vpcId: pulumi.Input<string>;
    subnetIds: pulumi.Input<pulumi.Input<string>[]>;
    version?: string;
    nodeGroups?: NodeGroupConfig[];
    tags?: { [key: string]: string };
}

export class EksCluster extends pulumi.ComponentResource {
    public readonly cluster: eks.Cluster;
    public readonly kubeconfig: pulumi.Output<string>;
    public readonly albControllerRole: aws.iam.Role;

    constructor(name: string, args: EksClusterArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:EksCluster", name, {}, opts);

        // Create the EKS cluster
        this.cluster = new eks.Cluster(name, {
            vpcId: args.vpcId,
            subnetIds: args.subnetIds,
            version: args.version || "1.28",
            instanceType: "t3.medium",
            desiredCapacity: 2,
            minSize: 1,
            maxSize: 4,
            nodePublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com",
            tags: {
                ...args.tags || {},
                "kubernetes.io/cluster/name": name,
            },
            createOidcProvider: true,
            skipDefaultNodeGroup: true,
        }, { parent: this });

        // Create node groups if defined
        if (args.nodeGroups && args.nodeGroups.length > 0) {
            args.nodeGroups.forEach((ng, index) => {
                const nodeGroupName = `${name}-${ng.name}`;
                
                // Convert taints array to the expected map format
                const formattedTaints = ng.taints ? 
                    ng.taints.reduce((acc, taint) => {
                        acc[taint.key] = {
                            value: taint.value,
                            effect: taint.effect,
                        };
                        return acc;
                    }, {} as {[key: string]: {value: string, effect: string}}) : 
                    undefined;

                // Create specific NodeGroup launch options based on capacity type
                const nodeGroupArgs: any = {
                    cluster: this.cluster,
                    instanceType: ng.instanceTypes[0], // Primary instance type
                    desiredCapacity: ng.desiredSize,
                    minSize: ng.minSize,
                    maxSize: ng.maxSize,
                    labels: ng.labels || {},
                    taints: formattedTaints,
                    nodeAssociatePublicIpAddress: false,
                    nodeRootVolumeSize: ng.diskSize || 20,
                    amiType: ng.amiType || "AL2_x86_64",
                    tags: {
                        ...args.tags || {},
                        Name: nodeGroupName,
                    },
                };

                // Add spot instance configuration via kubeletExtraArgs if specified
                if (ng.capacityType === "SPOT") {
                    nodeGroupArgs.kubeletExtraArgs = "--node-labels=node.kubernetes.io/lifecycle=spot";
                }

                new eks.NodeGroupV2(nodeGroupName, nodeGroupArgs, { parent: this });
            });
        }

        // Extract OIDC provider URL from the cluster
        // Using this pattern to handle nested outputs and undefined values safely
        const oidcProviderUrl = this.cluster.core.oidcProvider?.apply(provider => {
            if (!provider) {
                return "";
            }
            // Need to convert Output<string> to string safely
            // Use .apply to transform the output
            return provider.url.apply((url: string) => {
                return url.replace("https://", "");
            });
        });

        // Get AWS account ID
        const callerIdentity = aws.getCallerIdentity();
        
        // Create IAM role for ALB controller
        this.albControllerRole = new aws.iam.Role(`${name}-alb-controller`, {
            assumeRolePolicy: pulumi.all([oidcProviderUrl, callerIdentity]).apply(([providerUrl, identity]) => {
                if (!providerUrl) {
                    throw new Error("OIDC Provider URL is undefined or empty");
                }
                
                return JSON.stringify({
                    Version: "2012-10-17",
                    Statement: [{
                        Effect: "Allow",
                        Principal: {
                            Federated: `arn:aws:iam::${identity.accountId}:oidc-provider/${providerUrl}`
                        },
                        Action: "sts:AssumeRoleWithWebIdentity",
                        Condition: {
                            StringEquals: {
                                [`${providerUrl}:sub`]: "system:serviceaccount:kube-system:aws-load-balancer-controller"
                            }
                        }
                    }]
                });
            })
        }, { parent: this });

        // Attach ALB controller policy to the role
        new aws.iam.RolePolicy(`${name}-alb-controller-policy`, {
            role: this.albControllerRole.id,
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
                            "wafv2:DisassociateWebACL"
                        ],
                        Resource: "*"
                    }
                ]
            })
        }, { parent: this });

        // Properly assign the kubeconfig
        // Using pulumi.interpolate to ensure proper string typing
        this.kubeconfig = pulumi.interpolate`${this.cluster.kubeconfig}`;

        // Register outputs properly
        this.registerOutputs({
            kubeconfig: this.kubeconfig,
            albControllerRole: this.albControllerRole,
            clusterName: this.cluster.eksCluster.name,
        });
    }
}
