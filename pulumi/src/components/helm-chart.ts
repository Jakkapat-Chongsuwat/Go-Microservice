// pulumi/src/components/helm-chart.ts

import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import * as fs from "fs";
import * as yaml from "yaml";

/**
 * Configuration options for deploying a Helm chart.
 *
 * Think of this as the architectural blueprint for a new building in the city.
 * It defines all the specifications before construction begins.
 */
export interface HelmChartArgs {
  /**
   * The namespace to deploy the chart to.
   *
   * Like the specific district or zone in the city where the building will be constructed.
   * Different namespaces provide isolation between applications, just as districts
   * separate residential, commercial, and industrial areas.
   */
  namespace: pulumi.Input<string>;

  /**
   * Create namespace if it doesn't exist.
   * @default true
   *
   * Like automatically rezoning an area if the specified district doesn't exist yet.
   * If set to false and the namespace doesn't exist, deployment will fail,
   * similar to trying to build in a non-existent district.
   */
  createNamespace?: boolean;

  /**
   * Chart location - can be a local path or OCI repository URL.
   *
   * The architectural plans for the building - can be sourced from local files
   * or retrieved from a centralized repository of approved designs.
   */
  chart: pulumi.Input<string>;

  /**
   * Chart version to deploy.
   *
   * The specific version of the building design to use.
   * If omitted, the latest version will be used, which can lead to inconsistent
   * deployments if the chart is updated between deployments.
   */
  version?: pulumi.Input<string>;

  /**
   * Values to override in the chart.
   *
   * Customizations to the standard building design - like specifying different materials,
   * changing the height, or adding custom features not in the original plans.
   */
  values?: { [key: string]: any };

  /**
   * Optional chart value files to include.
   *
   * Additional specification documents that provide pre-defined customizations.
   * These are merged with the values parameter, with the latter taking precedence
   * in case of conflicts.
   */
  valueFiles?: string[];

  /**
   * OCI registry authentication (if using OCI chart).
   *
   * The credentials needed to access a secure architectural designs repository.
   * Without this, you can't retrieve plans from private repositories.
   */
  repositoryOpts?: {
    /**
     * Repository URL.
     *
     * The address of the architectural designs repository.
     */
    repository: pulumi.Input<string>;

    /**
     * Username for repository.
     *
     * Your identity credential to access the secured repository.
     * If omitted for a private repository, authentication will fail.
     */
    username?: pulumi.Input<string>;

    /**
     * Password for repository.
     *
     * Your access credential to the secured repository.
     * If omitted for a private repository, authentication will fail.
     */
    password?: pulumi.Input<string>;
  };

  /**
   * Transformations to apply to the chart resources.
   *
   * Like having architects review and adjust plans during construction.
   * This allows you to modify the Kubernetes resources before they are created.
   */
  transformations?: ((obj: any, opts?: pulumi.CustomResourceOptions) => void)[];

  /**
   * Skip await logic for resources
   *
   * Like proceeding with other construction tasks without waiting
   * for this building to be fully completed and inspected.
   */
  skipAwait?: boolean;

  /**
   * Whether to include test resources
   *
   * Like including testing facilities in the building that would normally
   * only be used during construction verification.
   */
  includeTestHookResources?: boolean;

  /**
   * Skip rendering CRDs
   *
   * Like omitting certain standardized utility infrastructure that might
   * be managed separately by the city.
   */
  skipCRDRendering?: boolean;
}

/**
 * HelmChart is a component resource that deploys a Helm chart to a Kubernetes cluster.
 *
 * Think of this as the construction company that takes the architectural plans,
 * prepares the building site, and handles all aspects of constructing the building
 * in the specified district of the city.
 */
export class HelmChart extends pulumi.ComponentResource {
  /**
   * The underlying Helm chart deployment.
   *
   * Represents the actual building constructed in the city.
   */
  public readonly chart: k8s.helm.v3.Chart;

  /**
   * The namespace where the chart is deployed.
   *
   * The district or zone in the city where the building was constructed.
   * May be undefined if an existing namespace was used.
   */
  public readonly namespace: k8s.core.v1.Namespace | undefined;

  constructor(
    name: string,
    args: HelmChartArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:HelmChart", name, {}, opts);

    // Create namespace if requested and it doesn't exist
    // Like establishing a new district in the city if it doesn't exist yet
    if (args.createNamespace !== false) {
      this.namespace = new k8s.core.v1.Namespace(
        name,
        {
          metadata: {
            name: args.namespace,
          },
        },
        { parent: this }
      );
    }

    // Prepare repository options for fetching the chart
    // Like preparing the credentials to access the architectural plans repository
    const fetchOpts = args.repositoryOpts
      ? {
          repo: args.repositoryOpts.repository,
          username: args.repositoryOpts.username,
          password: args.repositoryOpts.password,
        }
      : undefined;

    // Combine values from files and provided values
    // Like merging the standard specifications with client-specific customizations
    let combinedValues = args.values || {};
    if (args.valueFiles && args.valueFiles.length > 0) {
      for (const filePath of args.valueFiles) {
        try {
          if (fs.existsSync(filePath)) {
            const valueFileContent = fs.readFileSync(filePath, "utf8");
            const valueFileObj = yaml.parse(valueFileContent);
            combinedValues = this.deepMerge(combinedValues, valueFileObj);
          } else {
            console.log(`Values file ${filePath} does not exist, skipping`);
          }
        } catch (error) {
          console.warn(
            `Failed to load values file ${filePath}: ${error}. Continuing with existing values.`
          );
        }
      }
    }

    // Deploy the chart
    // Like starting the actual construction process with all specifications
    this.chart = new k8s.helm.v3.Chart(
      name,
      {
        chart: args.chart,
        version: args.version,
        namespace: args.namespace,
        values: combinedValues,
        fetchOpts: fetchOpts,
        transformations: args.transformations,
        skipAwait: args.skipAwait,
        includeTestHookResources: args.includeTestHookResources,
        skipCRDRendering: args.skipCRDRendering,
      },
      {
        parent: this,
        dependsOn: this.namespace ? [this.namespace] : undefined,
        ...opts,
      }
    );

    // Register outputs for this component
    // Like filing the final building permits and records with the city
    this.registerOutputs({
      chart: this.chart,
      namespace: this.namespace,
    });
  }

  /**
   * Deep merge two objects
   *
   * Like carefully integrating two sets of building specifications,
   * with the source specifications taking precedence where they conflict.
   */
  private deepMerge(target: any, source: any): any {
    const output = { ...target };

    if (this.isObject(target) && this.isObject(source)) {
      Object.keys(source).forEach((key) => {
        if (this.isObject(source[key])) {
          if (!(key in target)) {
            Object.assign(output, { [key]: source[key] });
          } else {
            output[key] = this.deepMerge(target[key], source[key]);
          }
        } else {
          Object.assign(output, { [key]: source[key] });
        }
      });
    }

    return output;
  }

  /**
   * Check if a value is an object
   *
   * Like verifying that what we're looking at is actually a set of building
   * specifications and not just a list of materials.
   */
  private isObject(item: any): boolean {
    return item && typeof item === "object" && !Array.isArray(item);
  }
}
