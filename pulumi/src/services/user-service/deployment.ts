import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as k8s from "@pulumi/kubernetes";
import * as docker from "@pulumi/docker";
import * as path from "path";
import { EcrRepositoryResource } from "../../infrastructure/storage/ecr";
import { HelmChartPipeline } from "../../components/helm-chart-pipeline";
import { createComponentName } from "../../utils/naming";
import { ServiceConfig, DatabaseConfig } from "../../config/types";

export interface UserServiceDeploymentArgs {
  /**
   * Namespace to deploy into
   */
  namespace: pulumi.Input<string>;

  /**
   * Application ECR repository
   */
  appRepository: EcrRepositoryResource;

  /**
   * Helm chart ECR repository
   */
  helmRepository: EcrRepositoryResource;

  /**
   * Environment name (dev, staging, production)
   */
  environment: string;

  /**
   * Project name
   */
  projectName: string;

  /**
   * Service configuration
   */
  serviceConfig: ServiceConfig;

  /**
   * Database configuration
   */
  dbConfig: DatabaseConfig;

  /**
   * Database service name
   */
  dbServiceName: pulumi.Input<string>;

  /**
   * AWS region
   */
  region?: string;
}

/**
 * UserServiceDeployment manages the deployment of the user service
 */
export class UserServiceDeployment {
  /**
   * Docker image for the service
   */
  public readonly image: docker.Image;

  /**
   * Helm deployment
   */
  public readonly helmDeployment: HelmChartPipeline;

  /**
   * HTTP endpoint for the service
   */
  public readonly httpEndpoint: pulumi.Output<string>;

  /**
   * gRPC endpoint for the service
   */
  public readonly grpcEndpoint: pulumi.Output<string>;

  /**
   * Creates a new user service deployment
   *
   * @param name Resource name
   * @param args Deployment configuration
   * @param provider Kubernetes provider
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: UserServiceDeploymentArgs,
    provider: k8s.Provider,
    opts?: pulumi.ComponentResourceOptions
  ) {
    const resourceName = createComponentName(name, "deployment");

    // Build and push the Docker image
    this.image = this.buildServiceImage(resourceName, args);

    // Deploy the Helm chart
    this.helmDeployment = this.deployHelmChart(
      resourceName,
      args,
      provider,
      opts
    );

    // Get service endpoints with proper fallbacks to ensure they're always Output<string>
    const defaultEndpoint = pulumi.output("Pending");

    // Initialize with default values to satisfy TypeScript
    this.httpEndpoint = defaultEndpoint;
    this.grpcEndpoint = defaultEndpoint;

    // Then set them to the actual endpoints if the deployment exists
    if (this.helmDeployment.helmDeployment?.chart) {
      this.httpEndpoint =
        this.getHttpEndpoint(this.helmDeployment, name) || defaultEndpoint;
      this.grpcEndpoint =
        this.getGrpcEndpoint(this.helmDeployment, name) || defaultEndpoint;
    }
  }

  /**
   * Builds and pushes the service Docker image
   */
  private buildServiceImage(
    name: string,
    args: UserServiceDeploymentArgs
  ): docker.Image {
    return new docker.Image(name, {
      imageName: pulumi.interpolate`${args.appRepository.repositoryUrl}:latest`,
      build: {
        context: "../../../", // Location of your source code (Go-microservice root)
        dockerfile: "../../../Dockerfile",
        args: {
          // Optionally pass build arguments
          GO_ENV: args.environment,
        },
      },
      registry: {
        server: args.appRepository.repositoryUrl.apply(
          (url) => url.split("/")[0]
        ),
        username: "AWS",
        password: args.appRepository.getAuthToken(),
      },
    });
  }

