import * as pulumi from "@pulumi/pulumi";
import {
  AppConfig,
  NetworkConfig,
  EksConfig,
  ServiceConfig,
  NodeGroupConfig,
  DatabaseConfig,
} from "./types";

/**
 * Get the base application configuration
 *
 * @returns Base application configuration
 */
export function getAppConfig(): AppConfig {
  const stack = pulumi.getStack();

  return {
    projectName: pulumi.getProject(),
    environment: stack,
    region: getAWSRegion(),
  };
}

/**
 * Get database configuration
 *
 * @returns Database configuration
 */
export function getDatabaseConfig(): DatabaseConfig {
  const config = new pulumi.Config();

  return {
    dbPassword: config.requireSecret("dbPassword"),
    dbUsername: "devuser", // In production, use secrets
    dbName: "user_service",
    dbSchema: "user_service",
    dbPort: 5555,
  };
}

/**
 * Get monitoring configuration
 *
 * @returns Grafana password
 */
export function getMonitoringConfig(): {
  grafanaPassword: pulumi.Output<string>;
} {
  const config = new pulumi.Config();

  return {
    grafanaPassword: config.requireSecret("grafanaPassword"),
  };
}

/**
 * Get AWS region from config or use default
 *
 * @returns AWS region
 */
function getAWSRegion(): string {
  const config = new pulumi.Config("aws");
  return config.get("region") || "us-west-2";
}

/**
 * Get environment-specific network configuration
 *
 * @returns Network configuration
 */
export function getNetworkConfig(): NetworkConfig {
  return {
    vpcCidrBlock: "10.0.0.0/16",
    availabilityZones: ["a", "b", "c"],
    createPrivateSubnets: true,
    createPublicSubnets: true,
  };
}

/**
 * Get environment-specific node groups configuration
 *
 * @returns Array of node group configurations
 */
export function getNodeGroups(): NodeGroupConfig[] {
  const environment = pulumi.getStack();
  const isProd = environment === "production";

  return [
    // General purpose nodes
    {
      name: "standard",
      instanceTypes: ["t3.medium"],
      desiredSize: isProd ? 3 : 2,
      minSize: isProd ? 2 : 1,
      maxSize: isProd ? 6 : 4,
      labels: {
        "workload-type": "standard",
      },
    },

    // Compute-optimized nodes for CPU-intensive workloads
    {
      name: "compute",
      instanceTypes: ["c5.xlarge"],
      desiredSize: isProd ? 2 : 1,
      minSize: isProd ? 2 : 0,
      maxSize: isProd ? 8 : 4,
      labels: {
        "workload-type": "cpu-intensive",
      },
    },

    // Spot instances for cost optimization
    {
      name: "spot",
      instanceTypes: ["t3.large", "t3.xlarge"],
      desiredSize: 1,
      minSize: 0,
      maxSize: isProd ? 10 : 5,
      capacityType: "SPOT",
      labels: {
        "workload-type": "spot",
      },
      taints: [
        {
          key: "spot",
          value: "true",
          effect: "NO_SCHEDULE",
        },
      ],
    },
  ];
}

/**
 * Get EKS cluster configuration
 *
 * @returns EKS configuration including node groups
 */
export function getEksConfig(): EksConfig {
  return {
    version: "1.28",
    nodeGroups: getNodeGroups(),
  };
}

/**
 * Get service configuration for user-service
 *
 * @returns User service configuration
 */
export function getUserServiceConfig(): ServiceConfig {
  const environment = pulumi.getStack();
  const isProd = environment === "production";

  return {
    replicaCount: isProd ? 3 : 1,
    resources: isProd
      ? {
          limits: {
            cpu: "1",
            memory: "1Gi",
          },
          requests: {
            cpu: "0.5",
            memory: "512Mi",
          },
        }
      : {
          limits: {
            cpu: "0.3",
            memory: "512Mi",
          },
          requests: {
            cpu: "0.1",
            memory: "128Mi",
          },
        },
    autoscaling: {
      enabled: isProd,
      minReplicas: isProd ? 3 : 1,
      maxReplicas: isProd ? 10 : 5,
      targetCpuUtilization: 70,
      targetMemoryUtilization: 80,
    },
    hostname: isProd ? "api.example.com" : `${environment}-api.example.com`,
  };
}
