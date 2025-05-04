import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface EcrRepositoryArgs {
    /**
     * The name of the ECR repository
     */
    repositoryName: pulumi.Input<string>; // Changed to Input<string>
    
    /**
     * Enable image scanning on push
     * @default true
     */
    scanOnPush?: boolean;
    
    /**
     * Image tag mutability 
     * @default "MUTABLE"
     */
    imageTagMutability?: "MUTABLE" | "IMMUTABLE";
    
    /**
     * Lifecycle policy to manage image retention
     * @default undefined
     */
    lifecyclePolicy?: {
        /** Maximum number of images to keep */
        maxImageCount?: number;
        /** Maximum age of images to keep (in days) */
        maxImageAge?: number;
    };
    
    /**
     * Tags to apply to the repository
     * @default {}
     */
    tags?: { [key: string]: string };
}

export class EcrRepository extends pulumi.ComponentResource {
    /**
     * The underlying ECR repository
     */
    public readonly repository: aws.ecr.Repository;
    
    /**
     * The URL of the ECR repository
     */
    public readonly repositoryUrl: pulumi.Output<string>;
    
    /**
     * The ARN of the ECR repository
     */
    public readonly repositoryArn: pulumi.Output<string>;
    
    constructor(name: string, args: EcrRepositoryArgs, opts?: pulumi.ComponentResourceOptions) {
        super("custom:resource:EcrRepository", name, {}, opts);
        
        // Create the ECR repository
        this.repository = new aws.ecr.Repository(name, {
            name: args.repositoryName, // AWS provider accepts Input<string>
            imageScanningConfiguration: {
                scanOnPush: args.scanOnPush !== undefined ? args.scanOnPush : true,
            },
            imageTagMutability: args.imageTagMutability || "MUTABLE",
            tags: {
                ...args.tags || {},
                ManagedBy: "pulumi",
            },
        }, { parent: this });
        
        // Add lifecycle policy if specified
        if (args.lifecyclePolicy) {
            const rules: any[] = [];
            
            if (args.lifecyclePolicy.maxImageCount) {
                rules.push({
                    rulePriority: 1,
                    description: `Keep only the last ${args.lifecyclePolicy.maxImageCount} images`,
                    selection: {
                        tagStatus: "any",
                        countType: "imageCountMoreThan",
                        countNumber: args.lifecyclePolicy.maxImageCount,
                    },
                    action: {
                        type: "expire",
                    },
                });
            }
            
            if (args.lifecyclePolicy.maxImageAge) {
                rules.push({
                    rulePriority: rules.length + 1,
                    description: `Expire images older than ${args.lifecyclePolicy.maxImageAge} days`,
                    selection: {
                        tagStatus: "any",
                        countType: "sinceImagePushed",
                        countUnit: "days",
                        countNumber: args.lifecyclePolicy.maxImageAge,
                    },
                    action: {
                        type: "expire",
                    },
                });
            }
            
            if (rules.length > 0) {
                new aws.ecr.LifecyclePolicy(`${name}-lifecycle-policy`, {
                    repository: this.repository.name,
                    policy: JSON.stringify({ rules }),
                }, { parent: this });
            }
        }
        
        this.repositoryUrl = this.repository.repositoryUrl;
        this.repositoryArn = this.repository.arn;
        
        this.registerOutputs({
            repositoryUrl: this.repositoryUrl,
            repositoryArn: this.repositoryArn,
        });
    }
}
