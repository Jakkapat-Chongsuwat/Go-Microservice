import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as random from "@pulumi/random";

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

/**
 * Main Infrastructure Stack
 */
export function deployInfrastructure(): void {
  // Get configuration
  const appConfig = getAppConfig();
  const networkConfig = getNetworkConfig();
  const eksConfig = getEksConfig();

  // Create common resource tags
  const commonTags = createResourceTags(
    appConfig.projectName,
    appConfig.environment
  );

  // Create base name for stack resources
  const stackName = `${appConfig.projectName}-${appConfig.environment}`;

  // Create networking infrastructure
  const network = new NetworkStack("network", {
    cidrBlock: networkConfig.vpcCidrBlock,
    availabilityZones: networkConfig.availabilityZones,
    tags: commonTags,
  });

  // Create compute infrastructure (EKS cluster and node groups)
  const compute = new ComputeStack("compute", {
    network: network,
    kubernetesVersion: eksConfig.version,
    nodeGroups: eksConfig.nodeGroups,
    tags: commonTags,
  });

  // Create storage resources (ECR repositories)
  const storage = new StorageStack("storage", {
    projectName: appConfig.projectName,
    tags: commonTags,
  });

  // Get the current AWS region for use in configurations
  const awsRegion = aws.getRegion().then((region) => region.name);

  // Deploy user service
  const userService = new UserService("user-service", {
    cluster: compute.cluster,
    k8sProvider: compute.k8sProvider,
    appRepository: storage.appRepository,
    helmRepository: storage.helmRepository,
    environment: appConfig.environment,
    dbPassword:
      appConfig.environment === "production"
        ? pulumi.secret(
            new random.RandomPassword("userServicePassword", {
              length: 16,
              special: true,
            }).result
          )
        : pulumi.secret("devPassword"),
    // Pass the region as a string directly without using Output
    region: process.env.AWS_REGION || "us-west-2",
    tags: commonTags,
  });

  // Export stack outputs
  registerOutputs(network, compute, storage, userService);
}

/**
 * Register stack outputs
 */
function registerOutputs(
  network: NetworkStack,
  compute: ComputeStack,
  storage: StorageStack,
  userService: UserService
): void {
  // Export infrastructure information
  exports.vpcId = network.vpcId;
  exports.kubeconfig = compute.kubeconfig;
  exports.clusterName = compute.clusterName;

  // Export repository URLs
  exports.userServiceImageRepo = storage.appRepositoryUrl;
  exports.helmChartsRepo = storage.helmRepositoryUrl;

  // Export service endpoints
  exports.userServiceHttpUrl = userService.httpEndpoint;
  exports.userServiceGrpcUrl = userService.grpcEndpoint;
}

// Deploy infrastructure
deployInfrastructure();
