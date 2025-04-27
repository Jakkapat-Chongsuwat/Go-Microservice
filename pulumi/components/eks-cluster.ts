// eks-cluster.ts
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as eks from "@pulumi/eks";
import { NodeGroup, NodeGroupArgs, NodeTaint } from "./node-group";

export interface EksClusterNodeGroupConfig extends Omit<NodeGroupArgs, 'clusterName' | 'nodeRoleArn' | 'subnetIds'> {
    // Any additional node group settings specific to your organization
}

export interface EksClusterArgs {
    vpcId: pulumi.Input<string>;
    subnetIds: pulumi.Input<pulumi.Input<string>[]>;
    nodeGroups?: EksClusterNodeGroupConfig[];
    skipDefaultNodeGroup?: boolean;
    version?: string;
    tags?: { [key: string]: string };
}

export class EksCluster extends pulumi.ComponentResource {
    public readonly cluster: eks.Cluster;
    public readonly nodeGroups: NodeGroup[] = [];
    public readonly kubeconfig: pulumi.Output<string>;
    public readonly nodeRole: aws.iam.Role;

    constructor(name: string, args: EksClusterArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:EksCluster", name, {}, opts);
        
        // Create the node IAM role
        this.nodeRole = this.createNodeRole(`${name}-node-role`);
        
        // Create cluster without default node group if requested or if we have custom node groups
        const skipDefaultNodeGroup = args.skipDefaultNodeGroup || (args.nodeGroups && args.nodeGroups.length > 0);
        
        // Create the EKS cluster
        this.cluster = new eks.Cluster(name, {
            vpcId: args.vpcId,
            subnetIds: args.subnetIds,
            skipDefaultNodeGroup: skipDefaultNodeGroup,
            version: args.version,
            tags: args.tags || {
                Environment: pulumi.getStack(),
                ManagedBy: "pulumi",
            },
        }, { parent: this });

        // Create custom node groups if specified
        if (args.nodeGroups && args.nodeGroups.length > 0) {
            for (let i = 0; i < args.nodeGroups.length; i++) {
                const ngConfig = args.nodeGroups[i];
                const ngName = ngConfig.name || `${name}-ng-${i}`;
                
                const nodeGroup = new NodeGroup(ngName, {
                    ...ngConfig,
                    clusterName: this.cluster.eksCluster.name,
                    nodeRoleArn: this.nodeRole.arn,
                    subnetIds: args.subnetIds,
                }, { parent: this });
                
                this.nodeGroups.push(nodeGroup);
            }
        } else if (!skipDefaultNodeGroup) {
            // Create default node group if no custom node groups and not explicitly skipped
            const defaultNodeGroup = new NodeGroup(`${name}-default`, {
                clusterName: this.cluster.eksCluster.name,
                nodeRoleArn: this.nodeRole.arn,
                subnetIds: args.subnetIds,
                minSize: 1,
                desiredSize: 2,
                maxSize: 3,
                instanceTypes: ["t3.medium"],
                tags: args.tags,
            }, { parent: this });
            
            this.nodeGroups.push(defaultNodeGroup);
        }
        
        // Handle kubeconfig properly
        this.kubeconfig = pulumi.output(this.cluster.kubeconfig).apply(config => {
            return typeof config === 'string' ? config : JSON.stringify(config);
        });

        this.registerOutputs({
            cluster: this.cluster,
            nodeGroups: this.nodeGroups,
            kubeconfig: this.kubeconfig,
            nodeRole: this.nodeRole,
        });
    }

    private createNodeRole(name: string): aws.iam.Role {
        // Create the node role
        const role = new aws.iam.Role(name, {
            assumeRolePolicy: JSON.stringify({
                Version: "2012-10-17",
                Statement: [{
                    Action: "sts:AssumeRole",
                    Effect: "Allow",
                    Principal: {
                        Service: "ec2.amazonaws.com",
                    },
                }],
            }),
        }, { parent: this });

        // Attach required policies
        new aws.iam.RolePolicyAttachment(`${name}-AmazonEKSWorkerNodePolicy`, {
            policyArn: "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
            role: role.name,
        }, { parent: this });

        new aws.iam.RolePolicyAttachment(`${name}-AmazonEKS_CNI_Policy`, {
            policyArn: "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
            role: role.name,
        }, { parent: this });

        new aws.iam.RolePolicyAttachment(`${name}-AmazonEC2ContainerRegistryReadOnly`, {
            policyArn: "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
            role: role.name,
        }, { parent: this });

        return role;
    }
}
