import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import { createComponentName } from "@utils/naming";
import { UserServiceDeployment } from "./deployment";
import { EcrRepositoryResource } from "@infrastructure/storage/ecr";
import { EksClusterResource } from "@infrastructure/compute/eks";
import {
  getAppConfig,
  getDatabaseConfig,
  getUserServiceConfig,
} from "@config/environment";

/**
 * Configuration arguments for the UserService
 */
export interface UserServiceArgs {
  /** EKS cluster to deploy into */
  cluster: EksClusterResource;

  /** Kubernetes provider */
  k8sProvider: k8s.Provider;

  /** Application repository */
  appRepository: EcrRepositoryResource;

  /** Helm chart repository */
  helmRepository: EcrRepositoryResource;

  /** Environment name (dev, staging, production) */
  environment: string;

  /** Database password (secret) */
  dbPassword: pulumi.Input<string>;

  /** AWS region */
  region: string;

  /** Tags to apply to resources */
  tags?: Record<string, string>;

  /** Pre-loaded chart values */
  preloadedValues?: any;
}

/**
 * UserService manages all resources for the user service
 */
export class UserService {
  /** Kubernetes namespace */
  public readonly namespace: k8s.core.v1.Namespace;

  /** Service deployment */
  public readonly deployment: UserServiceDeployment;

  /** HTTP endpoint */
  public readonly httpEndpoint: pulumi.Output<string>;

  /** gRPC endpoint */
  public readonly grpcEndpoint: pulumi.Output<string>;

  /**
   * Creates a new user service
   */
  constructor(
    name: string,
    args: UserServiceArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    // Get configurations
    const appConfig = getAppConfig();
    const dbConfig = getDatabaseConfig();
    const serviceConfig = getUserServiceConfig();

    // Create namespace for the service
    this.namespace = new k8s.core.v1.Namespace(
      createComponentName(name, "namespace"),
      {
        metadata: {
          name: name,
          labels: {
            "app.kubernetes.io/part-of": name,
            "app.kubernetes.io/managed-by": "pulumi",
          },
        },
      },
      {
        provider: args.k8sProvider,
        ...opts,
      }
    );

    // Create database password secret
    const dbSecret = new k8s.core.v1.Secret(
      createComponentName(name, "db-secret"),
      {
        metadata: {
          namespace: this.namespace.metadata.name,
          name: `${name}-db-credentials`,
        },
        type: "Opaque",
        stringData: {
          username: dbConfig.dbUsername,
          password: args.dbPassword,
        },
      },
      {
        provider: args.k8sProvider,
        dependsOn: [this.namespace],
        ...opts,
      }
    );

    // Internal database service name
    const dbServiceName = `${name}-postgresql`;

    // Deploy the service
    this.deployment = new UserServiceDeployment(
      name,
      {
        namespace: this.namespace.metadata.name,
        appRepository: args.appRepository,
        helmRepository: args.helmRepository,
        environment: args.environment,
        projectName: appConfig.projectName,
        serviceConfig: serviceConfig,
        dbConfig: dbConfig,
        dbServiceName: dbServiceName,
        region: args.region,
        preloadedChartValues: args.preloadedValues,
      },
      args.k8sProvider,
      {
        dependsOn: [this.namespace, dbSecret],
        ...opts,
      }
    );

    // Export endpoints
    this.httpEndpoint = this.deployment.httpEndpoint;
    this.grpcEndpoint = this.deployment.grpcEndpoint;
  }
}
