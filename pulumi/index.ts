// index.ts
import * as pulumi from "@pulumi/pulumi";
import { Vpc } from "./components/vpc";
import { EksCluster } from "./components/eks-cluster";

// Create a stack-specific configuration
const config = new pulumi.Config();
const projectName = pulumi.getProject();
const stackName = pulumi.getStack();

// Create the VPC
const vpc = new Vpc(`${projectName}-${stackName}`, {
    cidrBlock: "10.0.0.0/16",
    availabilityZones: ["a", "b", "c"], 
    tags: {
        Project: projectName,
        Environment: stackName,
        ManagedBy: "pulumi"
    }
});

// Create the EKS cluster with multiple node groups
const eksCluster = new EksCluster(`${projectName}-${stackName}`, {
    vpcId: vpc.vpc.id,
    subnetIds: vpc.privateSubnetIds, 
    version: "1.28",
    nodeGroups: [
        // General purpose nodes
        {
            name: "standard",
            instanceTypes: ["t3.medium"],
            desiredSize: 2,
            minSize: 1,
            maxSize: 5,
            labels: {
                "workload-type": "standard"
            }
        },
        // Compute-optimized nodes for CPU-intensive workloads
        {
            name: "compute",
            instanceTypes: ["c5.xlarge"],
            desiredSize: 2,
            minSize: 1,
            maxSize: 10,
            labels: {
                "workload-type": "cpu-intensive"
            }
        },
        // Spot instances for cost optimization
        {
            name: "spot",
            instanceTypes: ["t3.large", "t3.xlarge"],
            desiredSize: 1,
            minSize: 0,
            maxSize: 10,
            capacityType: "SPOT",
            labels: {
                "workload-type": "spot"
            },
            taints: [
                {
                    key: "spot",
                    value: "true",
                    effect: "NO_SCHEDULE"
                }
            ]
        }
    ],
    tags: {
        Project: projectName,
        Environment: stackName,
        ManagedBy: "pulumi"
    }
});

// Export values
export const vpcId = vpc.vpc.id;
export const kubeconfig = eksCluster.kubeconfig;
export const clusterName = eksCluster.cluster.eksCluster.name;
