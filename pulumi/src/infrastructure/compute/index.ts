// pulumi/src/infrastructure/compute/index.ts

import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import { EksClusterResource } from "./eks";
import { createNodeGroups } from "./node-groups";
import { NetworkStack } from "../networking";
import { createComponentName } from "../../utils/naming";

export interface ComputeStackArgs {
  /**
   * The network stack to deploy into
   */
  network: NetworkStack;

  /**
   * Kubernetes version to use
   */
  kubernetesVersion: string;

  /**
   * Node group configurations
   */
  nodeGroups?: any[];

  /**
   * Tags to apply to resources
   */
  tags?: Record<string, string>;
}

/**
 * ComputeStack creates compute resources
 */
export class ComputeStack {
  /**
   * The EKS cluster
   */
  public readonly cluster: EksClusterResource;

  /**
   * The Kubernetes provider
   */
  public readonly k8sProvider: k8s.Provider;

  /**
   * The Kubeconfig for accessing the cluster
   */
  public readonly kubeconfig: pulumi.Output<string>;

  /**
   * The cluster name
   */
  public readonly clusterName: pulumi.Output<string>;

  /**
   * The ALB controller chart
   */
  public readonly albController: k8s.helm.v3.Chart;

  /**
   * Creates a new compute stack
   *
   * @param name Base name for resources
   * @param args Compute stack configuration
   * @param opts Additional resource options
   */
  constructor(
    name: string,
    args: ComputeStackArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    const nodeGroups = args.nodeGroups || createNodeGroups(pulumi.getStack());

    this.cluster = new EksClusterResource(
      createComponentName(name, "eks"),
      {
        vpcId: args.network.vpcId,
        subnetIds: args.network.privateSubnetIds,
        version: args.kubernetesVersion,
        nodeGroups: nodeGroups,
        tags: args.tags,
      },
      opts
    );

    this.k8sProvider = this.cluster.k8sProvider;

    this.albController = this.cluster.installLoadBalancerController(
      createComponentName(name, "eks"),
      args.network.vpcId
    );

    this.kubeconfig = this.cluster.kubeconfig;
    this.clusterName = this.cluster.clusterName;
  }
}

export { EksClusterResource };
