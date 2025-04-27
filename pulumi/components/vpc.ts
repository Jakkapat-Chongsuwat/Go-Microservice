// vpc.ts
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface VpcArgs {
    cidrBlock: string;
    availabilityZones?: string[];
    tags?: { [key: string]: string };
}

export class Vpc extends pulumi.ComponentResource {
    public readonly vpc: aws.ec2.Vpc;
    public readonly publicSubnets: aws.ec2.Subnet[] = [];
    public readonly privateSubnets: aws.ec2.Subnet[] = [];
    public readonly publicSubnetIds: pulumi.Output<string>[] = [];
    public readonly privateSubnetIds: pulumi.Output<string>[] = [];

    constructor(name: string, args: VpcArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:Vpc", name, {}, opts);

        // Get the availability zones in the current region
        const availabilityZones = args.availabilityZones || ["a", "b", "c"];
        
        // Create the VPC
        this.vpc = new aws.ec2.Vpc(name, {
            cidrBlock: args.cidrBlock,
            enableDnsHostnames: true,
            enableDnsSupport: true,
            tags: { ...args.tags, Name: name },
        }, { parent: this });

        // Create an internet gateway for the VPC
        const igw = new aws.ec2.InternetGateway(`${name}-igw`, {
            vpcId: this.vpc.id,
            tags: { ...args.tags, Name: `${name}-igw` },
        }, { parent: this });

        // Create a public route table
        const publicRouteTable = new aws.ec2.RouteTable(`${name}-public-rt`, {
            vpcId: this.vpc.id,
            tags: { ...args.tags, Name: `${name}-public-rt` },
        }, { parent: this });

        // Add a route to the internet
        new aws.ec2.Route(`${name}-public-route`, {
            routeTableId: publicRouteTable.id,
            destinationCidrBlock: "0.0.0.0/0",
            gatewayId: igw.id,
        }, { parent: this });

        // Create NAT gateways and private route tables for each AZ
        const eips: aws.ec2.Eip[] = [];
        const natGateways: aws.ec2.NatGateway[] = [];
        const privateRouteTables: aws.ec2.RouteTable[] = [];

        for (let i = 0; i < availabilityZones.length; i++) {
            const az = availabilityZones[i];
            
            // Create an Elastic IP for the NAT gateway
            const eip = new aws.ec2.Eip(`${name}-eip-${i}`, {
                vpc: true,
                tags: { ...args.tags, Name: `${name}-eip-${i}` },
            }, { parent: this });
            eips.push(eip);
            
            // Create a private route table for this AZ
            const privateRouteTable = new aws.ec2.RouteTable(`${name}-private-rt-${i}`, {
                vpcId: this.vpc.id,
                tags: { ...args.tags, Name: `${name}-private-rt-${i}` },
            }, { parent: this });
            privateRouteTables.push(privateRouteTable);
        }

        // Create public and private subnets across the AZs
        for (let i = 0; i < availabilityZones.length; i++) {
            const az = availabilityZones[i];
            
            // Calculate subnet CIDR blocks
            const publicCidr = this.calculateSubnetCidr(args.cidrBlock, i, availabilityZones.length);
            const privateCidr = this.calculateSubnetCidr(args.cidrBlock, i + availabilityZones.length, availabilityZones.length);
            
            // Create a public subnet in this AZ
            const publicSubnet = new aws.ec2.Subnet(`${name}-public-${i}`, {
                vpcId: this.vpc.id,
                cidrBlock: publicCidr,
                availabilityZone: `${aws.config.region}${az}`,
                mapPublicIpOnLaunch: true,
                tags: {
                    ...args.tags,
                    Name: `${name}-public-${i}`,
                    "kubernetes.io/role/elb": "1",
                },
            }, { parent: this });
            this.publicSubnets.push(publicSubnet);
            this.publicSubnetIds.push(publicSubnet.id);

            // Associate the public subnet with the public route table
            new aws.ec2.RouteTableAssociation(`${name}-public-rta-${i}`, {
                subnetId: publicSubnet.id,
                routeTableId: publicRouteTable.id,
            }, { parent: this });
            
            // Create a NAT Gateway in the public subnet
            const natGateway = new aws.ec2.NatGateway(`${name}-nat-${i}`, {
                allocationId: eips[i].id,
                subnetId: publicSubnet.id,
                tags: { ...args.tags, Name: `${name}-nat-${i}` },
            }, { parent: this });
            natGateways.push(natGateway);
            
            // Create a private subnet in this AZ
            const privateSubnet = new aws.ec2.Subnet(`${name}-private-${i}`, {
                vpcId: this.vpc.id,
                cidrBlock: privateCidr,
                availabilityZone: `${aws.config.region}${az}`,
                tags: { 
                    ...args.tags, 
                    Name: `${name}-private-${i}`,
                    "kubernetes.io/role/internal-elb": "1",
                },
            }, { parent: this });
            this.privateSubnets.push(privateSubnet);
            this.privateSubnetIds.push(privateSubnet.id);

            // Associate the private subnet with its route table
            new aws.ec2.RouteTableAssociation(`${name}-private-rta-${i}`, {
                subnetId: privateSubnet.id,
                routeTableId: privateRouteTables[i].id,
            }, { parent: this });
            
            // Add a route to the NAT gateway from the private subnet
            new aws.ec2.Route(`${name}-private-route-${i}`, {
                routeTableId: privateRouteTables[i].id,
                destinationCidrBlock: "0.0.0.0/0",
                natGatewayId: natGateways[i].id,
            }, { parent: this });
        }

        this.registerOutputs({
            vpc: this.vpc,
            publicSubnets: this.publicSubnets,
            privateSubnets: this.privateSubnets,
            publicSubnetIds: this.publicSubnetIds,
            privateSubnetIds: this.privateSubnetIds,
        });
    }
    
    // Helper function to calculate subnet CIDR blocks
    private calculateSubnetCidr(vpcCidr: string, subnetIndex: number, totalSubnets: number): string {
        const [baseIp, baseCidr] = vpcCidr.split('/');
        
        const baseCidrNum = parseInt(baseCidr, 10);
        const newCidrNum = baseCidrNum + Math.ceil(Math.log2(totalSubnets * 2));
        
        // Calculate the subnet size
        const subnetSize = 2 ** (32 - newCidrNum);
        
        // Split the base IP into octets
        const octets = baseIp.split('.').map(octet => parseInt(octet, 10));
        
        // Convert the base IP to a number
        let ipNum = (octets[0] << 24) + (octets[1] << 16) + (octets[2] << 8) + octets[3];
        
        // Calculate the new IP for this subnet
        ipNum += subnetSize * subnetIndex;
        
        // Convert back to IP address format
        const newOctets = [
            (ipNum >>> 24) & 0xFF,
            (ipNum >>> 16) & 0xFF,
            (ipNum >>> 8) & 0xFF,
            ipNum & 0xFF
        ];
        
        return `${newOctets.join('.')}/${newCidrNum}`;
    }
}
