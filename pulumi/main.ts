import * as pulumi from "@pulumi/pulumi";
import * as fs from "fs";
import * as yaml from "yaml";
import * as path from "path";

// Import infrastructure stacks
import { NetworkStack } from "@infrastructure/networking";
import { ComputeStack } from "@infrastructure/compute";
import { StorageStack } from "@infrastructure/storage";

// Import services
import { UserService } from "@services/user-service";

// Import configuration and utilities
import {
  getAppConfig,
  getNetworkConfig,
  getEksConfig,
} from "@config/environment";
import { createResourceTags } from "@utils/tagging";

// Import types
import { ChartValues } from "@components/types";
import * as aws from "@pulumi/aws";

/**
 * InfrastructureDeployer - Orchestrates the deployment of cloud infrastructure
 */
export class InfrastructureDeployer {
  private readonly projectName: string;
  private environment: string;
  private readonly tags: Record<string, string>;
  private readonly region: string;
  private readonly config: pulumi.Config;

  // Infrastructure components
  public network?: NetworkStack;
  public compute?: ComputeStack;
  public storage?: StorageStack;
  public userService?: UserService;

  /**
   * Initialize infrastructure deployment with configuration
   */
  constructor(environment?: string) {
    // Load application configuration
    const appConfig = getAppConfig();

    this.projectName = appConfig.projectName;
    this.environment = environment || "dev";

    // Create consistent resource tags across all resources
    this.tags = createResourceTags(this.projectName, this.environment);

    // Set AWS region, falling back to a default
    this.region = process.env.AWS_REGION || "us-west-2";

    // Create Pulumi config reader
    this.config = new pulumi.Config();
  }

  /**
   * Deploy networking infrastructure (VPC, subnets, etc.)
   */
  public deployNetworking(): void {
    const networkConfig = getNetworkConfig();

    const azs = pulumi
      .output(aws.getAvailabilityZones({ state: "available" }))
      .apply((z) => z.names);

    this.network = new NetworkStack("network", {
      cidrBlock: networkConfig.vpcCidrBlock,
      availabilityZones: azs,
      tags: this.tags,
    });

    console.log("Network infrastructure deployed in AZs:", azs);
  }

  /**
   * Deploy compute infrastructure (EKS cluster)
   */
  public deployCompute(): void {
    if (!this.network) {
      throw new Error("Networking infrastructure must be deployed first");
    }

    const eksConfig = getEksConfig();

    // Deploy EKS cluster in the existing network
    this.compute = new ComputeStack("compute", {
      network: this.network,
      kubernetesVersion: eksConfig.version,
      nodeGroups: eksConfig.nodeGroups,
      tags: this.tags,
    });

    console.log("Compute infrastructure deployed");
  }

  /**
   * Deploy storage infrastructure (ECR repositories)
   */
  public deployStorage(): void {
    // Create ECR repositories for container images
    this.storage = new StorageStack("storage", {
      projectName: this.projectName,
      tags: this.tags,
    });

    console.log("Storage infrastructure deployed");
  }

  /**
   * Deploy the user service
   */
  public deployUserService(): void {
    if (!this.compute || !this.storage) {
      throw new Error("Compute and storage must be deployed before services");
    }

    // Get the database password from Pulumi configuration
    // This requires running: pulumi config set --secret Go-Microservices:dbPassword "value"
    const dbPassword = this.config.requireSecret("dbPassword");

    // Load pre-computed chart values
    const userServiceValues = this.loadChartValues("user-service");

    // Deploy user service with dependencies
    this.userService = new UserService("user-service", {
      cluster: this.compute.cluster,
      k8sProvider: this.compute.k8sProvider,
      appRepository: this.storage.appRepository,
      helmRepository: this.storage.helmRepository,
      environment: this.environment,
      dbPassword: dbPassword,
      region: this.region,
      tags: this.tags,
      preloadedValues: userServiceValues, // Pass pre-loaded chart values
    });

    console.log("User service deployed");
  }

  /**
   * Register stack outputs for resource information
   * Now supports partial infrastructure deployment
   */
  public registerOutputs(): void {
    console.log("Registering outputs for deployed components");

    // Create a collection of outputs based on what's deployed
    const outputs: Record<string, any> = {};

    // Network outputs
    if (this.network) {
      outputs.vpcId = this.network.vpcId;
      console.log("Registered network outputs");
    }

    // Compute outputs
    if (this.compute) {
      outputs.kubeconfig = this.compute.kubeconfig;
      outputs.clusterName = this.compute.clusterName;
      console.log("Registered compute outputs");
    }

    // Storage outputs
    if (this.storage) {
      outputs.userServiceImageRepo = this.storage.appRepositoryUrl;
      outputs.helmChartsRepo = this.storage.helmRepositoryUrl;
      console.log("Registered storage outputs");
    }

    // Service outputs
    if (this.userService) {
      outputs.userServiceHttpUrl = this.userService.httpEndpoint;
      outputs.userServiceGrpcUrl = this.userService.grpcEndpoint;
      console.log("Registered user service outputs");
    }

    // Add each output to the exports
    Object.entries(outputs).forEach(([key, value]) => {
      exports[key] = value;
    });

    console.log("All outputs registered successfully");
  }

  /**
   * Load chart values from files
   * @param chartName Name of the chart
   */
  private loadChartValues(chartName: string): ChartValues {
    const chartsDir = path.join(__dirname, "charts");

    // Base values
    let values: ChartValues = {};
    const baseValuesPath = path.join(chartsDir, chartName, "values.yaml");

    try {
      if (fs.existsSync(baseValuesPath)) {
        const baseValues = fs.readFileSync(baseValuesPath, "utf8");
        values = { ...values, ...yaml.parse(baseValues) };
        console.log(
          `Loaded base values for ${chartName} from ${baseValuesPath}`
        );
      } else {
        console.log(`Base values file not found at ${baseValuesPath}`);
      }

      // Environment-specific values
      const envValuesPath = path.join(
        chartsDir,
        chartName,
        `values-${this.environment}.yaml`
      );
      if (fs.existsSync(envValuesPath)) {
        const envValues = fs.readFileSync(envValuesPath, "utf8");
        values = { ...values, ...yaml.parse(envValues) };
        console.log(
          `Loaded ${this.environment} values for ${chartName} from ${envValuesPath}`
        );
      } else {
        console.log(`Environment values file not found at ${envValuesPath}`);
      }

      return values;
    } catch (error) {
      console.error(`Error loading chart values: ${error}`);
      return {}; // Return empty object if values can't be loaded
    }
  }
}

/**
 * Deploy infrastructure in the correct sequence
 */
function deployInfrastructure(): void {
  // Create infrastructure deployer
  const deployer = new InfrastructureDeployer("dev");
  console.log("Starting infrastructure deployment...");

  try {
    // Deploy networking layer
    deployer.deployNetworking();
    // deployer.deployCompute();
    // deployer.deployStorage();
    // deployer.deployUserService();

    // Register outputs for whatever components were deployed
    deployer.registerOutputs();

    console.log("Infrastructure deployment completed successfully");
  } catch (error) {
    console.error(`Error during infrastructure deployment: ${error}`);
    throw error; // Re-throw to ensure Pulumi knows deployment failed
  }
}

// Run the deployment
deployInfrastructure();
