import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

/**
 * Arguments for creating an ECR repository.
 *
 * An ECR repository is like a specialized warehouse for your container images -
 * a secure, organized place to store the blueprints for your application containers.
 */
export interface EcrRepositoryArgs {
  /**
   * The name of the ECR repository
   *
   * Like naming a warehouse so your delivery trucks (CI/CD pipelines) know
   * exactly where to store and retrieve container images.
   */
  repositoryName: pulumi.Input<string>; // Changed to Input<string>

  /**
   * Enable image scanning on push
   *
   * Similar to how airport security scans luggage, this feature examines your
   * container images for known vulnerabilities when they arrive at the repository.
   *
   * @default true
   */
  scanOnPush?: boolean;

  /**
   * Image tag mutability
   *
   * Controls whether image tags can be overwritten:
   * - MUTABLE: Like allowing new editions of books to replace old ones on the same shelf
   * - IMMUTABLE: Like requiring each new edition to get its own dedicated shelf space
   *
   * For EKS workloads, immutable tags provide stronger guarantees about which
   * image version is actually running.
   *
   * @default "MUTABLE"
   */
  imageTagMutability?: "MUTABLE" | "IMMUTABLE";

  /**
   * Lifecycle policy to manage image retention
   *
   * Like automated inventory management for your container warehouse - decides
   * which container images to keep and which to discard based on specific rules.
   *
   * @default undefined
   */
  lifecyclePolicy?: {
    /**
     * Maximum number of images to keep
     *
     * Like limiting a library's collection to a specific number of books,
     * keeping only the most recent editions.
     */
    maxImageCount?: number;

    /**
     * Maximum age of images to keep (in days)
     *
     * Similar to food products having an expiration date, this removes images
     * that are older than a certain age to prevent clutter and reduce costs.
     */
    maxImageAge?: number;
  };

  /**
   * Tags to apply to the repository
   *
   * Like attaching classification labels to organize and manage your warehouses
   * in a large logistics network.
   *
   * @default {}
   */
  tags?: { [key: string]: string };
}

/**
 * EcrRepository creates and manages an Elastic Container Registry repository.
 *
 * In the EKS ecosystem, ECR serves as the secure storage location for your container images,
 * much like how a well-organized warehouse serves a manufacturing operation - storing,
 * cataloging, and distributing the components needed to run your applications.
 */
export class EcrRepository extends pulumi.ComponentResource {
  /**
   * The underlying ECR repository
   *
   * This is the actual AWS resource that stores your container images,
   * like the physical warehouse building itself.
   */
  public readonly repository: aws.ecr.Repository;

  /**
   * The URL of the ECR repository
   *
   * This is like the address of your warehouse - Kubernetes and CI/CD systems
   * use this URL to locate where to push or pull container images.
   */
  public readonly repositoryUrl: pulumi.Output<string>;

  /**
   * The ARN of the ECR repository
   *
   * The Amazon Resource Name is like a globally unique identifier or deed
   * for the repository, used for granting permissions and in cross-account scenarios.
   */
  public readonly repositoryArn: pulumi.Output<string>;

  constructor(
    name: string,
    args: EcrRepositoryArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:EcrRepository", name, {}, opts);

    // Create the ECR repository
    // This is like filing the paperwork to establish a new warehouse facility
    this.repository = new aws.ecr.Repository(
      name,
      {
        name: args.repositoryName, // The official name registered with AWS

        // Configure security scanning, similar to setting up inspection procedures
        imageScanningConfiguration: {
          scanOnPush: args.scanOnPush !== undefined ? args.scanOnPush : true,
        },

        // Set tag mutability, determining if existing images can be replaced
        imageTagMutability: args.imageTagMutability || "MUTABLE",

        // Apply identification tags, like property markers
        tags: {
          ...(args.tags || {}),
          ManagedBy: "pulumi", // Indicates this warehouse is managed by Pulumi's systems
        },
      },
      { parent: this }
    );

    // Add lifecycle policy if specified
    // This establishes automated inventory management rules
    if (args.lifecyclePolicy) {
      const rules: any[] = [];

      // Rule to keep only the newest X images
      if (args.lifecyclePolicy.maxImageCount) {
        rules.push({
          rulePriority: 1, // Higher priority rules execute first, like emergency protocols
          description: `Keep only the last ${args.lifecyclePolicy.maxImageCount} images`,
          selection: {
            tagStatus: "any", // Applies to all images, both tagged and untagged
            countType: "imageCountMoreThan", // The counting mechanism - like saying "if shelf contains more than X books"
            countNumber: args.lifecyclePolicy.maxImageCount, // The threshold number that triggers the rule
          },
          action: {
            type: "expire", // The action to take - like discarding inventory
          },
        });
      }

      // Rule to expire images older than X days
      if (args.lifecyclePolicy.maxImageAge) {
        rules.push({
          rulePriority: rules.length + 1, // Like the next item in a sequence of instructions
          description: `Expire images older than ${args.lifecyclePolicy.maxImageAge} days`,
          selection: {
            tagStatus: "any", // Applies to all images regardless of their labeling
            countType: "sinceImagePushed", // Time-based selection, like checking dates on products
            countUnit: "days", // The unit of time measurement
            countNumber: args.lifecyclePolicy.maxImageAge, // How many days before expiration
          },
          action: {
            type: "expire", // Remove the image, like clearing old inventory
          },
        });
      }

      // Only create a lifecycle policy if we have rules
      // No sense in hanging an empty rule book on the wall
      if (rules.length > 0) {
        new aws.ecr.LifecyclePolicy(
          `${name}-lifecycle-policy`,
          {
            repository: this.repository.name, // Link policy to our specific repository
            policy: JSON.stringify({ rules }), // Convert our rules to the format AWS expects
          },
          { parent: this }
        );
      }
    }

    // Expose key properties for external use
    // Like publishing the warehouse address and credentials in a directory
    this.repositoryUrl = this.repository.repositoryUrl;
    this.repositoryArn = this.repository.arn;

    // Register outputs for Pulumi state management
    // Like filing the final property records
    this.registerOutputs({
      repositoryUrl: this.repositoryUrl,
      repositoryArn: this.repositoryArn,
    });
  }
}
