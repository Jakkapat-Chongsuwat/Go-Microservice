// --- pulumi/src/components/subnet.ts ---
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { SubnetArgs } from "./types";

export class Subnet extends pulumi.ComponentResource {
  public readonly subnet: aws.ec2.Subnet;

  constructor(
    name: string,
    args: SubnetArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:Subnet", name, {}, opts);

    // Compute merged tags as an Input<Record<string,string>>
    const mergedTags = pulumi.output(args.tags ?? {}).apply((t) => ({
      ...t,
      Name: name,
      ...(args.type === "public" ? { "kubernetes.io/role/elb": "1" } : {}),
      ...(args.type === "private"
        ? { "kubernetes.io/role/internal-elb": "1" }
        : {}),
    }));

    // Create the subnet
    this.subnet = new aws.ec2.Subnet(
      name,
      {
        vpcId: args.vpcId,
        cidrBlock: args.cidrBlock,
        availabilityZone: args.availabilityZone,
        mapPublicIpOnLaunch: args.mapPublicIpOnLaunch ?? args.type === "public",
        tags: mergedTags, // <-- pass the Input-mapped tags
      },
      { parent: this }
    );

    this.registerOutputs({ subnet: this.subnet });
  }
}
