// node-group.ts
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

// Define our own interfaces based on AWS API structure
export interface NodeTaint {
    key: string;
    value: string;
    effect: string;  // NO_SCHEDULE, NO_EXECUTE, or PREFER_NO_SCHEDULE
}

export interface NodeGroupLaunchTemplate {
    version: pulumi.Input<string>;
    id?: pulumi.Input<string>;
    name?: pulumi.Input<string>;
}

export interface NodeGroupScalingConfig {
    desiredSize: pulumi.Input<number>;
    maxSize: pulumi.Input<number>;
    minSize: pulumi.Input<number>;
}

export interface NodeGroupUpdateConfig {
    maxUnavailable?: pulumi.Input<number>;
    maxUnavailablePercentage?: pulumi.Input<number>;
}

export interface NodeGroupArgs {
    // Required parameters
    clusterName: pulumi.Input<string>;
    nodeRoleArn: pulumi.Input<string>;
    subnetIds: pulumi.Input<pulumi.Input<string>[]>;
    
    // Optional configuration
    name?: string;
    instanceTypes?: pulumi.Input<pulumi.Input<string>[]>;
    minSize?: number;
    desiredSize?: number;
    maxSize?: number;
    diskSize?: pulumi.Input<number>;
    labels?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
    taints?: pulumi.Input<pulumi.Input<NodeTaint>[]>;
    capacityType?: pulumi.Input<string>;
    amiType?: pulumi.Input<string>;
    releaseVersion?: pulumi.Input<string>;
    
    // Launch template (optional)
    launchTemplate?: pulumi.Input<NodeGroupLaunchTemplate>;
    
    // Update configuration
    maxUnavailable?: pulumi.Input<number>;
    maxUnavailablePercentage?: pulumi.Input<number>;
    
    // Tags
    tags?: pulumi.Input<{[key: string]: pulumi.Input<string>}>;
}

export class NodeGroup extends pulumi.ComponentResource {
    public readonly nodeGroup: aws.eks.NodeGroup;
    
    constructor(name: string, args: NodeGroupArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:NodeGroup", name, {}, opts);
        
        // Set default values
        const nodeGroupName = args.name || name;
        const instanceTypes = args.instanceTypes || ["t3.medium"];
        const minSize = args.minSize || 1;
        const desiredSize = args.desiredSize || 2; 
        const maxSize = args.maxSize || 3;
        
        // Prepare scaling config
        const scalingConfig: NodeGroupScalingConfig = {
            minSize: minSize,
            desiredSize: desiredSize,
            maxSize: maxSize,
        };
        
        // Prepare update config if specified
        const updateConfig: NodeGroupUpdateConfig = {};
        if (args.maxUnavailable !== undefined) {
            updateConfig.maxUnavailable = args.maxUnavailable;
        } else if (args.maxUnavailablePercentage !== undefined) {
            updateConfig.maxUnavailablePercentage = args.maxUnavailablePercentage;
        } else {
            updateConfig.maxUnavailable = 1; // Default to 1 for controlled rolling updates
        }
        
        // Create the node group with strongly typed parameters
        this.nodeGroup = new aws.eks.NodeGroup(name, {
            clusterName: args.clusterName,
            nodeRoleArn: args.nodeRoleArn,
            subnetIds: args.subnetIds,
            nodeGroupName: nodeGroupName,
            
            // Scaling configuration
            scalingConfig: scalingConfig,
            
            // Node configuration
            instanceTypes: instanceTypes,
            diskSize: args.diskSize,
            labels: args.labels,
            taints: args.taints,
            
            // Optional configurations
            amiType: args.amiType,
            capacityType: args.capacityType || "ON_DEMAND",
            releaseVersion: args.releaseVersion,
            
            // Launch template if provided
            launchTemplate: args.launchTemplate,
            
            // Update configuration
            updateConfig: updateConfig,
            
            // Tags
            tags: args.tags,
        }, { parent: this });
        
        this.registerOutputs({
            nodeGroup: this.nodeGroup,
        });
    }
}
