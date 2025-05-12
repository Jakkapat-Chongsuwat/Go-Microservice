import * as pulumi from "@pulumi/pulumi";
import { EcrRepositoryResource } from "./ecr";
import { createComponentName } from "@utils/naming";

export interface StorageStackArgs {
  projectName: string;
  tags?: Record<string, string>;
}

export class StorageStack {
  public readonly appRepository: EcrRepositoryResource;
  public readonly helmRepository: EcrRepositoryResource;
  public readonly appRepositoryUrl: pulumi.Output<string>;
  public readonly helmRepositoryUrl: pulumi.Output<string>;

  constructor(
    name: string,
    args: StorageStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    this.appRepository = this.createAppRepository(name, args, opts);
    this.helmRepository = this.createHelmRepository(name, args, opts);

    this.appRepositoryUrl = this.appRepository.repositoryUrl;
    this.helmRepositoryUrl = this.helmRepository.repositoryUrl;
  }

  private createAppRepository(
    name: string,
    args: StorageStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ): EcrRepositoryResource {
    return new EcrRepositoryResource(
      createComponentName(name, "app-repo"),
      {
        repositoryName: this.sanitizeRepositoryName(
          `${args.projectName}-user-service`
        ),
        maxImageCount: 10,
        forceDelete: true,
        tags: args.tags,
      },
      opts
    );
  }

  private createHelmRepository(
    name: string,
    args: StorageStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ): EcrRepositoryResource {
    return new EcrRepositoryResource(
      createComponentName(name, "helm-repo"),
      {
        repositoryName: this.sanitizeRepositoryName(
          `${args.projectName}-helm-charts`
        ),
        maxImageCount: 10,
        forceDelete: true,
        tags: args.tags,
      },
      opts
    );
  }

  public getAuthTokens(): {
    appRepoAuth: pulumi.Output<string>;
    helmRepoAuth: pulumi.Output<string>;
  } {
    return {
      appRepoAuth: this.appRepository.getAuthToken(),
      helmRepoAuth: this.helmRepository.getAuthToken(),
    };
  }

  private sanitizeRepositoryName(name: string): string {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "-")
      .replace(/-+/g, "-");
  }
}

export { EcrRepositoryResource };
