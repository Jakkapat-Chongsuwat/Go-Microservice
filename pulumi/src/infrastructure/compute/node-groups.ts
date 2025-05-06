import { NodeGroupConfig } from "../../config/types";

/**
 * Creates standard node groups based on environment
 *
 * @param environment The deployment environment
 * @returns Array of node group configurations
 */
export function createStandardNodeGroups(
  environment: string
): NodeGroupConfig[] {
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
  ];
}

/**
 * Creates a spot instance node group
 *
 * @param environment The deployment environment
 * @returns Spot instance node group configuration
 */
export function createSpotNodeGroup(environment: string): NodeGroupConfig {
  const isProd = environment === "production";

  return {
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
  };
}

/**
 * Creates node groups for the given environment
 *
 * @param environment The deployment environment
 * @returns Array of node group configurations
 */
export function createNodeGroups(environment: string): NodeGroupConfig[] {
  return [
    ...createStandardNodeGroups(environment),
    createSpotNodeGroup(environment),
  ];
}
