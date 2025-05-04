// index.ts
import * as pulumi from "@pulumi/pulumi";

// Import infrastructure stacks
import { NetworkStack } from "./infrastructure/networking";
import { ComputeStack } from "./infrastructure/compute";
import { StorageStack } from "./infrastructure/storage";
import { ObservabilityStack } from "./infrastructure/observability";

// Import service deployment
import { UserServiceDeployment } from "./services/user-service";

// Import configuration and utilities
import { getConfig, getEnvironmentConfig } from "./config/environment";
import { createResourceTags } from "./utils/tagging";

/**
 * Main Application Stack
 *
 * This is the main entry point for the infrastructure deployment.
 * It orchestrates the creation of all infrastructure components.
 */
export function deployInfrastructure(): void {
  // Initialize configuration
  const config = getConfig();
  const envConfig = getEnvironmentConfig(pulumi.getStack());
  const baseTags = createResourceTags(config.projectName, pulumi.getStack());

  // Create networking infrastructure
  const network = new NetworkStack("network", {
    cidrBlock: "10.0.0.0/16",
    availabilityZones: ["a", "b", "c"],
    tags: baseTags,
  });

  // Create compute infrastructure (EKS cluster and node groups)
  const compute = new ComputeStack("compute", {
    network: network,
    kubernetesVersion: "1.28",
    nodeGroups: envConfig.nodeGroups,
    tags: baseTags,
  });

  // Create storage resources (ECR repositories)
  const storage = new StorageStack("storage", {
    projectName: config.projectName,
    tags: baseTags,
  });

  // Deploy user service
  const userService = new UserServiceDeployment("user-service", {
    cluster: compute.cluster,
    k8sProvider: compute.k8sProvider,
    appRepository: storage.appRepository,
    helmRepository: storage.helmRepository,
    environment: pulumi.getStack(),
    dbPassword: config.dbPassword,
    tags: baseTags,
  });

  // Set up monitoring and observability
  const observability = new ObservabilityStack("observability", {
    cluster: compute.cluster,
    k8sProvider: compute.k8sProvider,
    grafanaPassword: config.grafanaPassword,
    nodeSelector: {
      "workload-type": "standard",
    },
    tags: baseTags,
  });

  // Export stack outputs
  exportStackOutputs(network, compute, storage, userService, observability);
}

/**
 * Export the stack outputs
 */
function exportStackOutputs(
  network: NetworkStack,
  compute: ComputeStack,
  storage: StorageStack,
  userService: UserServiceDeployment,
  observability: ObservabilityStack
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

  // Export monitoring endpoints
  exports.grafanaEndpoint = observability.grafanaEndpoint;
}

// Deploy the infrastructure
deployInfrastructure();
