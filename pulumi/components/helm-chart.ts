import * as pulumi from "@pulumi/pulumi";
import * as kubernetes from "@pulumi/kubernetes";
import * as fs from "fs";

export interface HelmChartArgs {
    /**
     * The namespace to deploy the chart to
     */
    namespace: pulumi.Input<string>;
    
    /**
     * Create namespace if it doesn't exist
     * @default true
     */
    createNamespace?: boolean;
    
    /**
     * Chart location - can be a local path or OCI repository URL
     */
    chart: pulumi.Input<string>;
    
    /**
     * Chart version to deploy
     */
    version?: pulumi.Input<string>;
    
    /**
     * Values to override in the chart
     */
    values?: { [key: string]: any };
    
    /**
     * Optional chart value files to include
     */
    valueFiles?: string[];
    
    /**
     * Set atomic deployment (rollback on failure)
     * @default true
     */
    atomic?: boolean;
    
    /**
     * Clean up on failure
     * @default true
     */
    cleanup?: boolean;
    
    /**
     * Maximum time to wait for resources (in seconds)
     * @default 300
     */
    timeout?: number;
    
    /**
     * Skip CRD creation/deletion
     * @default false
     */
    skipCrds?: boolean;
    
    /**
     * OCI registry authentication (if using OCI chart)
     */
    repositoryOpts?: {
        /**
         * Repository URL
         */
        repository: pulumi.Input<string>;
        
        /**
         * Username for repository
         */
        username?: pulumi.Input<string>;
        
        /**
         * Password for repository
         */
        password?: pulumi.Input<string>;
    };
}

export class HelmChart extends pulumi.ComponentResource {
    /**
     * The underlying Helm chart deployment
     */
    public readonly chart: kubernetes.helm.v3.Chart;
    
    /**
     * The namespace where the chart is deployed
     */
    public readonly namespace: kubernetes.core.v1.Namespace | undefined;
    
    constructor(name: string, args: HelmChartArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:HelmChart", name, {}, opts);
        
        if (args.createNamespace !== false) {
            this.namespace = new kubernetes.core.v1.Namespace(name, {
                metadata: {
                    name: args.namespace,
                },
            }, { parent: this });
        }
        
        const fetchOpts = args.repositoryOpts ? {
            repo: args.repositoryOpts.repository,
            username: args.repositoryOpts.username,
            password: args.repositoryOpts.password,
        } : undefined;
        
        let combinedValues = args.values || {};
        if (args.valueFiles && args.valueFiles.length > 0) {
            for (const filePath of args.valueFiles) {
                try {
                    if (fs.existsSync(filePath)) {
                        const valueFileContent = fs.readFileSync(filePath, 'utf8');
                        const valueFileObj = require('js-yaml').load(valueFileContent);
                        combinedValues = deepMerge(combinedValues, valueFileObj);
                    }
                } catch (error) {
                    console.warn(`Failed to load values file ${filePath}: ${error}`);
                }
            }
        }
        
        this.chart = new kubernetes.helm.v3.Chart(name, {
            chart: args.chart,
            version: args.version,
            namespace: args.namespace,
            values: combinedValues,
            skipAwait: false,
            fetchOpts: fetchOpts,
            transformations: [
                (obj: any, opts: pulumi.CustomResourceOptions) => {
                }
            ],
        }, { parent: this, dependsOn: this.namespace ? [this.namespace] : undefined });
        
        this.registerOutputs({
            chart: this.chart,
            namespace: this.namespace,
        });
    }
}

function deepMerge(target: any, source: any): any {
    const output = { ...target };
    
    if (isObject(target) && isObject(source)) {
        Object.keys(source).forEach(key => {
            if (isObject(source[key])) {
                if (!(key in target)) {
                    Object.assign(output, { [key]: source[key] });
                } else {
                    output[key] = deepMerge(target[key], source[key]);
                }
            } else {
                Object.assign(output, { [key]: source[key] });
            }
        });
    }
    
    return output;
}

function isObject(item: any): boolean {
    return (item && typeof item === 'object' && !Array.isArray(item));
}
