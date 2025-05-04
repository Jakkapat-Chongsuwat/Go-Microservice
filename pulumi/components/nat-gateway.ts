import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { NatGatewayArgs, EipArgs } from "./types";

export class NatGateway extends pulumi.ComponentResource {
  public readonly natGateway: aws.ec2.NatGateway;
  public readonly elasticIp?: aws.ec2.Eip; // Make it optional with ?

  constructor(
    name: string,
    args: NatGatewayArgs & {
      createElasticIp?: boolean;
      elasticIpArgs?: EipArgs;
      allocationId?: pulumi.Input<string>; // Make sure this is defined in the args
    },
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:NatGateway", name, {}, opts);

    let allocationId: pulumi.Input<string>;

    // Create elastic IP for the NAT gateway if needed
    if (args.createElasticIp !== false) {
      this.elasticIp = new aws.ec2.Eip(
        `${name}-eip`,
        {
          vpc: true,
          tags: { ...args.tags, Name: `${name}-eip` },
          ...args.elasticIpArgs,
        },
        { parent: this }
      );

      // Use the new EIP's allocation ID
      allocationId = this.elasticIp.id;
    } else {
      // Must use the provided allocation ID if not creating an EIP
      if (!args.allocationId) {
        throw new Error(
          `No allocation ID provided for NAT Gateway ${name} and createElasticIp is false`
        );
      }
      allocationId = args.allocationId;
    }

    // Create the NAT Gateway with the appropriate allocation ID
    this.natGateway = new aws.ec2.NatGateway(
      name,
      {
        allocationId,
        subnetId: args.subnetId,
        tags: { ...args.tags, Name: name },
      },
      { parent: this }
    );

    // Register outputs, handling the optional elasticIp
    const outputs: {
      natGateway: aws.ec2.NatGateway;
      elasticIp?: aws.ec2.Eip;
    } = {
      natGateway: this.natGateway,
    };

    if (this.elasticIp) {
      outputs.elasticIp = this.elasticIp;
    }

    this.registerOutputs(outputs);
  }
}
