// pulumi/src/infrastructure/compute/node-groups/factory.ts

import * as pulumi from "@pulumi/pulumi";
import * as eks from "@pulumi/eks";
import { NodeGroupConfig } from "../../../config/types";
import { NodeGroupBuilder } from "./builder";

export interface NodeGroupFactoryArgs {
  cluster: eks.Cluster;
  nodeRole: pulumi.Input<string>;
  baseName: string;
  tags?: Record<string, string>;
}

/**
 * Factory class for creating EKS managed node groups
 */
export class NodeGroupFactory {
  private readonly cluster: eks.Cluster;
  private readonly nodeRole: pulumi.Input<string>;
  private readonly baseName: string;
  private readonly tags?: Record<string, string>;

  constructor(args: NodeGroupFactoryArgs) {
    this.cluster = args.cluster;
    this.nodeRole = args.nodeRole;
    this.baseName = args.baseName;
    this.tags = args.tags;
  }

  /**
   * Create an EKS managed node group from configuration
   */
  public createNodeGroup(config: NodeGroupConfig): eks.ManagedNodeGroup {
    const builder = new NodeGroupBuilder({
      cluster: this.cluster,
      nodeRole: this.nodeRole,
      nodeGroupName: `${this.baseName}-${config.name}`,
      config,
      tags: this.tags,
    });

    return builder.build();
  }

  /**
   * Create multiple EKS managed node groups from configurations
   */
  public createNodeGroups(configs: NodeGroupConfig[]): eks.ManagedNodeGroup[] {
    return configs.map((config) => this.createNodeGroup(config));
  }
}
