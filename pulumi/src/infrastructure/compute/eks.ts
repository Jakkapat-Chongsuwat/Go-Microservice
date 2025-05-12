/* eksClusterResource.ts */
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as k8s from "@pulumi/kubernetes";
import { EksCluster } from "@components/eks-cluster";
import { NodeGroupConfig } from "@config/types";
import * as childProcess from "child_process";
import * as yaml from "js-yaml";

export interface EksClusterArgs {
  vpcId: pulumi.Input<string>;
  subnetIds: pulumi.Input<pulumi.Input<string>[]>;
  version: string;
  nodeGroups: NodeGroupConfig[];
  tags?: Record<string, string>;
}

/**
 * Sets the AWS CLI output format to JSON.
 */
function setAwsCliOutputFormat(): void {
  try {
    process.env.AWS_DEFAULT_OUTPUT = "json";
    childProcess.execSync("aws configure set default.output json", {
      stdio: "inherit",
    });
    console.log("Set AWS CLI output format to JSON");
  } catch (error) {
    console.warn("Failed to set AWS CLI output format:", error);
  }
}

/**
 * Transforms a kubeconfig string that might be a multi-document YAML into a single document.
 */
function transformKubeconfig(kubeconfig: any): string {
  try {
    if (typeof kubeconfig !== "string") {
      kubeconfig = JSON.stringify(kubeconfig, null, 2);
    }
    const trimmed = kubeconfig.trim();
    const docs = yaml.loadAll(trimmed);
    if (Array.isArray(docs) && docs.length > 0) {
      console.log("Kubeconfig transformed to a single document.");
      return yaml.dump(docs[0]);
    }
    return kubeconfig;
  } catch (error) {
    console.error("Error transforming kubeconfig:", error);
    return kubeconfig;
  }
}

export class EksClusterResource {
  public readonly cluster: EksCluster;
  public readonly k8sProvider: k8s.Provider;
  public readonly kubeconfig: pulumi.Output<string>;
  public readonly clusterName: pulumi.Output<string>;
  public readonly albControllerRoleArn: pulumi.Output<string>;

  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    try {
      console.log("Starting creation of EKS cluster component");

      // Set AWS CLI configuration to JSON format.
      setAwsCliOutputFormat();

      // Create the underlying EKS cluster
      this.cluster = new EksCluster(
        name,
        {
          vpcId: args.vpcId,
          subnetIds: args.subnetIds,
          version: args.version || "1.28",
          nodeGroups: args.nodeGroups || [],
          tags: args.tags || {},
        },
        opts
      );
      console.log("EKS cluster component created.");

      // Export the cluster name and ALB controller role ARN
      this.clusterName = this.cluster.cluster.eksCluster.name;
      this.albControllerRoleArn = this.cluster.albControllerRole.arn;

      // Transform the kubeconfig (in case it is multi-document YAML)
      this.kubeconfig = this.cluster.kubeconfig.apply((configStr: string) => {
        return transformKubeconfig(configStr);
      });

      // Create the Kubernetes provider using the transformed kubeconfig.
      console.log("Creating Kubernetes provider");
      this.k8sProvider = new k8s.Provider(
        `${name}-k8s-provider`,
        { kubeconfig: this.kubeconfig },
        { dependsOn: [this.cluster] }
      );
      console.log("Kubernetes provider created.");
      console.log("EksClusterResource constructor completed.");
    } catch (err) {
      console.error("Error in EksClusterResource constructor:", err);
      throw err;
    }
  }

  /**
   * Installs the AWS Load Balancer Controller using a Helm chart.
   */
  public installLoadBalancerController(
    name: string,
    vpcId: pulumi.Input<string>
  ): k8s.helm.v3.Chart {
    try {
      console.log("Installing load balancer controller...");

      const region = pulumi.output(aws.getRegion()).name;

      const chart = new k8s.helm.v3.Chart(
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

      console.log("ALB controller installed successfully.");
      return chart;
    } catch (err) {
      console.error("Failed to install load balancer controller:", err);
      throw err;
    }
  }
}
