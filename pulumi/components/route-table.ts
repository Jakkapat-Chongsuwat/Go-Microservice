import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { RouteTableArgs, RouteArgs } from "./types";

export class RouteTable extends pulumi.ComponentResource {
  public readonly routeTable: aws.ec2.RouteTable;

  constructor(
    name: string,
    args: RouteTableArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:RouteTable", name, {}, opts);

    this.routeTable = new aws.ec2.RouteTable(
      name,
      {
        vpcId: args.vpcId,
        tags: { ...args.tags, Name: name },
      },
      { parent: this }
    );

    this.registerOutputs({
      routeTable: this.routeTable,
    });
  }

  /**
   * Add a route to this route table
   */
  public addRoute(name: string, args: RouteArgs): aws.ec2.Route {
    if (!this.routeTable) {
      throw new Error("Route table is not initialized");
    }

    return new aws.ec2.Route(
      name,
      {
        routeTableId: this.routeTable.id,
        destinationCidrBlock: args.destinationCidrBlock,
        gatewayId: args.gatewayId,
        natGatewayId: args.natGatewayId,
      },
      { parent: this }
    );
  }

  /**
   * Associate this route table with a subnet
   */
  public associateWithSubnet(
    name: string,
    args: { subnetId: pulumi.Input<string> }
  ): aws.ec2.RouteTableAssociation {
    if (!this.routeTable) {
      throw new Error("Route table is not initialized");
    }

    return new aws.ec2.RouteTableAssociation(
      name,
      {
        routeTableId: this.routeTable.id,
        subnetId: args.subnetId,
      },
      { parent: this }
    );
  }
}
