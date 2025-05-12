import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import * as docker from "@pulumi/docker";
import * as path from "path";
import * as fs from "fs";
import * as yaml from "yaml";
import { EcrRepositoryResource } from "@infrastructure/storage/ecr";
import { HelmChartPipeline } from "@components/helm-chart-pipeline";
import { createComponentName } from "@utils/naming";
import { ServiceConfig, DatabaseConfig } from "@config/types";

/**
 * Configuration arguments for the UserServiceDeployment
 *
 * Think of this as the detailed construction specifications for building
 * the user service building, including materials, layout, and connections
 */
export interface UserServiceDeploymentArgs {
  /**
   * Namespace to deploy into
   *
   * Like the specific plot of land where we'll build
   */
  namespace: pulumi.Input<string>;

  /**
   * Application ECR repository
   *
   * Like the warehouse containing custom-built components (Docker images)
   * for our user administration building
   */
  appRepository: EcrRepositoryResource;

  /**
   * Helm chart ECR repository
   *
   * Like the blueprint repository containing the approved
   * architectural designs for user services
   */
  helmRepository: EcrRepositoryResource;

  /**
   * Environment name (dev, staging, production)
   *
   * Like specifying if we're building a prototype, a model,
   * or the final production building
   */
  environment: string;

  /**
   * Project name
   *
   * Like the master project that this building is part of
   * (e.g., "City Center Revitalization")
   */
  projectName: string;

  /**
   * Service configuration
   *
   * Like detailed specifications for the building's capacity,
   * materials, and features
   */
  serviceConfig: ServiceConfig;

  /**
   * Database configuration
   *
   * Like specifications for the records storage system
   * in the building
   */
  dbConfig: DatabaseConfig;

  /**
   * Database service name
   *
   * Like the name of the records department service
   * that this building will connect to
   */
  dbServiceName: pulumi.Input<string>;

  /**
   * AWS region
   *
   * Like the geographic region for building materials
   * and service connections
   */
  region?: string;

  /**
   * Pre-loaded chart values (optional)
   *
   * Like pre-approved building plans that have already gone through review
   */
  preloadedChartValues?: any;

  /**
   * Override values file path (optional)
   *
   * Like a custom specifications document that overrides standard plans
   */
  overrideValuesPath?: string;
}

/**
 * UserServiceDeployment manages the deployment of the user service
 *
 * Think of this as the construction manager who handles the actual
 * building construction process using the blueprints and materials
 */
export class UserServiceDeployment {
  /**
   * Docker image for the service
   *
   * Like the custom prefabricated components for our building,
   * packaged and ready for installation
   */
  public readonly image: docker.Image;

  /**
   * Helm deployment
   *
   * The actual construction project that uses our blueprints (Helm charts)
   * to build the user service building
   */
  public readonly helmDeployment: HelmChartPipeline;

  /**
   * HTTP endpoint for the service
   *
   * Like the main entrance address of our finished building,
   * where visitors can enter
   */
  public readonly httpEndpoint: pulumi.Output<string>;

