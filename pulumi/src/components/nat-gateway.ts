import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { NatGatewayArgs, EipArgs } from "./types";

/**
 * NatGateway represents an AWS NAT Gateway with its associated Elastic IP.
 *
 * In our city planning analogy, a NAT Gateway functions like a special toll plaza
 * or customs checkpoint that enables one-way traffic flow. It allows resources in
 * private areas of your city (private subnets) to send traffic to the outside world,
 * while preventing unwanted visitors from entering directly.
 *
 * For EKS clusters, NAT Gateways are crucial infrastructure because:
 * 1. They allow private worker nodes to download container images and updates
 * 2. They enable pods in private subnets to call external APIs securely
 * 3. They provide outbound internet access without exposing nodes to inbound attacks
 */
export class NatGateway extends pulumi.ComponentResource {
  /**
   * The underlying AWS NAT Gateway resource
   *
   * Like the actual checkpoint facility with all its infrastructure,
   * this handles the translation between private internal addresses
   * and public internet addresses.
   */
  public readonly natGateway: aws.ec2.NatGateway;

  /**
   * The Elastic IP associated with the NAT Gateway
   *
   * Similar to the official, permanent address assigned to the checkpoint,
   * this static IP is how the outside world recognizes traffic coming from
   * your private resources.
   */
  public readonly elasticIp?: aws.ec2.Eip; // Made optional with ?

  /**
   * Creates a new NAT Gateway in a specific subnet with an optional Elastic IP
   *
   * @param name The unique name for this gateway
   * @param args Configuration arguments including subnet ID and elastic IP options
   * @param opts Optional Pulumi resource options
   */
  constructor(
    name: string,
    args: NatGatewayArgs & {
      /**
       * Whether to create a new Elastic IP for this NAT Gateway
       *
       * Like deciding whether to obtain a new official address for the checkpoint
       * or use an existing one that's already been reserved.
       */
      createElasticIp?: boolean;

      /**
       * Additional configuration for the Elastic IP if created
       *
       * Similar to specifying additional details when registering a new address.
       */
      elasticIpArgs?: EipArgs;

      /**
       * An existing Elastic IP allocation ID to use
       *
       * Like providing the deed to an existing property that will house the checkpoint.
       */
      allocationId?: pulumi.Input<string>;
    },
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:NatGateway", name, {}, opts);

    let allocationId: pulumi.Input<string>;

    // Create elastic IP for the NAT gateway if needed
    // This is like acquiring a new permanent address for your checkpoint
    if (args.createElasticIp !== false) {
      this.elasticIp = new aws.ec2.Eip(
        `${name}-eip`,
        {
          vpc: true, // Indicates this is for use in a VPC, like zoning the address for commercial use
          tags: { ...args.tags, Name: `${name}-eip` }, // Label the address for identification
          ...args.elasticIpArgs, // Apply any additional configuration options
        },
        { parent: this }
      );

      // Use the newly created EIP's allocation ID
      // Like using the deed number of your newly acquired property
      allocationId = this.elasticIp.id;
    } else {
      // Must use an existing allocation ID if not creating a new EIP
      // This is like using a pre-existing address that's already been registered
      if (!args.allocationId) {
        throw new Error(
          `No allocation ID provided for NAT Gateway ${name} and createElasticIp is false`
        );
      }
      allocationId = args.allocationId;
    }

    // Create the NAT Gateway with the appropriate allocation ID
    // This is like constructing the actual checkpoint facility at the designated address
    this.natGateway = new aws.ec2.NatGateway(
      name,
      {
        allocationId, // The address identification, whether new or existing
        subnetId: args.subnetId, // The specific neighborhood where the checkpoint is placed
        tags: { ...args.tags, Name: name }, // Labels for identification and management
      },
      { parent: this }
    );

    // Register outputs, handling the optional elasticIp
    // Like filing the final documentation about this checkpoint in city records
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
