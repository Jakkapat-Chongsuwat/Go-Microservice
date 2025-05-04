import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { InternetGatewayArgs } from "./types";

export class InternetGateway extends pulumi.ComponentResource {
    public readonly gateway: aws.ec2.InternetGateway;

    constructor(name: string, args: InternetGatewayArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:InternetGateway", name, {}, opts);

        this.gateway = new aws.ec2.InternetGateway(name, {
            vpcId: args.vpcId,
            tags: { ...args.tags, Name: name },
        }, { parent: this });

        this.registerOutputs({
            gateway: this.gateway,
        });
    }
}
