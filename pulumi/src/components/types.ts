import * as pulumi from "@pulumi/pulumi";

export interface BaseResourceArgs {
  tags?: { [key: string]: string };
}

export interface VpcArgs extends BaseResourceArgs {
  cidrBlock: string;
  enableDnsHostnames?: boolean;
  enableDnsSupport?: boolean;
}

export interface SubnetArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
  cidrBlock: string;
  availabilityZone: string;
  mapPublicIpOnLaunch?: boolean;
  type: "public" | "private";
}

export interface InternetGatewayArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
}

export interface RouteTableArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
}

export interface RouteArgs {
  routeTableId: pulumi.Input<string>;
  destinationCidrBlock: string;
  gatewayId?: pulumi.Input<string>;
  natGatewayId?: pulumi.Input<string>;
}

export interface RouteTableAssociationArgs {
  routeTableId: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
}

export interface NatGatewayArgs extends BaseResourceArgs {
  allocationId: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
}

export interface EipArgs extends BaseResourceArgs {
  vpc: boolean;
}

export interface NetworkArgs extends BaseResourceArgs {
  vpcCidrBlock: string;
  availabilityZones?: string[];
  createPrivateSubnets?: boolean;
  createPublicSubnets?: boolean;
}

export type ChartValues = {
  [key: string]: string | number | boolean | ChartValues | ChartValues[];
};
