import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { RouteTableArgs, RouteArgs } from "./types";

/**
 * RouteTable represents an AWS Route Table resource.
 *
 * In our city planning analogy, a Route Table is like the system of street signs and
 * traffic directions that tell vehicles how to navigate through the city. It contains
 * the rules that determine where traffic should flow based on its destination.
 *
 * For EKS clusters, route tables are essential because they:
 * 1. Direct pod and service traffic between subnets
 * 2. Determine which traffic goes to the internet vs. staying internal
 * 3. Enable proper segmentation between public and private resources
 */
export class RouteTable extends pulumi.ComponentResource {
  /**
   * The underlying AWS Route Table resource
   *
   * Like the master traffic control plan for a section of the city,
   * this contains all the routing rules for a set of subnets.
   */
  public readonly routeTable: aws.ec2.RouteTable;

  /**
   * Creates a new Route Table within a VPC
   *
   * @param name The unique name for this route table
   * @param args Configuration arguments including VPC ID and tags
   * @param opts Optional Pulumi resource options
   */
  constructor(
    name: string,
    args: RouteTableArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:RouteTable", name, {}, opts);

    // Create the Route Table
    // Like establishing a new set of traffic rules for a district
    this.routeTable = new aws.ec2.RouteTable(
      name,
      {
        vpcId: args.vpcId, // The city this routing system belongs to
        tags: { ...args.tags, Name: name }, // Official labels for identification
      },
      { parent: this }
    );

    // Register the route table as an output
    // Like filing the traffic plan with the city planning office
    this.registerOutputs({
      routeTable: this.routeTable,
    });
  }

  /**
   * Add a route to this route table
   *
   * This is like adding a new directional sign that says
   * "To reach destination X, go via path Y"
   *
   * @param name Unique identifier for this route
   * @param args Configuration including destination and target gateway
   * @returns The created AWS Route resource
   */
  public addRoute(name: string, args: RouteArgs): aws.ec2.Route {
    if (!this.routeTable) {
      throw new Error("Route table is not initialized");
    }

    // Create a new route within this route table
    // Like adding a specific traffic instruction to the master plan
    return new aws.ec2.Route(
      name,
      {
        // The route table this rule belongs to
        routeTableId: this.routeTable.id,

        // Where traffic is trying to go
        // Like "To reach downtown (0.0.0.0/0) or specific district (10.0.1.0/24)"
        destinationCidrBlock: args.destinationCidrBlock,

        // How to get there - via Internet Gateway
        // Like "Take the main highway entrance"
        gatewayId: args.gatewayId,

        // Or via NAT Gateway
        // Like "Take the toll road exit"
        natGatewayId: args.natGatewayId,
      },
      { parent: this }
    );
  }

  /**
   * Associate this route table with a subnet
   *
   * This is like designating which neighborhood will follow
   * this specific set of traffic rules.
   *
   * @param name Unique identifier for this association
   * @param args Configuration including the subnet to associate
   * @returns The created AWS RouteTableAssociation resource
   */
  public associateWithSubnet(
    name: string,
    args: { subnetId: pulumi.Input<string> }
  ): aws.ec2.RouteTableAssociation {
    if (!this.routeTable) {
      throw new Error("Route table is not initialized");
    }

    // Create the association between route table and subnet
    // Like officially posting these traffic rules in a specific neighborhood
    return new aws.ec2.RouteTableAssociation(
      name,
      {
        routeTableId: this.routeTable.id, // The set of rules
        subnetId: args.subnetId, // The neighborhood they apply to
      },
      { parent: this }
    );
  }
}
