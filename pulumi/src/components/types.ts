import * as pulumi from "@pulumi/pulumi";

/** Base args for anything that takes tags */
export interface BaseResourceArgs {
  /** Accepts either plain `{[k:string]:string}` or a Pulumi Output thereof */
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
}

/** VPC settings */
export interface VpcArgs extends BaseResourceArgs {
  cidrBlock: string;
  enableDnsHostnames?: boolean;
  enableDnsSupport?: boolean;
}

/** Subnet settings */
export interface SubnetArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
  cidrBlock: string;
  availabilityZone: pulumi.Input<string>;
  mapPublicIpOnLaunch?: boolean;
  type: "public" | "private";
}

/** Internet Gateway settings */
export interface InternetGatewayArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
}

/** Route Table settings */
export interface RouteTableArgs extends BaseResourceArgs {
  vpcId: pulumi.Input<string>;
}

/** Single-route settings */
export interface RouteArgs {
  routeTableId: pulumi.Input<string>;
  destinationCidrBlock: string;
  gatewayId?: pulumi.Input<string>;
  natGatewayId?: pulumi.Input<string>;
}

/** RouteTableAssociation settings */
export interface RouteTableAssociationArgs {
  routeTableId: pulumi.Input<string>;
  subnetId: pulumi.Input<string>;
}

/** Elastic IP settings */
export interface EipArgs extends BaseResourceArgs {
  vpc: boolean;
}

/** “Umbrella” Network component settings — plain JS for config, but tags are Inputs */
export interface NetworkArgs {
  vpcCidrBlock: string;
  availabilityZones?: string[];
  createPrivateSubnets?: boolean;
  createPublicSubnets?: boolean;
  tags?: pulumi.Input<Record<string, pulumi.Input<string>>>;
}

export type ChartValues = {
  [key: string]: string | number | boolean | ChartValues | ChartValues[];
};
