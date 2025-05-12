import * as pulumi from "@pulumi/pulumi";
import { EcrRepository } from "@components/ecr-repository";
import * as aws from "@pulumi/aws";

export interface EcrRepositoryArgs {
  repositoryName: string;
  maxImageCount?: number;
  forceDelete?: boolean;
  tags?: Record<string, string>;
}

export class EcrRepositoryResource {
  public readonly repository: EcrRepository;
  public readonly repositoryUrl: pulumi.Output<string>;

  constructor(
    name: string,
    args: EcrRepositoryArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    this.repository = this.createRepository(name, args, opts);
    this.repositoryUrl = this.repository.repositoryUrl;
  }

  private createRepository(
    name: string,
    args: EcrRepositoryArgs,
    opts?: pulumi.ComponentResourceOptions
  ): EcrRepository {
    return new EcrRepository(
      name,
      {
        repositoryName: args.repositoryName,
        lifecyclePolicy: this.createLifecyclePolicy(args.maxImageCount),
        forceDelete: this.getForceDeleteValue(args.forceDelete),
        tags: args.tags,
      },
      opts
    );
  }

  private createLifecyclePolicy(
    maxImageCount?: number
  ): { maxImageCount?: number } | undefined {
    return maxImageCount ? { maxImageCount } : undefined;
  }

  private getForceDeleteValue(forceDelete?: boolean): boolean {
    return forceDelete !== undefined ? forceDelete : true;
  }

  public getAuthToken(): pulumi.Output<string> {
    return pulumi.output(aws.ecr.getAuthorizationToken()).password;
  }
}
