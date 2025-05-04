import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { VpcArgs } from "./types";

export class Vpc extends pulumi.ComponentResource {
    public readonly vpc: aws.ec2.Vpc;

    constructor(name: string, args: VpcArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:Vpc", name, {}, opts);

        this.vpc = new aws.ec2.Vpc(name, {
            cidrBlock: args.cidrBlock,
            enableDnsHostnames: args.enableDnsHostnames ?? true,
            enableDnsSupport: args.enableDnsSupport ?? true,
            tags: { ...args.tags, Name: name },
        }, { parent: this });

        this.registerOutputs({
            vpc: this.vpc,
        });
    }
}
