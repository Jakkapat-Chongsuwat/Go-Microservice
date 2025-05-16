// src/infrastructure/networking/index.ts

import * as pulumi from "@pulumi/pulumi";
import { createComponentName } from "@utils/naming";
import { Network } from "@components/index";
import { NetworkArgs } from "@components/types";

export interface NetworkStackArgs {
  /** Full CIDR for the VPC */
  cidrBlock: string;

  /** AZs can now be an Output<string[]> */
  availabilityZones: pulumi.Input<string[]>;

  /** Whether to create private subnets */
  createPrivateSubnets?: boolean;

  /** Whether to create public subnets */
  createPublicSubnets?: boolean;

  /** Arbitrary plain tags */
  tags?: Record<string, string>;
}

export class NetworkStack extends pulumi.ComponentResource {
  public readonly network: Network;
  public readonly vpcId: pulumi.Output<string>;
  public readonly privateSubnetIds: pulumi.Output<string>[];
  public readonly publicSubnetIds: pulumi.Output<string>[];

  constructor(
    name: string,
    args: NetworkStackArgs, // ← note Input<string[]> here
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("pkg:networking:NetworkStack", name, {}, opts);

    this.network = new Network(
      createComponentName(name, "network"),
      {
        vpcCidrBlock: args.cidrBlock,
        availabilityZones: args.availabilityZones,
        createPrivateSubnets: args.createPrivateSubnets ?? true,
        createPublicSubnets: args.createPublicSubnets ?? true,
        tags: args.tags,
      } as NetworkArgs,
      { parent: this }
    );

    this.vpcId = this.network.vpc.id;
    this.privateSubnetIds = this.network.privateSubnetIds;
    this.publicSubnetIds = this.network.publicSubnetIds;

    this.registerOutputs({
      vpcId: this.vpcId,
      privateSubnetIds: this.privateSubnetIds,
      publicSubnetIds: this.publicSubnetIds,
    });
  }
}
