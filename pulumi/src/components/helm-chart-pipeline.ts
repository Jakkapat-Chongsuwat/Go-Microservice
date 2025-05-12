// pulumi/src/components/helm-chart-pipeline.ts

import * as pulumi from "@pulumi/pulumi";
import * as command from "@pulumi/command";
import * as aws from "@pulumi/aws";
import * as path from "path";
import * as fs from "fs";
import * as yaml from "yaml";
import { EcrRepository } from "./ecr-repository";
import { HelmChart } from "./helm-chart";

/**
 * Configuration options for the Helm chart deployment pipeline.
 *
 * Think of this as the project plan for constructing and storing
 * standardized building designs in the city's architectural repository.
 */
export interface HelmChartPipelineArgs {
  /**
   * Name of the ECR repository to create.
   *
   * Like the name of the architectural repository where building designs will be stored.
   * This name should be unique within your AWS account/region.
   */
  ecrRepositoryName: pulumi.Input<string>;

  /**
   * AWS region for ECR.
   *
   * The geographic region where the architectural repository will be located.
   * Different regions may have different compliance requirements and access patterns.
   */
  region: string;

  /**
   * Local path to the Helm chart directory.
   *
   * The local folder where the building design files are stored before being
   * uploaded to the central repository.
   */
  helmChartPath: string;

  /**
   * Chart name to use.
   *
   * The formal name of the building design, used for identification in the repository.
   */
  chartName: string;

  /**
   * Chart version to use.
   *
   * The version number of the building design.
   * Using semantic versioning helps track changes and ensures consistency.
   */
  chartVersion: string;

  /**
   * Deployment target namespace.
   *
   * The district or zone in the city where the building will be constructed.
   */
  namespace: pulumi.Input<string>;

  /**
   * Release name to use for the chart.
   *
   * The specific name for this particular building instance.
   * Multiple buildings can use the same design but have different names.
   */
  releaseName: string;

  /**
   * Values to override in the chart.
   *
   * Customizations applied to the standard building design,
   * like size adjustments, material choices, or feature additions.
   */
  values?: { [key: string]: any };

  /**
   * Environment to use for loading environment-specific values.
   *
   * Like specifying whether this is a residential or commercial district,
   * which affects what types of buildings and features are appropriate.
   */
  environment?: string;

  /**
   * Whether to deploy the chart after pushing.
   * @default true
   *
   * Whether to immediately start construction after uploading the designs,
   * or just store the plans for later use.
   */
  deploy?: boolean;

  /**
   * Tags to apply to ECR repository.
   *
   * Labels that help categorize and identify the repository in the larger system.
   * Important for governance, cost allocation, and organizational purposes.
   */
  tags?: { [key: string]: string };

  /**
   * Optional resource transformations
   *
   * Like modifications to the building plans that need to be made before
   * construction begins, ensuring they meet local regulations or special needs.
   */
  transforms?: ((obj: any) => any)[];

  /**
   * Skip await logic for resources (default: false)
   *
   * Like proceeding with other construction tasks without waiting for
   * the current building to be fully completed and inspected.
   */
  skipAwait?: boolean;

  /**
   * Whether to include test resources (default: false)
   *
   * Like including testing facilities in the building that would normally
   * only be used during construction verification.
   */
  includeTestHookResources?: boolean;

  /**
   * Skip rendering CRDs (default: false)
   *
   * Like omitting certain standardized utility infrastructure that might
   * be managed separately by the city.
   */
  skipCRDRendering?: boolean;
}

/**
 * HelmChartPipeline handles the end-to-end process of packaging, storing, and deploying Helm charts.
 *
 * This is like the complete city planning department that handles creating standardized
 * building designs, storing them in a central repository, and optionally beginning construction.
 */
export class HelmChartPipeline extends pulumi.ComponentResource {
  /**
   * The ECR repository where charts are stored.
   *
   * The architectural repository where building designs are archived.
   */
  public readonly ecrRepository: EcrRepository;

  /**
   * The Helm chart deployment (if deployed).
   *
   * The actual building constructed from the stored design.
   * This is undefined if deploy is set to false.
   */
  public readonly helmDeployment?: HelmChart;

  /**
   * The ECR repository URL.
   *
   * The address of the architectural repository where designs are stored.
   */
  public readonly repositoryUrl: pulumi.Output<string>;