  /**
   * gRPC endpoint for the service
   *
   * Like the service entrance address where other city services
   * connect to our building (internal communications)
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
    // Like manufacturing and delivering the custom building components
    this.image = this.buildServiceImage(resourceName, args);

    // Deploy the Helm chart
    this.helmDeployment = this.deployHelmChart(
      resourceName,
      args,
      provider,
      opts
    );

    // Get service endpoints with proper fallbacks
    const defaultEndpoint = pulumi.output("Pending");

    // Initialize with default values
    this.httpEndpoint = defaultEndpoint;
    this.grpcEndpoint = defaultEndpoint;

    // Set them to the actual endpoints if the deployment exists
    if (this.helmDeployment.helmDeployment?.chart) {
      this.httpEndpoint =
        this.getHttpEndpoint(this.helmDeployment, name) || defaultEndpoint;
      this.grpcEndpoint =
        this.getGrpcEndpoint(this.helmDeployment, name) || defaultEndpoint;
    }
  }

  /**
   * Builds and pushes the service Docker image
   *
   * Like manufacturing the custom components for our building
   * and delivering them to the construction site
   */
  private buildServiceImage(
    name: string,
    args: UserServiceDeploymentArgs
  ): docker.Image {
    return new docker.Image(name, {
      imageName: pulumi.interpolate`${args.appRepository.repositoryUrl}:latest`,
      build: {
        context: path.resolve(__dirname, "../../../../user-service"),
        dockerfile: path.resolve(
          __dirname,
          "../../../../user-service/Dockerfile"
        ),
        platform: "linux/amd64",
        args: {
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
   *
   * Like managing the construction process using the blueprint (Helm chart)
   */
  private deployHelmChart(
    name: string,
    args: UserServiceDeploymentArgs,
    provider: k8s.Provider,
    opts?: pulumi.ComponentResourceOptions
  ): HelmChartPipeline {
    const isProd = args.environment === "production";

    // Get a string region value
    let regionValue = args.region;
    if (!regionValue || typeof regionValue !== "string") {
      regionValue = process.env.AWS_REGION || "us-west-2";
      console.log(`Using default region: ${regionValue}`);
    }

    // Load chart values in proper precedence order
    // Like gathering all the building specifications from various sources
    const chartValues = this.loadAndMergeChartValues(args);

    // Log what we're using
    console.log(`Deploying ${name} with environment: ${args.environment}`);

    // Create and return the Helm chart pipeline
    // This combines chart packaging, storage, and deployment
    return new HelmChartPipeline(
      name,
      {
        // ECR repository details
        ecrRepositoryName: args.helmRepository.repository.repository.name,
        region: regionValue,

        // Chart information
        helmChartPath: path.resolve(__dirname, "../../../charts/user-service"),
        chartName: "user-service",
        chartVersion: "0.1.0",

        // Deployment details
        namespace: args.namespace,
        releaseName: "user-service",

        // Use our merged values
        values: chartValues,

        // Environment for automatic loading of values-{env}.yaml
        environment: args.environment,

        // Always deploy the chart
        deploy: true,

        // Transformations to help with Pulumi preview
        transforms: [
          (obj: any) => {
            // Add metadata to help Pulumi preview
            if (obj.metadata && !obj.metadata.annotations) {
              obj.metadata.annotations = {};
            }
            if (obj.metadata && obj.metadata.annotations) {
              obj.metadata.annotations["pulumi.com/preview"] = "true";
            }
            return obj;
          },
        ],
      },
      {
        provider: provider,
        dependsOn: [this.image],
        ...opts,
      }
    );
  }

  /**
   * Load and merge chart values in proper precedence order
   *
   * Like gathering all the building specifications from various sources
   * and merging them according to priority
   */
  private loadAndMergeChartValues(args: UserServiceDeploymentArgs): any {
    const isProd = args.environment === "production";

    // Start with pre-loaded values if provided
    // Otherwise, the HelmChartPipeline will load them automatically
    const baseValues = args.preloadedChartValues || {};

    // Load override values if a path is specified
    let overrideValues = {};
    if (args.overrideValuesPath && fs.existsSync(args.overrideValuesPath)) {
      try {
        const overrideContent = fs.readFileSync(
          args.overrideValuesPath,
          "utf8"
        );
        overrideValues = yaml.parse(overrideContent);
        console.log(`Loaded override values from ${args.overrideValuesPath}`);
      } catch (error) {
        console.warn(`Failed to load override values: ${error}`);
      }
    }

    // Define runtime values that must be calculated by Pulumi
    // These take highest precedence
    const runtimeValues = {
      // Global settings
      global: {
        environment: args.environment,
      },

      // Container image settings
      image: {
        repository: args.appRepository.repositoryUrl,
        tag: "latest",
        pullPolicy: isProd ? "Always" : "IfNotPresent",
      },

      // Service configuration
      replicaCount: args.serviceConfig.replicaCount,
      resources: args.serviceConfig.resources,

      // Node selectors
      nodeSelector: {
        "workload-type": "standard",
      },

      // Auto-scaling configuration
      autoscaling: args.serviceConfig.autoscaling.enabled
        ? {
            enabled: args.serviceConfig.autoscaling.enabled,
            minReplicas: args.serviceConfig.autoscaling.minReplicas,
            maxReplicas: args.serviceConfig.autoscaling.maxReplicas,
            targetCPUUtilizationPercentage:
              args.serviceConfig.autoscaling.targetCpuUtilization,
            targetMemoryUtilizationPercentage:
              args.serviceConfig.autoscaling.targetMemoryUtilization,
          }
        : undefined,

      // Ingress configuration
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

      // Database connection details
      config: {
        db: {
          host: args.dbServiceName,
          user: args.dbConfig.dbUsername,
          password: args.dbConfig.dbPassword,
          name: args.dbConfig.dbName,
          schema: args.dbConfig.dbSchema,
          port: args.dbConfig.dbPort,
          sslmode: isProd ? "require" : "disable",
          driver: "postgres",
        },
      },

      // RBAC settings
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
    };

    // Merge values in order of increasing precedence:
    // 1. baseValues (from values.yaml and values-{env}.yaml)
    // 2. overrideValues (from override file if provided)
    // 3. runtimeValues (dynamic values from Pulumi)
    return this.deepMerge(
      this.deepMerge(baseValues, overrideValues),
      runtimeValues
    );
  }

  /**
   * Deeply merge objects (helper method)
   *
   * Like carefully integrating multiple sets of building specifications
   * with later ones taking precedence over earlier ones
   */
  private deepMerge(target: any, source: any): any {
    // Return source if target is not an object
    if (!this.isObject(target)) return source;

    // Return target if source is not an object
    if (!this.isObject(source)) return target;

    // Create a new object to avoid mutating inputs
    const output = { ...target };

    // Merge properties from source into target
    Object.keys(source).forEach((key) => {
      // If both target and source have an object at this key, recurse
      if (this.isObject(target[key]) && this.isObject(source[key])) {
        output[key] = this.deepMerge(target[key], source[key]);
      }
      // If source key is undefined, don't override target
      else if (source[key] === undefined) {
        // Keep target value
      }
      // Otherwise use the source value
      else {
        output[key] = source[key];
      }
    });

    return output;
  }

  /**
   * Check if a value is an object (helper method)
   */
  private isObject(item: any): boolean {
    return item && typeof item === "object" && !Array.isArray(item);
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
