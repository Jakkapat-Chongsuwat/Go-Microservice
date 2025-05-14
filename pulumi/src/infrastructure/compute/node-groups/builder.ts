// pulumi/src/infrastructure/compute/node-groups/builder.ts

import * as pulumi from "@pulumi/pulumi";
import * as eks from "@pulumi/eks";
import { NodeGroupConfig } from "../../../config/types";

export interface NodeGroupBuilderArgs {
  cluster: eks.Cluster;
  nodeRole: pulumi.Input<string>;
  nodeGroupName: string;
  config: NodeGroupConfig;
  tags?: Record<string, string>;
}

/**
 * Builder class for creating an EKS managed node group
 */
export class NodeGroupBuilder {
  private readonly args: NodeGroupBuilderArgs;

  constructor(args: NodeGroupBuilderArgs) {
    this.args = args;
  }

  /**
   * Format taints into the structure expected by AWS
   */
  private formatTaints() {
    const { config } = this.args;

    if (!config.taints || config.taints.length === 0) {
      return undefined;
    }

    return config.taints.map((taint) => ({
      key: taint.key,
      value: taint.value,
      effect: taint.effect,
    }));
  }

  /**
   * Build and return an EKS managed node group
   */
  public build(): eks.ManagedNodeGroup {
    const { cluster, nodeRole, nodeGroupName, config, tags } = this.args;

    const isSpot = config.capacityType === "SPOT";

    const formattedTaints = this.formatTaints();

    return new eks.ManagedNodeGroup(nodeGroupName, {
      cluster,
      nodeRoleArn: nodeRole,
      nodeGroupName,
      instanceTypes: config.instanceTypes,
      diskSize: config.diskSize || 20,
      scalingConfig: {
        desiredSize: config.desiredSize,
        minSize: config.minSize,
        maxSize: config.maxSize,
      },
      labels: config.labels || {},
      taints: formattedTaints,
      capacityType: isSpot ? "SPOT" : "ON_DEMAND",
      tags: {
        ...(tags || {}),
        Name: nodeGroupName,
        NodeGroup: config.name,
        ManagedBy: "pulumi",
        SpotInstance: isSpot ? "true" : "false",
      },
    });
  }
}
