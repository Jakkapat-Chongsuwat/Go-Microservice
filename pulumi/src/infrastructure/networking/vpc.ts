import * as pulumi from "@pulumi/pulumi";
import { Network } from "@components/index";
import { NetworkConfig } from "@config/types";

export interface VpcArgs {
  /**
   * Configuration for the network
   */
  config: NetworkConfig;

  /**
   * Tags to apply to resources
   */
  tags?: Record<string, string>;
}

/**
 * VPC creates and manages a VPC and its associated resources
 */
export class Vpc {
  /**
   * The Network component that manages the VPC
   */
  public readonly network: Network;

  /**
   * The VPC ID
   */
  public readonly vpcId: pulumi.Output<string>;

  /**
   * Private subnet IDs
   */
  public readonly privateSubnetIds: pulumi.Output<string>[];

  /**
   * Public subnet IDs
   */
  public readonly publicSubnetIds: pulumi.Output<string>[];

  /**
   * Creates a new VPC infrastructure
   *
   * @param name The resource name
   * @param args The VPC arguments
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: VpcArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    // Create the network component
    this.network = new Network(
      name,
      {
        vpcCidrBlock: args.config.vpcCidrBlock,
        availabilityZones: args.config.availabilityZones,
        createPrivateSubnets: args.config.createPrivateSubnets,
        createPublicSubnets: args.config.createPublicSubnets,
        tags: args.tags,
      },
      opts
    );

    this.vpcId = this.network.vpc.id;
    this.privateSubnetIds = this.network.privateSubnetIds;
    this.publicSubnetIds = this.network.publicSubnetIds;
  }
}
