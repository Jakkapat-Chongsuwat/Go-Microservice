import * as pulumi from "@pulumi/pulumi";
import { EcrRepository } from "../../components/ecr-repository";
import * as aws from "@pulumi/aws";

export interface EcrRepositoryArgs {
  /**
   * Name of the repository
   */
  repositoryName: string;

  /**
   * Maximum number of images to keep (for lifecycle policy)
   */
  maxImageCount?: number;

  /**
   * Tags to apply to the repository
   */
  tags?: Record<string, string>;
}

/**
 * EcrRepositoryResource manages an ECR repository for container images
 */
export class EcrRepositoryResource {
  /**
   * The ECR repository
   */
  public readonly repository: EcrRepository;

  /**
   * The repository URL
   */
  public readonly repositoryUrl: pulumi.Output<string>;

  /**
   * Creates a new ECR repository
   *
   * @param name Base name for resources
   * @param args Repository configuration
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: EcrRepositoryArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    const lifecyclePolicy = args.maxImageCount
      ? { maxImageCount: args.maxImageCount }
      : undefined;

    this.repository = new EcrRepository(
      name,
      {
        repositoryName: args.repositoryName,
        lifecyclePolicy: lifecyclePolicy,
        tags: args.tags,
      },
      opts
    );

    this.repositoryUrl = this.repository.repositoryUrl;
  }

  /**
   * Get the ECR login password
   *
   * @returns A pulumi.Output with the ECR authorization token
   */
  public getAuthToken(): pulumi.Output<string> {
    return pulumi.output(aws.ecr.getAuthorizationToken()).password;
  }
}
