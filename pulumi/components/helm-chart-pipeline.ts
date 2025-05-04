import * as pulumi from "@pulumi/pulumi";
import * as command from "@pulumi/command";
import * as aws from "@pulumi/aws";
import { EcrRepository } from "./ecr-repository";
import { HelmChart } from "./helm-chart";

export interface HelmChartPipelineArgs {
    /**
     * Name of the ECR repository to create
     */
    ecrRepositoryName: pulumi.Input<string>;
    
    /**
     * AWS region for ECR
     */
    region: string;
    
    /**
     * Local path to the Helm chart directory
     */
    helmChartPath: string;
    
    /**
     * Chart name to use
     */
    chartName: string;
    
    /**
     * Chart version to use
     */
    chartVersion: string;
    
    /**
     * Deployment target namespace
     */
    namespace: pulumi.Input<string>;
    
    /**
     * Release name to use for the chart
     */
    releaseName: string;
    
    /**
     * Values to override in the chart
     */
    values?: { [key: string]: any };
    
    /**
     * Whether to deploy the chart after pushing
     * @default true
     */
    deploy?: boolean;
    
    /**
     * Tags to apply to ECR repository
     */
    tags?: { [key: string]: string };
}

export class HelmChartPipeline extends pulumi.ComponentResource {
    /**
     * The ECR repository where charts are stored
     */
    public readonly ecrRepository: EcrRepository;
    
    /**
     * The Helm chart deployment (if deployed)
     */
    public readonly helmDeployment?: HelmChart;
    
    /**
     * The ECR repository URL
     */
    public readonly repositoryUrl: pulumi.Output<string>;
    
    constructor(name: string, args: HelmChartPipelineArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:HelmChartPipeline", name, {}, opts);
        
        // Create ECR repository
        this.ecrRepository = new EcrRepository(name, {
            // Use the original input type instead of trying to convert it
            repositoryName: args.ecrRepositoryName,
            imageTagMutability: "IMMUTABLE", // Good practice for Helm charts
            lifecyclePolicy: {
                maxImageCount: 10,
            },
            tags: {
                ...args.tags || {},
                Service: args.chartName,
            },
        }, { parent: this });
        
        this.repositoryUrl = this.ecrRepository.repositoryUrl;
        
        // Get AWS account ID
        const callerIdentity = aws.getCallerIdentity();
        
        // Package and push Helm chart to ECR
        const packageAndPushChart = new command.local.Command(`${name}-package-and-push`, {
            create: pulumi.interpolate`
                cd "${args.helmChartPath}" && \
                helm package . --version ${args.chartVersion} && \
                aws ecr get-login-password --region ${args.region} | \
                helm registry login --username AWS --password-stdin ${callerIdentity.then(id => `${id.accountId}.dkr.ecr.${args.region}.amazonaws.com`)} && \
                helm push ${args.chartName}-${args.chartVersion}.tgz \
                oci://${callerIdentity.then(id => `${id.accountId}.dkr.ecr.${args.region}.amazonaws.com/${args.ecrRepositoryName}`)}
            `,
            environment: {
                AWS_REGION: args.region,
            },
        }, { parent: this, dependsOn: [this.ecrRepository] });
        
        // Deploy chart if requested
        if (args.deploy !== false) {
            const chartUrl = pulumi.interpolate`oci://${callerIdentity.then(id => `${id.accountId}.dkr.ecr.${args.region}.amazonaws.com/${args.ecrRepositoryName}/${args.chartName}`)}`;
            
            this.helmDeployment = new HelmChart(`${name}-deployment`, {
                namespace: args.namespace,
                createNamespace: true,
                chart: chartUrl,
                version: args.chartVersion,
                values: args.values || {},
                repositoryOpts: {
                    repository: pulumi.interpolate`${callerIdentity.then(id => `${id.accountId}.dkr.ecr.${args.region}.amazonaws.com/${args.ecrRepositoryName}`)}`,
                },
            }, { parent: this, dependsOn: [packageAndPushChart] });
        }
        
        this.registerOutputs({
            ecrRepository: this.ecrRepository,
            helmDeployment: this.helmDeployment,
            repositoryUrl: this.repositoryUrl,
        });
    }
}
