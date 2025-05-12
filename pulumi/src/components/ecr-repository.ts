// pulumi/src/components/ecr-repository.ts

import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface EcrRepositoryArgs {
  repositoryName: pulumi.Input<string>;
  scanOnPush?: boolean;
  imageTagMutability?: "MUTABLE" | "IMMUTABLE";
  lifecyclePolicy?: {
    maxImageCount?: number;
    maxImageAge?: number;
  };
  forceDelete?: boolean;
  tags?: { [key: string]: string };
}

export class EcrRepository extends pulumi.ComponentResource {
  public readonly repository: aws.ecr.Repository;
  public readonly repositoryUrl: pulumi.Output<string>;
  public readonly repositoryArn: pulumi.Output<string>;

  constructor(
    name: string,
    args: EcrRepositoryArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:EcrRepository", name, {}, opts);

    this.repository = this.createRepository(name, args);

    if (this.hasLifecycleRules(args.lifecyclePolicy)) {
      this.applyLifecyclePolicy(name, args.lifecyclePolicy);
    }

    this.repositoryUrl = this.repository.repositoryUrl;
    this.repositoryArn = this.repository.arn;

    this.registerOutputs({
      repositoryUrl: this.repositoryUrl,
      repositoryArn: this.repositoryArn,
    });
  }

  private createRepository(
    name: string,
    args: EcrRepositoryArgs
  ): aws.ecr.Repository {
    return new aws.ecr.Repository(
      name,
      {
        name: args.repositoryName,
        imageScanningConfiguration: {
          scanOnPush: this.getScanOnPushValue(args.scanOnPush),
        },
        imageTagMutability: args.imageTagMutability || "MUTABLE",
        forceDelete: this.getForceDeleteValue(args.forceDelete),
        tags: this.createRepositoryTags(args.tags),
      },
      { parent: this }
    );
  }

  private getScanOnPushValue(scanOnPush?: boolean): boolean {
    return scanOnPush !== undefined ? scanOnPush : true;
  }

  private getForceDeleteValue(forceDelete?: boolean): boolean {
    return forceDelete !== undefined ? forceDelete : true;
  }

  private createRepositoryTags(tags?: { [key: string]: string }): {
    [key: string]: string;
  } {
    return {
      ...(tags || {}),
      ManagedBy: "pulumi",
    };
  }

  private hasLifecycleRules(
    lifecyclePolicy?: EcrRepositoryArgs["lifecyclePolicy"]
  ): boolean {
    return (
      !!lifecyclePolicy &&
      (!!lifecyclePolicy.maxImageCount || !!lifecyclePolicy.maxImageAge)
    );
  }

  private applyLifecyclePolicy(
    name: string,
    lifecyclePolicy?: EcrRepositoryArgs["lifecyclePolicy"]
  ): aws.ecr.LifecyclePolicy | undefined {
    if (!lifecyclePolicy) return undefined;

    const rules = this.buildLifecycleRules(lifecyclePolicy);

    if (rules.length === 0) return undefined;

    return new aws.ecr.LifecyclePolicy(
      `${name}-lifecycle-policy`,
      {
        repository: this.repository.name,
        policy: JSON.stringify({ rules }),
      },
      { parent: this }
    );
  }

  private buildLifecycleRules(
    lifecyclePolicy: EcrRepositoryArgs["lifecyclePolicy"]
  ): any[] {
    const rules: any[] = [];

    if (lifecyclePolicy?.maxImageCount) {
      rules.push(this.createMaxImageCountRule(lifecyclePolicy.maxImageCount));
    }

    // Add a rule to delete untagged images
    rules.push(this.createUntaggedImageRule());

    if (lifecyclePolicy?.maxImageAge) {
      rules.push(
        this.createMaxImageAgeRule(
          lifecyclePolicy.maxImageAge,
          rules.length + 1
        )
      );
    }

    return rules;
  }

  private createUntaggedImageRule(): any {
    return {
      rulePriority: 2,
      description: "Remove untagged images",
      selection: {
        tagStatus: "untagged",
        countType: "imageCountMoreThan",
        countNumber: 0,
      },
      action: {
        type: "expire",
      },
    };
  }

  private createMaxImageCountRule(maxImageCount: number): any {
    return {
      rulePriority: 1,
      description: `Keep only the last ${maxImageCount} images`,
      selection: {
        tagStatus: "any",
        countType: "imageCountMoreThan",
        countNumber: maxImageCount,
      },
      action: {
        type: "expire",
      },
    };
  }

  private createMaxImageAgeRule(maxImageAge: number, priority: number): any {
    return {
      rulePriority: priority,
      description: `Expire images older than ${maxImageAge} days`,
      selection: {
        tagStatus: "any",
        countType: "sinceImagePushed",
        countUnit: "days",
        countNumber: maxImageAge,
      },
      action: {
        type: "expire",
      },
    };
  }

  public getAuthToken(): pulumi.Output<string> {
    return pulumi.output(aws.ecr.getAuthorizationToken()).password;
  }
}
