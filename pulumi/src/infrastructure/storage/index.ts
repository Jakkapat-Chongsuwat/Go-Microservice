import * as pulumi from "@pulumi/pulumi";
import { EcrRepositoryResource } from "./ecr";
import { createComponentName } from "../../utils/naming";

export interface StorageStackArgs {
  /**
   * Project name used for repository naming
   */
  projectName: string;

  /**
   * Tags to apply to resources
   */
  tags?: Record<string, string>;
}

/**
 * StorageStack creates and manages storage resources
 */
export class StorageStack {
  /**
   * Application container repository
   */
  public readonly appRepository: EcrRepositoryResource;

  /**
   * Helm charts repository
   */
  public readonly helmRepository: EcrRepositoryResource;

  /**
   * Application repository URL
   */
  public readonly appRepositoryUrl: pulumi.Output<string>;

  /**
   * Helm repository URL
   */
  public readonly helmRepositoryUrl: pulumi.Output<string>;

  /**
   * Creates a new storage stack
   *
   * @param name Base name for resources
   * @param args Storage stack configuration
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: StorageStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    this.appRepository = new EcrRepositoryResource(
      createComponentName(name, "app-repo"),
      {
        repositoryName: `${args.projectName}-user-service`,
        maxImageCount: 10,
        tags: args.tags,
      },
      opts
    );

    this.helmRepository = new EcrRepositoryResource(
      createComponentName(name, "helm-repo"),
      {
        repositoryName: `${args.projectName}-helm-charts`,
        maxImageCount: 10,
        tags: args.tags,
      },
      opts
    );

    this.appRepositoryUrl = this.appRepository.repositoryUrl;
    this.helmRepositoryUrl = this.helmRepository.repositoryUrl;
  }

  /**
   * Get ECR authentication tokens for both repositories
   *
   * @returns Object containing auth tokens
   */
  public getAuthTokens(): {
    appRepoAuth: pulumi.Output<string>;
    helmRepoAuth: pulumi.Output<string>;
  } {
    return {
      appRepoAuth: this.appRepository.getAuthToken(),
      helmRepoAuth: this.helmRepository.getAuthToken(),
    };
  }
}

export { EcrRepositoryResource };
