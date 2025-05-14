// pulumi/src/infrastructure/compute/eks.ts

import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as k8s from "@pulumi/kubernetes";
import { EksCluster } from "@components/eks-cluster";
import { NodeGroupConfig } from "@config/types";
import * as childProcess from "child_process";
import * as yaml from "js-yaml";
import * as eks from "@pulumi/eks";

export interface EksClusterArgs {
  vpcId: pulumi.Input<string>;
  subnetIds: pulumi.Input<pulumi.Input<string>[]>;
  version: string;
  nodeGroups: NodeGroupConfig[];
  tags?: Record<string, string>;
}

/**
 * Sets the AWS CLI output format to JSON
 */
function setAwsCliOutputFormat(): void {
  try {
    process.env.AWS_DEFAULT_OUTPUT = "json";
    childProcess.execSync("aws configure set default.output json", {
      stdio: "inherit",
    });
    console.log("AWS CLI output format set to JSON");
  } catch (error) {
    console.warn("Failed to set AWS CLI format:", error);
  }
}

/**
 * Transforms a multi-document YAML kubeconfig into a single document
 */
function transformKubeconfig(kubeconfig: string): string {
  try {
    const docs = yaml.loadAll(kubeconfig.trim());
    return Array.isArray(docs) && docs.length > 0
      ? yaml.dump(docs[0])
      : kubeconfig;
  } catch (error) {
    console.error("Kubeconfig transformation error:", error);
    return kubeconfig;
  }
}

export class EksClusterResource {
  public readonly cluster: EksCluster;
  public readonly k8sProvider: k8s.Provider;
  public readonly kubeconfig: pulumi.Output<string>;
  public readonly clusterName: pulumi.Output<string>;
  public readonly albControllerRoleArn: pulumi.Output<string>;
  public readonly nodeGroups: eks.ManagedNodeGroup[];

  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    setAwsCliOutputFormat();

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

    this.clusterName = this.cluster.cluster.eksCluster.name;
    this.albControllerRoleArn = this.cluster.albControllerRole.arn;
    this.nodeGroups = this.cluster.nodeGroups;

    this.kubeconfig = this.cluster.kubeconfig.apply(transformKubeconfig);

    this.k8sProvider = new k8s.Provider(
      `${name}-k8s`,
      { kubeconfig: this.kubeconfig },
      { dependsOn: [this.cluster.cluster, ...this.nodeGroups] }
    );
  }

  /**
   * Installs the AWS Load Balancer Controller using a Helm chart
   */
  public installLoadBalancerController(
    name: string,
    vpcId: pulumi.Input<string>
  ): k8s.helm.v3.Chart {
    const region = aws.getRegionOutput().name;

    return new k8s.helm.v3.Chart(
      `${name}-alb-controller`,
      {
        namespace: "kube-system",
        chart: "aws-load-balancer-controller",
        version: "1.6.1",
        fetchOpts: { repo: "https://aws.github.io/eks-charts" },
        values: {
          clusterName: this.clusterName,
          serviceAccount: {
            name: "aws-load-balancer-controller",
            annotations: {
              "eks.amazonaws.com/role-arn": this.albControllerRoleArn,
            },
          },
          region,
          vpcId,
          resources: {
            limits: { cpu: "500m", memory: "1024Mi" },
            requests: { cpu: "200m", memory: "512Mi" },
          },
          podAnnotations: {
            "cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
          },
          tolerations: [
            {
              key: "node.kubernetes.io/not-ready",
              operator: "Exists",
              effect: "NoExecute",
              tolerationSeconds: 300,
            },
            // Allow running on spot instances
            {
              key: "spot",
              operator: "Equal",
              value: "true",
              effect: "NoSchedule",
            },
          ],
          enableCertManager: false,
          ingressClassResource: {
            default: true,
          },
          // Set webhook timeout to avoid controller issues during cluster updates
          webhook: {
            timeoutSeconds: 30,
          },
        },
      },
      {
        provider: this.k8sProvider,
        dependsOn: [
          this.cluster.cluster,
          this.cluster.albControllerRole,
          ...this.nodeGroups,
        ],
      }
    );
  }
}
