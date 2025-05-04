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

    // Create a properly typed tags object
    // Use Record<string, string> to allow any string keys
    let tags: Record<string, string> = {
      ...(args.tags || {}),
      Name: name,
    };

    // Add Kubernetes-specific tags based on subnet type
    if (args.type === "public") {
      tags["kubernetes.io/role/elb"] = "1";
    } else if (args.type === "private") {
      tags["kubernetes.io/role/internal-elb"] = "1";
    }

    this.subnet = new aws.ec2.Subnet(
      name,
      {
        vpcId: args.vpcId,
        cidrBlock: args.cidrBlock,
        availabilityZone: args.availabilityZone,
        mapPublicIpOnLaunch: args.mapPublicIpOnLaunch ?? args.type === "public",
        tags,
      },
      { parent: this }
    );

    this.registerOutputs({
      subnet: this.subnet,
    });
  }
}
