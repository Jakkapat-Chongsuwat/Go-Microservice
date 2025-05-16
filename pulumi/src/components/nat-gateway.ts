import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { EipArgs, BaseResourceArgs } from "./types";

export interface NatGatewayComponentArgs extends BaseResourceArgs {
  subnetId: pulumi.Input<string>;
  createElasticIp?: boolean;
  elasticIpArgs?: EipArgs;
  allocationId?: pulumi.Input<string>;
}

export class NatGateway extends pulumi.ComponentResource {
  public readonly natGateway: aws.ec2.NatGateway;
  public readonly elasticIp?: aws.ec2.Eip;

  constructor(
    name: string,
    args: NatGatewayComponentArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:NatGateway", name, {}, opts);

    let allocationId: pulumi.Input<string>;

    if (args.createElasticIp ?? true) {
      this.elasticIp = new aws.ec2.Eip(
        `${name}-eip`,
        {
          domain: "vpc",
          tags: { ...(args.tags ?? {}), Name: `${name}-eip` },
          ...args.elasticIpArgs,
        },
        { parent: this }
      );
      allocationId = this.elasticIp.id;
    } else {
      if (!args.allocationId) {
        throw new Error(
          `NatGateway ${name}: must provide allocationId if not creating EIP`
        );
      }
      allocationId = args.allocationId;
    }

    this.natGateway = new aws.ec2.NatGateway(
      name,
      {
        subnetId: args.subnetId,
        allocationId,
        tags: { ...(args.tags ?? {}), Name: name },
      },
      { parent: this }
    );

    this.registerOutputs({
      natGateway: this.natGateway,
      elasticIp: this.elasticIp,
    });
  }
}