  constructor(
    name: string,
    args: HelmChartPipelineArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:HelmChartPipeline", name, {}, opts);

    // Create ECR repository
    // Like establishing a new architectural repository in the city archives
    this.ecrRepository = new EcrRepository(
      name,
      {
        repositoryName: args.ecrRepositoryName,
        // IMMUTABLE ensures that chart versions can't be overwritten
        // Like ensuring approved building designs can't be altered once finalized
        imageTagMutability: "IMMUTABLE",
        lifecyclePolicy: {
          // Keep only the 10 most recent chart versions
          // Like keeping only the 10 most recent revisions of building designs
          // to prevent the repository from becoming cluttered
          maxImageCount: 10,
        },
        tags: {
          ...(args.tags || {}),
          Service: args.chartName,
        },
      },
      { parent: this }
    );

    this.repositoryUrl = this.ecrRepository.repositoryUrl;

    // Get AWS account ID
    // Like retrieving your city's unique identifier for paperwork
    const callerIdentity = aws.getCallerIdentity();
    const accountIdOutput = pulumi.output(callerIdentity).accountId;
    const ecrDomain = pulumi.interpolate`${accountIdOutput}.dkr.ecr.${args.region}.amazonaws.com`;

    // Load chart values from files before packaging and pushing
    // Like gathering all the building specifications before finalizing the plans
    const chartValues = this.loadChartValues(
      args.helmChartPath,
      args.environment
    );

    // Merge with runtime values
    // Like adding last-minute changes and specific requirements for this particular build
    const mergedValues = {
      ...chartValues,
      ...(args.values || {}),
    };

    // Log what values we're using
    console.log(
      `Deploying ${args.chartName} with merged values from base chart and ${
        args.environment || "default"
      } environment`
    );

    // Package and push Helm chart to ECR
    // Like preparing the building designs and archiving them in the central repository
    const packageAndPushChart = new command.local.Command(
      `${name}-package-and-push`,
      {
        create: pulumi.interpolate`
            cd "${args.helmChartPath}" && \
            helm package . --version ${args.chartVersion} && \
            aws ecr get-login-password --region ${args.region} | \
            helm registry login --username AWS --password-stdin ${ecrDomain} && \
            helm push ${args.chartName}-${args.chartVersion}.tgz \
            oci://${ecrDomain}/${args.ecrRepositoryName}
        `,
        environment: {
          AWS_REGION: args.region,
        },
      },
      { parent: this, dependsOn: [this.ecrRepository] }
    );

    // Deploy chart if requested
    // Like starting construction after the designs are archived, if requested
    if (args.deploy !== false) {
      // Construct the OCI URL for the chart
      // Like formulating the precise address to retrieve the design from the archives
      const chartUrl = pulumi.interpolate`oci://${ecrDomain}/${args.ecrRepositoryName}/${args.chartName}`;

      // Deploy the chart using the HelmChart component
      // Like starting the actual building construction using the stored designs
      this.helmDeployment = new HelmChart(
        `${name}-deployment`,
        {
          // Namespace where the chart will be deployed
          // The district where the building will be constructed
          namespace: args.namespace,

          // Create the namespace if it doesn't exist
          // Like automatically establishing a new district if needed
          createNamespace: true,

          // The chart to deploy, from the ECR repository
          // The building design to use from the central repository
          chart: chartUrl,

          // The specific version to deploy
          // The exact revision of the building design to use
          version: args.chartVersion,

          // Values to override in the chart - using our merged values
          // Customizations to apply to the standard design
          values: mergedValues,

          // Repository authentication options
          // Credentials needed to access the design repository
          repositoryOpts: {
            repository: pulumi.interpolate`${ecrDomain}/${args.ecrRepositoryName}`,
          },

          // Apply any transformations to resources during deployment
          // Like modifications to the building plans before construction begins
          transformations: args.transforms,

          // Control whether to wait for resources to be ready
          skipAwait: args.skipAwait,

          // Control whether to include test resources
          includeTestHookResources: args.includeTestHookResources,

          // Control whether to skip rendering CRDs
          skipCRDRendering: args.skipCRDRendering,
        },
        { parent: this, dependsOn: [packageAndPushChart] }
      );
    }

    // Register outputs
    // Like filing the final project documentation with the city
    this.registerOutputs({
      ecrRepository: this.ecrRepository,
      helmDeployment: this.helmDeployment,
      repositoryUrl: this.repositoryUrl,
    });
  }

  /**
   * Load chart values from files
   *
   * Like gathering all the building specifications from the available documents
   * before finalizing the construction plans.
   */
  private loadChartValues(chartPath: string, environment?: string): any {
    try {
      let values = {};

      // Check if the chart path exists
      if (!fs.existsSync(chartPath)) {
        console.log(
          `Chart path ${chartPath} does not exist, using default empty values`
        );
        return values;
      }

      // Load base values.yaml from the chart
      const baseValuesPath = path.join(chartPath, "values.yaml");
      if (fs.existsSync(baseValuesPath)) {
        const baseContent = fs.readFileSync(baseValuesPath, "utf8");
        values = yaml.parse(baseContent);
        console.log(`Loaded base values from ${baseValuesPath}`);
      } else {
        console.log(
          `No base values.yaml found at ${baseValuesPath}, using default empty values`
        );
      }

      // Load environment-specific values if environment is provided
      if (environment) {
        const envValuesPath = path.join(
          chartPath,
          `values-${environment}.yaml`
        );
        if (fs.existsSync(envValuesPath)) {
          const envContent = fs.readFileSync(envValuesPath, "utf8");
          const envValues = yaml.parse(envContent);

          // Deep merge the values
          values = this.deepMerge(values, envValues);
          console.log(`Merged environment values from ${envValuesPath}`);
        } else {
          console.log(
            `No environment values file found at ${envValuesPath}, continuing with base values only`
          );
        }
      }

      return values;
    } catch (error) {
      throw new Error(`Error loading chart values from ${chartPath}: ${error}`);
    }
  }

  /**
   * Deep merge two objects
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
   */
  private isObject(item: any): boolean {
    return item && typeof item === "object" && !Array.isArray(item);
  }
}
