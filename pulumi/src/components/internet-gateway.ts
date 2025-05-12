// pulumi/src/components/internet-gateway.ts

import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { InternetGatewayArgs } from "./types";

/**
 * InternetGateway represents an AWS Internet Gateway resource.
 *
 * In the city planning analogy, an Internet Gateway is like the main highway system
 * connecting your city (VPC) to the outside world (internet). Just as highways allow
 * people and goods to enter and exit a city, the Internet Gateway enables traffic to
 * flow between your VPC resources and the public internet.
 *
 * For EKS clusters, the Internet Gateway is essential for:
 * 1. Allowing worker nodes to download updates and container images
 * 2. Enabling public-facing services (like load balancers) to receive external traffic
 * 3. Permitting outbound communication from private resources via NAT Gateways
 */
export class InternetGateway extends pulumi.ComponentResource {
  /**
   * The underlying AWS Internet Gateway resource
   *
   * This is the actual infrastructure component that routes traffic between
   * your VPC and the internet, similar to how major highways and interchanges
   * connect a city to the broader transportation network.
   */
  public readonly gateway: aws.ec2.InternetGateway;

  /**
   * Creates a new Internet Gateway attached to a VPC
   *
   * @param name The unique name for this gateway
   * @param args Configuration arguments including VPC ID and tags
   * @param opts Optional Pulumi resource options
   */
  constructor(
    name: string,
    args: InternetGatewayArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:InternetGateway", name, {}, opts);

    // Create the Internet Gateway resource
    // This is like constructing the city's main interchange with the interstate highway
    this.gateway = new aws.ec2.InternetGateway(
      name,
      {
        // Attach to the specified VPC, just as highways connect to specific cities
        vpcId: args.vpcId,

        // Tag the gateway for identification and management
        // Like placing signs and official markers on the highway infrastructure
        tags: { ...args.tags, Name: name },
      },
      { parent: this }
    );

    // Register the gateway as an output for Pulumi state management
    // Similar to recording the completed highway in official city planning documents
    this.registerOutputs({
      gateway: this.gateway,
    });
  }
}
