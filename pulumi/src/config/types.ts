import * as pulumi from "@pulumi/pulumi";

/**
 * Type definitions for configuration
 */

/**
 * Base application configuration
 */
export interface AppConfig {
  /**
   * Project name used for resource naming and tagging
   */
  projectName: string;

  /**
   * AWS region to deploy resources
   */
  region: string;
}

/**
 * Database credentials configuration
 */
export interface DatabaseConfig {
  /**
   * Database password (secret)
   */
  dbPassword: pulumi.Output<string>;

  /**
   * Database username
   */
  dbUsername: string;

  /**
   * Database name
   */
  dbName: string;

  /**
   * Database schema
   */
  dbSchema: string;

  /**
   * Database port
   */
  dbPort: number;
}

/**
 * Network configuration
 */
export interface NetworkConfig {
  vpcCidrBlock: string;
  availabilityZones: pulumi.Input<pulumi.Input<string>[]>;
  createPrivateSubnets: boolean;
  createPublicSubnets: boolean;
}

/**
 * Node group configuration for EKS
 */
export interface NodeGroupConfig {
  /**
   * Node group name
   */
  name: string;

  /**
   * Instance types to use
   */
  instanceTypes: string[];

  /**
   * Desired number of nodes
   */
  desiredSize: number;

  /**
   * Minimum number of nodes
   */
  minSize: number;

  /**
   * Maximum number of nodes
   */
  maxSize: number;

  /**
   * Capacity type (ON_DEMAND or SPOT)
   */
  capacityType?: "ON_DEMAND" | "SPOT";

  /**
   * Node labels to apply
   */
  labels?: Record<string, string>;

  /**
   * Node taints to apply
   */
  taints?: {
    key: string;
    value: string;
    effect: "NO_SCHEDULE" | "PREFER_NO_SCHEDULE" | "NO_EXECUTE";
  }[];

  /**
   * Disk size in GB
   */
  diskSize?: number;
}

/**
 * EKS cluster configuration
 */
export interface EksConfig {
  /**
   * Kubernetes version
   */
  version: string;

  /**
   * Node groups configuration
   */
  nodeGroups: NodeGroupConfig[];
}

/**
 * Application resources configuration
 */
export interface ResourceConfig {
  /**
   * Compute limits
   */
  limits: {
    cpu: string;
    memory: string;
  };

  /**
   * Compute requests
   */
  requests: {
    cpu: string;
    memory: string;
  };
}

/**
 * Application autoscaling configuration
 */
export interface AutoscalingConfig {
  /**
   * Whether autoscaling is enabled
   */
  enabled: boolean;

  /**
   * Minimum replicas
   */
  minReplicas: number;

  /**
   * Maximum replicas
   */
  maxReplicas: number;

  /**
   * Target CPU utilization percentage
   */
  targetCpuUtilization: number;

  /**
   * Target memory utilization percentage
   */
  targetMemoryUtilization: number;
}

/**
 * Service configuration
 */
export interface ServiceConfig {
  /**
   * Number of replicas
   */
  replicaCount: number;

  /**
   * Resource configuration
   */
  resources: ResourceConfig;

  /**
   * Autoscaling configuration
   */
  autoscaling: AutoscalingConfig;

  /**
   * Service hostname
   */
  hostname: string;
}
