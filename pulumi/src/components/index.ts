import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

import { Vpc } from "./vpc";
import { InternetGateway } from "./internet-gateway";
import { RouteTable } from "./route-table";
import { Subnet } from "./subnet";
import { NatGateway } from "./nat-gateway";
import { CidrCalculator } from "./utils/cidr-calculator";
import { AvailabilityZoneUtils } from "./utils/availability-zones";

import { NetworkArgs } from "./types";

// Re-export everything:
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

    // availabilityZones is still a plain string[]
    const azs = AvailabilityZoneUtils.getAvailabilityZoneSuffixes(
      args.availabilityZones ?? []
    );
    const azCount = azs.length;

    // Create the VPC
    const vpcComp = new Vpc(
      `${name}-vpc`,
      { cidrBlock: args.vpcCidrBlock, tags: args.tags },
      { parent: this }
    );
    this.vpc = vpcComp.vpc;

    // Internet Gateway + public RT
    if (args.createPublicSubnets ?? true) {
      const igwComp = new InternetGateway(
        `${name}-igw`,
        { vpcId: this.vpc.id, tags: args.tags },
        { parent: this }
      );
      this.internetGateway = igwComp.gateway;

      const prtComp = new RouteTable(
        `${name}-public-rt`,
        { vpcId: this.vpc.id, tags: args.tags },
        { parent: this }
      );
      this.publicRouteTable = prtComp.routeTable;

      if (this.internetGateway && this.publicRouteTable) {
        prtComp.addRoute(`${name}-public-route`, {
          routeTableId: this.publicRouteTable.id,
          destinationCidrBlock: "0.0.0.0/0",
          gatewayId: this.internetGateway.id,
        });
      }
    }

    // Private RTs
    if (args.createPrivateSubnets ?? true) {
      for (let i = 0; i < azCount; i++) {
        const prt = new RouteTable(
          `${name}-private-rt-${i}`,
          {
            vpcId: this.vpc.id,
            tags: { ...(args.tags ?? {}), Name: `${name}-private-rt-${i}` },
          },
          { parent: this }
        );
        this.privateRouteTables.push(prt.routeTable);
      }
    }

    // Subnets & NAT Gateways
    for (let i = 0; i < azCount; i++) {
      const suffix = azs[i];
      const azName = `${aws.config.region}${suffix}`;

      // Public subnet
      if (args.createPublicSubnets ?? true) {
        const pubCidr = CidrCalculator.calculateSubnetCidr(
          args.vpcCidrBlock,
          i,
          azCount
        );
        const pub = new Subnet(
          `${name}-public-${i}`,
          {
            vpcId: this.vpc.id,
            cidrBlock: pubCidr,
            availabilityZone: azName,
            mapPublicIpOnLaunch: true,
            type: "public",
            tags: args.tags,
          },
          { parent: this }
        );
        this.publicSubnets.push(pub.subnet);
        this.publicSubnetIds.push(pub.subnet.id);

        if (this.publicRouteTable) {
          new aws.ec2.RouteTableAssociation(
            `${name}-public-rta-${i}`,
            {
              subnetId: pub.subnet.id,
              routeTableId: this.publicRouteTable.id,
            },
            { parent: this }
          );
        }

        // NAT for private subnets
        if (args.createPrivateSubnets ?? true) {
          const ng = new NatGateway(
            `${name}-nat-${i}`,
            {
              subnetId: pub.subnet.id,
              createElasticIp: true,
              tags: args.tags,
            },
            { parent: this }
          );
          this.natGateways.push(ng.natGateway);
        }
      }

      // Private subnet
      if (args.createPrivateSubnets ?? true) {
        const privCidr = CidrCalculator.calculateSubnetCidr(
          args.vpcCidrBlock,
          i + azCount,
          azCount
        );
        const priv = new Subnet(
          `${name}-private-${i}`,
          {
            vpcId: this.vpc.id,
            cidrBlock: privCidr,
            availabilityZone: azName,
            type: "private",
            tags: args.tags,
          },
          { parent: this }
        );
        this.privateSubnets.push(priv.subnet);
        this.privateSubnetIds.push(priv.subnet.id);

        const prt = this.privateRouteTables[i];
        if (prt) {
          new aws.ec2.RouteTableAssociation(
            `${name}-private-rta-${i}`,
            {
              subnetId: priv.subnet.id,
              routeTableId: prt.id,
            },
            { parent: this }
          );
          if (args.createPublicSubnets ?? true) {
            new aws.ec2.Route(
              `${name}-private-route-${i}`,
              {
                routeTableId: prt.id,
                destinationCidrBlock: "0.0.0.0/0",
                natGatewayId: this.natGateways[i]?.id,
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
