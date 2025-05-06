import * as pulumi from "@pulumi/pulumi";
import { Vpc } from "./vpc";
import { NetworkConfig } from "@config/types";
import { createResourceTags } from "@utils/tagging";
import { createComponentName } from "@utils/naming";

export interface NetworkStackArgs {
  /**
   * VPC CIDR block
   */
  cidrBlock: string;

  /**
   * Availability zones to deploy into
   */
  availabilityZones: string[];

  /**
   * Tags to apply to resources
   */
  tags?: Record<string, string>;
}

/**
 * NetworkStack creates all networking resources
 */
export class NetworkStack {
  /**
   * The VPC resource
   */
  public readonly vpc: Vpc;

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
   * Creates a new network stack
   *
   * @param name Base name for resources
   * @param args Network configuration arguments
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: NetworkStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    this.vpc = new Vpc(
      createComponentName(name, "vpc"),
      {
        config: {
          vpcCidrBlock: args.cidrBlock,
          availabilityZones: args.availabilityZones,
          createPrivateSubnets: true,
          createPublicSubnets: true,
        },
        tags: args.tags,
      },
      opts
    );

    // Export the network resources
    this.vpcId = this.vpc.vpcId;
    this.privateSubnetIds = this.vpc.privateSubnetIds;
    this.publicSubnetIds = this.vpc.publicSubnetIds;
  }
}

export { Vpc };
