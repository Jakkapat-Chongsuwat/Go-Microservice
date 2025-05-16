// src/infrastructure/networking/vpc.ts

import * as pulumi from "@pulumi/pulumi";
import { Network } from "@components/index";
import { NetworkArgs } from "@components/types";

/** Mirror your stack config shape */
export interface NetworkConfig {
  vpcCidrBlock: string;
  availabilityZones?: string[];
  createPrivateSubnets?: boolean;
  createPublicSubnets?: boolean;
}

export class VpcInfra extends pulumi.ComponentResource {
  public readonly network: Network;
  public readonly vpcId: pulumi.Output<string>;
  public readonly privateSubnetIds: pulumi.Output<string>[];
  public readonly publicSubnetIds: pulumi.Output<string>[];

  constructor(name: string, opts?: pulumi.ComponentResourceOptions) {
    // Load JSON/YAML under “network.config” in your Pulumi.<stack>.yaml
    const cfg = new pulumi.Config("network");
    const netCfg = cfg.requireObject<NetworkConfig>("config");

    super("custom:resource:VpcInfra", name, {}, opts);

    // Defaults
    const azs = netCfg.availabilityZones ?? [];
    const createP = netCfg.createPublicSubnets ?? true;
    const createR = netCfg.createPrivateSubnets ?? true;

    // Instantiate the Network component with plain values only
    this.network = new Network(
      name,
      {
        vpcCidrBlock: netCfg.vpcCidrBlock,
        availabilityZones: azs,
        createPrivateSubnets: createR,
        createPublicSubnets: createP,
        tags: { Environment: pulumi.getStack() },
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

export default VpcInfra;