  /**
   * Deploys the service using a Helm chart
   */
  private deployHelmChart(
    name: string,
    args: UserServiceDeploymentArgs,
    provider: k8s.Provider,
    opts?: pulumi.ComponentResourceOptions
  ): HelmChartPipeline {
    const isProd = args.environment === "production";

    // Get a string region value either from args or by awaiting the region output
    let regionValue = args.region;

    // If no region provided or it's an Output, use a default directly
    if (!regionValue || typeof regionValue !== "string") {
      // For Pulumi, we can use a hardcoded default region if needed
      // Or fetch from environment variables
      regionValue = process.env.AWS_REGION || "us-west-2";
      console.log(`Using default region: ${regionValue}`);
    }

    const chartPath = path.join(__dirname, "../../../charts/user-service");

    return new HelmChartPipeline(
      name,
      {
        ecrRepositoryName: args.helmRepository.repository.repository.name,
        region: regionValue,
        helmChartPath: chartPath,
        chartName: "user-service",
        chartVersion: "0.1.0",
        namespace: args.namespace,
        releaseName: "user-service",
        values: {
          // Set environment-specific values
          global: {
            environment: args.environment,
          },

          // Reference the built image
          image: {
            repository: args.appRepository.repositoryUrl,
            tag: "latest",
            pullPolicy: isProd ? "Always" : "IfNotPresent",
          },

          // Set replica count based on environment
          replicaCount: args.serviceConfig.replicaCount,

          // Configure resources based on environment
          resources: args.serviceConfig.resources,

          // Select node group based on workload type
          nodeSelector: {
            "workload-type": "standard",
          },

          // Auto-scaling configuration
          autoscaling: {
            enabled: args.serviceConfig.autoscaling.enabled,
            minReplicas: args.serviceConfig.autoscaling.minReplicas,
            maxReplicas: args.serviceConfig.autoscaling.maxReplicas,
            targetCPUUtilizationPercentage:
              args.serviceConfig.autoscaling.targetCpuUtilization,
            targetMemoryUtilizationPercentage:
              args.serviceConfig.autoscaling.targetMemoryUtilization,
          },

          // Configure ingress based on environment
          ingress: {
            enabled: true,
            className: "alb",
            annotations: {
              "kubernetes.io/ingress.class": "alb",
              "alb.ingress.kubernetes.io/scheme": "internet-facing",
              "alb.ingress.kubernetes.io/target-type": "ip",
              "alb.ingress.kubernetes.io/group.name": `${args.projectName}-${args.environment}`,
            },
            hosts: [
              {
                host: args.serviceConfig.hostname,
                paths: [
                  {
                    path: "/",
                    pathType: "Prefix",
                    port: 50052,
                  },
                ],
              },
            ],
            tls: isProd
              ? [
                  {
                    secretName: "user-service-tls",
                    hosts: [args.serviceConfig.hostname],
                  },
                ]
              : [],
          },

          // Configure database connection
          config: {
            db: {
              host: args.dbServiceName,
              user: args.dbConfig.dbUsername,
              password: args.dbConfig.dbPassword,
              name: args.dbConfig.dbName,
              schema: args.dbConfig.dbSchema,
              port: args.dbConfig.dbPort,
              sslmode: "disable",
              driver: "postgres",
            },
          },

          // Configure RBAC with appropriate permissions
          rbac: {
            create: true,
            rules: [
              {
                apiGroups: [""],
                resources: ["configmaps", "secrets"],
                verbs: ["get"],
              },
              {
                apiGroups: [""],
                resources: ["pods"],
                verbs: ["get", "list"],
              },
            ],
          },
        },
        deploy: true,
      },
      {
        provider: provider,
        dependsOn: [this.image],
        ...opts,
      }
    );
  }

  /**
   * Get the HTTP endpoint for the service
   */
  private getHttpEndpoint(
    helmDeployment: HelmChartPipeline,
    name: string
  ): pulumi.Output<string> | undefined {
    try {
      return helmDeployment.helmDeployment?.chart
        .getResourceProperty(
          "networking.k8s.io/v1/Ingress",
          `${name}/${name}`,
          "status"
        )
        .apply(
          (status) => status?.loadBalancer?.ingress?.[0]?.hostname || "Pending"
        );
    } catch (error) {
      console.warn("Failed to get HTTP endpoint:", error);
      return pulumi.output("Pending");
    }
  }

  /**
   * Get the gRPC endpoint for the service
   */
  private getGrpcEndpoint(
    helmDeployment: HelmChartPipeline,
    name: string
  ): pulumi.Output<string> | undefined {
    try {
      return helmDeployment.helmDeployment?.chart
        .getResourceProperty("v1/Service", `${name}/${name}`, "status")
        .apply(
          (status) =>
            `${status?.loadBalancer?.ingress?.[0]?.hostname || "Pending"}:50051`
        );
    } catch (error) {
      console.warn("Failed to get gRPC endpoint:", error);
      return pulumi.output("Pending:50051");
    }
  }
}
