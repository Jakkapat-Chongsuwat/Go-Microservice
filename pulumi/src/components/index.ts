// pulumi/components/index.ts

import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { Vpc } from "./vpc";
import { InternetGateway } from "./internet-gateway";
import { RouteTable } from "./route-table";
import { Subnet } from "./subnet";
import { NatGateway } from "./nat-gateway";
import { CidrCalculator } from "@components/utils/cidr-calculator";
import { AvailabilityZoneUtils } from "@components/utils/availability-zones";
import { NetworkArgs } from "@types";

// Export all component types and classes
export * from "./types";
export * from "./vpc";
export * from "./internet-gateway";
export * from "./route-table";
export * from "./subnet";
export * from "./nat-gateway";
export * from "./utils/cidr-calculator";
export * from "./utils/availability-zones";

export class Network extends pulumi.ComponentResource {
  public readonly vpc: aws.ec2.Vpc;
  public readonly publicSubnets: aws.ec2.Subnet[] = [];
  public readonly privateSubnets: aws.ec2.Subnet[] = [];
  public readonly publicSubnetIds: pulumi.Output<string>[] = [];
  public readonly privateSubnetIds: pulumi.Output<string>[] = [];
  public readonly internetGateway?: aws.ec2.InternetGateway;
  public readonly publicRouteTable?: aws.ec2.RouteTable;
  public readonly privateRouteTables: aws.ec2.RouteTable[] = [];
  public readonly natGateways: aws.ec2.NatGateway[] = [];

  constructor(
    name: string,
    args: NetworkArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:Network", name, {}, opts);

    // Get availability zones
    const azSuffixes = AvailabilityZoneUtils.getAvailabilityZoneSuffixes(
      args.availabilityZones
    );
    const azCount = azSuffixes.length;

    // Create the VPC
    const vpcComponent = new Vpc(
      `${name}-vpc`,
      {
        cidrBlock: args.vpcCidrBlock,
        tags: args.tags,
      },
      { parent: this }
    );
    this.vpc = vpcComponent.vpc;

    // Create Internet Gateway if public subnets are enabled
    if (args.createPublicSubnets !== false) {
      const igwComponent = new InternetGateway(
        `${name}-igw`,
        {
          vpcId: this.vpc.id,
          tags: args.tags,
        },
        { parent: this }
      );
      this.internetGateway = igwComponent.gateway;

      // Create public route table
      const publicRouteTableComponent = new RouteTable(
        `${name}-public-rt`,
        {
          vpcId: this.vpc.id,
          tags: args.tags,
        },
        { parent: this }
      );
      this.publicRouteTable = publicRouteTableComponent.routeTable;

      // Add route to internet - fixed null check
      if (this.publicRouteTable && this.internetGateway) {
        publicRouteTableComponent.addRoute(`${name}-public-route`, {
          routeTableId: this.publicRouteTable.id,
          destinationCidrBlock: "0.0.0.0/0",
          gatewayId: this.internetGateway.id,
        });
      }
    }

    // Create private route tables for each AZ if private subnets are enabled
    if (args.createPrivateSubnets !== false) {
      for (let i = 0; i < azCount; i++) {
        const privateRouteTableComponent = new RouteTable(
          `${name}-private-rt-${i}`,
          {
            vpcId: this.vpc.id,
            tags: { ...args.tags, Name: `${name}-private-rt-${i}` },
          },
          { parent: this }
        );
        this.privateRouteTables.push(privateRouteTableComponent.routeTable);
      }
    }

    // Create subnets
    for (let i = 0; i < azCount; i++) {
      const az = azSuffixes[i];
      const azName = `${aws.config.region}${az}`;

      // Create public subnets if enabled
      if (args.createPublicSubnets !== false) {
        const publicCidr = CidrCalculator.calculateSubnetCidr(
          args.vpcCidrBlock,
          i,
          azCount
        );

        const publicSubnetComponent = new Subnet(
          `${name}-public-${i}`,
          {
            vpcId: this.vpc.id,
            cidrBlock: publicCidr,
            availabilityZone: azName,
            mapPublicIpOnLaunch: true,
            type: "public",
            tags: args.tags,
          },
          { parent: this }
        );

        const publicSubnet = publicSubnetComponent.subnet;
        this.publicSubnets.push(publicSubnet);
        this.publicSubnetIds.push(publicSubnet.id);

        // Associate with public route table
        if (this.publicRouteTable) {
          new aws.ec2.RouteTableAssociation(
            `${name}-public-rta-${i}`,
            {
              subnetId: publicSubnet.id,
              routeTableId: this.publicRouteTable.id,
            },
            { parent: this }
          );
        }

        // Create NAT Gateway in public subnet if private subnets are enabled
        if (args.createPrivateSubnets !== false) {
          const natGatewayComponent = new NatGateway(
            `${name}-nat-${i}`,
            {
              subnetId: publicSubnet.id,
              createElasticIp: true,
              tags: args.tags,
            } as any,
            { parent: this }
          ); // Using 'any' due to parameter incompatibility

          const natGateway = natGatewayComponent.natGateway;
          this.natGateways.push(natGateway);
        }
      }

      // Create private subnets if enabled
      if (args.createPrivateSubnets !== false) {
        const privateCidr = CidrCalculator.calculateSubnetCidr(
          args.vpcCidrBlock,
          i + azCount,
          azCount
        );

        const privateSubnetComponent = new Subnet(
          `${name}-private-${i}`,
          {
            vpcId: this.vpc.id,
            cidrBlock: privateCidr,
            availabilityZone: azName,
            type: "private",
            tags: args.tags,
          },
          { parent: this }
        );

        const privateSubnet = privateSubnetComponent.subnet;
        this.privateSubnets.push(privateSubnet);
        this.privateSubnetIds.push(privateSubnet.id);

        // Associate with private route table
        if (this.privateRouteTables[i]) {
          new aws.ec2.RouteTableAssociation(
            `${name}-private-rta-${i}`,
            {
              subnetId: privateSubnet.id,
              routeTableId: this.privateRouteTables[i].id,
            },
            { parent: this }
          );

          // Add route to NAT gateway if public subnets are enabled
          if (args.createPublicSubnets !== false && this.natGateways[i]) {
            new aws.ec2.Route(
              `${name}-private-route-${i}`,
              {
                routeTableId: this.privateRouteTables[i].id,
                destinationCidrBlock: "0.0.0.0/0",
                natGatewayId: this.natGateways[i].id,
              },
              { parent: this }
            );
          }
        }
      }
    }

    this.registerOutputs({
      vpc: this.vpc,
      publicSubnets: this.publicSubnets,
      privateSubnets: this.privateSubnets,
      publicSubnetIds: this.publicSubnetIds,
      privateSubnetIds: this.privateSubnetIds,
      internetGateway: this.internetGateway,
      publicRouteTable: this.publicRouteTable,
      privateRouteTables: this.privateRouteTables,
      natGateways: this.natGateways,
    });
  }
}
