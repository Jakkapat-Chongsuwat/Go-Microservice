import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as k8s from "@pulumi/kubernetes";
import { EksCluster } from "../../components/eks-cluster";
import { NodeGroupConfig } from "../../config/types";
import { createResourceTags } from "../../utils/tagging";
import { createComponentName } from "../../utils/naming";

export interface EksClusterArgs {
  /**
   * VPC ID to deploy the cluster into
   */
  vpcId: pulumi.Input<string>;

  /**
   * Subnet IDs for the cluster (should be private subnets)
   */
  subnetIds: pulumi.Input<pulumi.Input<string>[]>;

  /**
   * Kubernetes version to use
   */
  version: string;

  /**
   * Node groups to create
   */
  nodeGroups: NodeGroupConfig[];

  /**
   * Tags to apply to resources
   */
  tags?: Record<string, string>;
}

/**
 * EksClusterResource creates an EKS cluster and related resources
 */
export class EksClusterResource {
  /**
   * The EKS cluster component
   */
  public readonly cluster: EksCluster;

  /**
   * The Kubernetes provider for this cluster
   */
  public readonly k8sProvider: k8s.Provider;

  /**
   * The kubeconfig for the cluster
   */
  public readonly kubeconfig: pulumi.Output<string>;

  /**
   * The name of the cluster
   */
  public readonly clusterName: pulumi.Output<string>;

  /**
   * The ARN of the IAM role for the AWS Load Balancer Controller
   */
  public readonly albControllerRoleArn: pulumi.Output<string>;

  /**
   * Creates a new EKS cluster and related resources
   *
   * @param name Base name for resources
   * @param args EKS cluster configuration
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    // Create the EKS cluster
    this.cluster = new EksCluster(
      name,
      {
        vpcId: args.vpcId,
        subnetIds: args.subnetIds,
        version: args.version,
        nodeGroups: args.nodeGroups,
        tags: args.tags,
      },
      opts
    );

    // Create a Kubernetes provider for this cluster
    this.k8sProvider = new k8s.Provider(
      `${name}-k8s-provider`,
      {
        kubeconfig: this.cluster.kubeconfig,
      },
      { dependsOn: [this.cluster] }
    );

    // Export the cluster resources
    this.kubeconfig = this.cluster.kubeconfig;
    this.clusterName = this.cluster.cluster.eksCluster.name;
    this.albControllerRoleArn = this.cluster.albControllerRole.arn;
  }

  /**
   * Installs the AWS Load Balancer Controller into the cluster
   *
   * @param name Base name for resources
   * @param vpcId VPC ID for the controller
   * @returns The Helm chart for the controller
   */
  public installLoadBalancerController(
    name: string,
    vpcId: pulumi.Input<string>
  ): k8s.helm.v3.Chart {
    // Get the AWS region
    const region = pulumi.output(aws.getRegion()).name;

    // Install the AWS Load Balancer Controller
    return new k8s.helm.v3.Chart(
      `${name}-alb-controller`,
      {
        namespace: "kube-system",
        chart: "aws-load-balancer-controller",
        version: "1.4.6",
        fetchOpts: {
          repo: "https://aws.github.io/eks-charts",
        },
        values: {
          clusterName: this.clusterName,
          serviceAccount: {
            create: true,
            annotations: {
              "eks.amazonaws.com/role-arn": this.albControllerRoleArn,
            },
          },
          region: region,
          vpcId: vpcId,
        },
      },
      {
        provider: this.k8sProvider,
        dependsOn: [this.cluster],
      }
    );
  }
}
