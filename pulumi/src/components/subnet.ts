import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { SubnetArgs } from "./types";

/**
 * Subnet represents an AWS Subnet resource.
 *
 * In our city planning analogy, a subnet is like a neighborhood or district within
 * the city (VPC). Each subnet has its own characteristics, regulations, and purpose,
 * just like residential, commercial, or industrial zones in a city.
 *
 * For EKS clusters, subnets are foundational infrastructure because:
 * 1. They provide logical isolation for different parts of your application
 * 2. They span different Availability Zones for high availability
 * 3. They have distinct routing policies (public vs. private)
 * 4. Kubernetes uses special tags to identify which subnets to use for load balancers
 */
export class Subnet extends pulumi.ComponentResource {
  /**
   * The underlying AWS Subnet resource
   *
   * This is the actual network segment where EC2 instances, EKS nodes,
   * and other resources will reside, similar to how a particular
   * neighborhood contains buildings and infrastructure.
   */
  public readonly subnet: aws.ec2.Subnet;

  /**
   * Creates a new Subnet within a VPC
   *
   * @param name The unique name for this subnet
   * @param args Configuration arguments including VPC ID, CIDR block, and type
   * @param opts Optional Pulumi resource options
   */
  constructor(
    name: string,
    args: SubnetArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:Subnet", name, {}, opts);

    // Create a properly typed tags object
    // Tags are like official markers that identify properties of this area
    let tags: Record<string, string> = {
      ...(args.tags || {}),
      Name: name, // The official name of this neighborhood
    };

    // Add Kubernetes-specific tags based on subnet type
    // These special tags are like zoning designations that tell Kubernetes
    // which areas are suitable for certain types of infrastructure
    if (args.type === "public") {
      // Tag for public subnets - tells AWS to create public load balancers here
      // Like designating an area for public-facing commercial buildings
      tags["kubernetes.io/role/elb"] = "1";
    } else if (args.type === "private") {
      // Tag for private subnets - tells AWS to create internal load balancers here
      // Like designating an area for internal-only facilities
      tags["kubernetes.io/role/internal-elb"] = "1";
    }

    // Create the subnet resource
    // This is like formally establishing a new neighborhood with its boundaries and rules
    this.subnet = new aws.ec2.Subnet(
      name,
      {
        vpcId: args.vpcId, // The city this neighborhood belongs to

        cidrBlock: args.cidrBlock, // The address range (size and location of this district)

        availabilityZone: args.availabilityZone, // Which physical zone this area is in

        // Whether instances get public IPs by default
        // Like deciding if buildings automatically get listed in public directories
        mapPublicIpOnLaunch: args.mapPublicIpOnLaunch ?? args.type === "public",

        // Apply all our identification tags
        tags,
      },
      { parent: this }
    );

    // Register the subnet as an output
    // Like recording this neighborhood in official city records
    this.registerOutputs({
      subnet: this.subnet,
    });
  }
}
