import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as eks from "@pulumi/eks";

/**
 * Configuration for an EKS node group.
 *
 * Just as a city requires different types of workers - police, firefighters,
 * maintenance crews - a Kubernetes cluster needs different types of nodes
 * to handle various workload requirements.
 */
export interface NodeGroupConfig {
  /**
   * The name of the node group.
   *
   * Like a department name in city government, this helps identify
   * the purpose of this specific group of workers. Without clear naming,
   * you'd struggle to know which node group serves which purpose, similar
   * to a city with unnamed departments where citizens don't know where to go.
   */
  name: string;

  /**
   * Types of instances to use for this node group.
   *
   * Similar to specifying what skills or equipment city workers need -
   * some jobs need heavy machinery, others need specialized tools.
   *
   * For example, memory-intensive applications (like databases) need instances
   * with more RAM (r5, r6 types), just as the city water department needs
   * specialized equipment different from what the parks department uses.
   *
   * Not specifying appropriate instance types is like asking firefighters
   * to fight fires without proper equipment - your workloads won't have the
   * resources they need to perform effectively.
   */
  instanceTypes: string[];

  /**
   * The number of nodes the group should maintain under normal conditions.
   *
   * Like staffing a department with the right number of employees to
   * handle the typical daily workload of a city. Too few workers means
   * services get backlogged; too many means you're paying workers to stand idle.
   *
   * Setting this too low risks application performance issues during normal operation,
   * while setting it too high unnecessarily increases costs when that capacity
   * isn't needed.
   */
  desiredSize: number;

  /**
   * Minimum number of nodes to keep running at all times.
   *
   * Similar to the minimum staffing requirement to keep essential
   * city services operational, even during off-hours. Just as a fire station
   * must always have some firefighters on duty, your cluster must maintain
   * enough nodes to handle critical workloads even during low-demand periods.
   *
   * If set too low, your critical applications might lack resources during
   * scale-down events. If set too high, you'll waste money on idle resources.
   */
  minSize: number;

  /**
   * Maximum number of nodes the group can scale up to.
   *
   * Like the maximum number of workers a city can call in during
   * a major event or emergency situation. When demand spikes - like
   * a festival bringing tourists to the city - you need a plan for
   * how many additional workers you can deploy.
   *
   * Setting this too low limits your ability to handle traffic spikes,
   * potentially causing application outages. Setting it too high could
   * allow runaway scaling that creates unexpectedly large bills if an
   * autoscaling event goes unchecked.
   */
  maxSize: number;

  /**
   * Size of the root disk in GB.
   *
   * Similar to how much equipment storage each worker gets. Just as
   * city workers need adequate storage space for their tools and materials,
   * nodes need enough disk space for the OS, application data, logs, and
   * container images.
   *
   * The default is 20GB, which is sufficient for basic workloads but may
   * need to be increased for disk-intensive applications or if you're storing
   * many large container images locally.
   *
   * If not set appropriately, nodes can run out of disk space, causing
   * applications to fail with "No space left on device" errors, similar to
   * city crews unable to accept new materials because their storage is full.
   */
  diskSize?: number;

  /**
   * AMI type for the node instances.
   *
   * Like choosing the standard-issue equipment and training package
   * for a specific department of workers. Just as police officers need
   * different equipment than crossing guards, some workloads require
   * specialized operating system configurations.
   *
   * Options include:
   * - AL2_x86_64 (Amazon Linux 2) - Standard general purpose OS
   * - AL2_x86_64_GPU - For GPU workloads, like machine learning
   * - AL2_ARM_64 - For ARM-based instances (graviton), offering better price/performance
   *
   * Using the wrong AMI type is like sending road workers to fix electrical wires -
   * they won't have the right tools or training for the job, causing compatibility
   * issues or preventing specialized hardware from being utilized properly.
   */
  amiType?: string;

  /**
   * Whether to use on-demand or spot instances.
   *
   * Similar to hiring full-time employees (on-demand) vs. temporary
   * workers (spot) who cost less but might leave on short notice.
   *
   * ON_DEMAND: Like permanent city employees with guaranteed shifts.
   * They cost more but provide stability and reliability.
   *
   * SPOT: Like seasonal workers hired at lower rates during busy periods.
   * They cost 70-90% less but can be terminated with only 2 minutes notice
   * if AWS needs the capacity back.
   *
   * If not specified, defaults to ON_DEMAND for reliability. Using SPOT for
   * critical workloads without proper fault tolerance is like staffing your
   * emergency services with only temporary workers - they might all leave
   * during a crisis.
   */
  capacityType?: "ON_DEMAND" | "SPOT";

  /**
   * Kubernetes labels to apply to nodes.
   *
   * Like ID badges that identify which department and skills
   * each worker has, helping assign the right jobs to them.
   *
   * For example, a label like {"gpu": "true"} ensures GPU workloads
   * go to nodes with GPU hardware, just as "water department" badges
   * ensure water specialists get assigned to plumbing projects.
   *
   * Without proper labels, workloads might run on inappropriate nodes,
   * like assigning park maintenance to traffic controllers because
   * there was no way to distinguish their skills.
   */
  labels?: { [key: string]: string };

  /**
   * Taints to apply to nodes, preventing certain workloads from running on them.
   *
   * Similar to work restrictions for specialized city employees -
   * "This crew only handles electrical work, not plumbing."
   *
   * Taints actively repel certain workloads unless they have matching tolerations.
   * For example, a taint of {"dedicated": "gpu", effect: "NoSchedule"} ensures
   * only workloads that specifically tolerate this taint will run on these nodes.
   *
   * Effects determine how strict the restriction is:
   * - NO_SCHEDULE: Firm barrier - like "Authorized Personnel Only" areas
   * - PREFER_NO_SCHEDULE: Soft barrier - "Please avoid unless necessary"
   * - NO_EXECUTE: Eviction barrier - "Leave immediately if you're not authorized"
   *
   * Without taints, specialized expensive nodes might run general workloads,
   * wasting their specialized capabilities - like having brain surgeons
   * perform routine checkups instead of surgeries.
   */
  taints?: {
    key: string;
    value: string;
    effect: "NO_SCHEDULE" | "PREFER_NO_SCHEDULE" | "NO_EXECUTE";
  }[];

  /**
   * Maximum number of unavailable nodes during update operations.
   *
   * Like the maximum number of workers that can be in training
   * at any one time without disrupting city operations.
   *
   * When AWS updates the node group (e.g., for security patches or AMI updates),
   * this controls how many nodes can be unavailable simultaneously.
   *
   * Setting this too high (e.g., 5 in a 10-node group) could cause 50% capacity
   * reduction during updates, potentially disrupting services - like putting
   * half your police force in training during a major event. Setting it too low
   * (e.g., 1 in a 100-node cluster) makes updates take an extremely long time.
   *
   * Default is 1, which is conservative but safe. Consider setting to a percentage
   * of your group size (like 10-20%) for large node groups.
   */
  maxUnavailable?: number;
}

/**
 * Arguments for creating an EKS cluster.
 *
 * Like the master blueprint for developing a new self-contained city district,
 * these parameters define the foundation and structure of your Kubernetes environment.
 */
export interface EksClusterArgs {
  /**
   * The VPC ID where the cluster will be created.
   *
   * Similar to the designated land area where a new city district will be built.
   * The VPC defines the network boundaries and isolation for your cluster,
   * much like city limits establish where city services and governance apply.
   *
   * Not specifying a VPC ID will result in using the default VPC, which might
   * lack proper security controls or network design - like building a new
   * neighborhood without proper planning for utilities and roads.
   */
  vpcId: pulumi.Input<string>;

  /**
   * Subnet IDs where cluster resources will be deployed.
   *
   * Like the different neighborhoods and zones within the city plan,
   * spread across various geographical areas for resilience.
   *
   * Subnets should span multiple Availability Zones for high availability,
   * similar to how a city builds firehouses in different districts to
   * ensure coverage even if one area is inaccessible.
   *
   * Not specifying appropriate subnets can result in:
   * - Single point of failure if only using one AZ
   * - Network conflicts if subnets don't have enough IP addresses
   * - Connectivity issues if network ACLs or route tables are misconfigured
   */
  subnetIds: pulumi.Input<pulumi.Input<string>[]>;

  /**
   * Kubernetes version to use.
   *
   * Similar to which building code version the city will follow -
   * newer versions have improved standards and features.
   *
   * Kubernetes versions introduce new features, bug fixes, and security patches.
   * Choosing too old a version might miss critical security fixes, while
   * choosing the very latest might expose you to newly introduced bugs.
   *
   * If not specified, defaults to a recent, stable version (1.28), which
   * balances new features with stability.
   *
   * Not specifying this is like letting a contractor choose which building
   * code to follow - you might get modern standards or outdated practices.
   */
  version?: string;

  /**
   * Configuration for node groups within the cluster.
   *
   * Like planning different work crews and departments that will
   * keep the city operational and serve different functions.
   *
   * A well-designed cluster typically needs multiple node groups to:
   * - Separate workloads with different resource requirements
   * - Isolate critical applications from less important ones
   * - Use cost-optimized instance types for different purposes
   *
   * Not specifying node groups would rely on the default node group, which
   * is like staffing your entire city with generic workers - they can do
   * basic tasks but aren't specialized for particular roles.
   */
  nodeGroups?: NodeGroupConfig[];

  /**
   * Tags to apply to cluster resources.
   *
   * Similar to official markers and identification placed on
   * city infrastructure for management and categorization.
   *
   * Tags allow for:
   * - Cost allocation (which team/project is paying for what)
   * - Automation (target resources based on tags)
   * - Governance (identify purpose, ownership, security requirements)
   *
   * Not using tags is like having unmarked city property - you can't
   * easily see which department owns what, what budget it belongs to,
   * or what safety standards apply to it.
   */
  tags?: { [key: string]: string };
}

/**
 * EksCluster represents a fully configured Amazon EKS cluster with required supporting resources.
 *
 * Following the Single Responsibility Principle, this class focuses solely on creating and
 * managing an EKS cluster and its directly associated resources.
 *
 * Like a city planning department that handles the overall design and infrastructure
 * of the city but delegates specific implementation details to specialized departments.
 */
export class EksCluster extends pulumi.ComponentResource {
  /**
   * The underlying EKS cluster resource.
   *
   * This is the central command structure of our city - the main administrative
   * center that coordinates all the workers and services.
   *
   * The cluster contains the Kubernetes control plane, which manages scheduling,
   * scaling, updates, and the overall state of your applications, similar to how
   * city hall contains the mayor, council, and departments that govern city operations.
   */
  public readonly cluster: eks.Cluster;

  /**
   * The kubeconfig for accessing the cluster.
   *
   * Like the master key and security credentials that grant administrative
   * access to all city systems and buildings.
   *
   * This configuration file contains the authentication details needed to
   * connect to your cluster. It's equivalent to having the security codes,
   * keys, and authorization papers to enter secure government facilities.
   *
   * Protecting this kubeconfig is crucial - if it falls into unauthorized hands,
   * they gain administrator access to your entire cluster.
   */
  public readonly kubeconfig: pulumi.Output<string>;

  /**
   * IAM role for the AWS Load Balancer Controller.
   *
   * Similar to the official authorization for the city's transportation department
   * to manage entry points and traffic flow across city boundaries.
   *
   * This IAM role grants the AWS Load Balancer Controller permission to create
   * and manage AWS resources like load balancers when Kubernetes services request them.
   *
   * Without this role, your cluster would be unable to create external load balancers,
   * similar to a city with well-developed internal roads but no highways connecting
   * to neighboring areas.
   */
  public readonly albControllerRole: aws.iam.Role;

  constructor(
    name: string,
    args: EksClusterArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("custom:resource:EksCluster", name, {}, opts);

    // Create the EKS cluster - the foundation of our Kubernetes city
    this.cluster = this.createCluster(name, args);

    // Create worker node groups - the various departments of city workers
    if (args.nodeGroups && args.nodeGroups.length > 0) {
      this.createNodeGroups(name, args.nodeGroups, args.tags);
    }

    // Create IAM role for ALB controller - official permissions for traffic management
    this.albControllerRole = this.createAlbControllerRole(name);

    // Properly assign the kubeconfig - the master access credentials
    this.kubeconfig = pulumi.interpolate`${this.cluster.kubeconfig}`;

    // Register outputs for Pulumi state management - file the official city records
    this.registerOutputs({
      kubeconfig: this.kubeconfig,
      albControllerRole: this.albControllerRole,
      clusterName: this.cluster.eksCluster.name,
    });
  }

  /**
   * Creates the main EKS cluster.
   *
   * Following the principle of separation of concerns, this method handles only
   * the cluster creation logic, making it easier to test and maintain.
   *
   * This is like the initial phase of city development where the basic infrastructure
   * is established - roads, utilities, and government buildings - before individual
   * neighborhoods and services are developed.
   *
   * @param name The name of the cluster
   * @param args The cluster configuration arguments
   * @returns The created EKS cluster
   */
  private createCluster(name: string, args: EksClusterArgs): eks.Cluster {
    // Create the core city infrastructure - the central administration and services
    return new eks.Cluster(
      name,
      {
        vpcId: args.vpcId,
        subnetIds: args.subnetIds,
        version: args.version || "1.28", // Default to a recent, stable version if not specified
        instanceType: "t3.medium", // Default worker type for initial setup
        desiredCapacity: 2, // Start with a minimal but resilient worker count
        minSize: 1, // Minimum to keep services running
        maxSize: 4, // Maximum to handle increased demand
        nodePublicKey:
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com",
        tags: {
          ...(args.tags || {}),
          "kubernetes.io/cluster/name": name, // Required tag for proper cluster discovery
        },
        createOidcProvider: true, // Enable IAM roles for service accounts
        skipDefaultNodeGroup: true, // We'll create our own specialized node groups
      },
      { parent: this }
    );
  }

  /**
   * Creates node groups based on the provided configurations.
   *
   * Following the Interface Segregation Principle, this method focuses only
   * on node group creation without handling unrelated concerns.
   *
   * This is similar to establishing different departments in a city government -
   * each with specialized workers, equipment, and responsibilities. Just as a city
   * needs police, firefighters, maintenance crews, and administrative staff, a
   * Kubernetes cluster often needs specialized nodes for different workloads.
   *
   * @param clusterName The name of the cluster
   * @param nodeGroups Array of node group configurations
   * @param tags Optional tags to apply to node groups
   */
  private createNodeGroups(
    clusterName: string,
    nodeGroups: NodeGroupConfig[],
    tags?: { [key: string]: string }
  ): void {
    // Create the specialized worker departments for our Kubernetes city
    nodeGroups.forEach((ng) => {
      const nodeGroupName = `${clusterName}-${ng.name}`;

      // Format taints for the AWS API - converting worker restrictions to the required format
      const formattedTaints = this.formatTaints(ng.taints);

      // Prepare update configuration for controlled rollouts
      const updateConfig: any = {};
      if (ng.maxUnavailable !== undefined) {
        updateConfig.maxUnavailable = ng.maxUnavailable;
      } else {
        // Default to 1 for controlled rolling updates
        updateConfig.maxUnavailable = 1;
      }

      // Create node group with appropriate configuration
      new eks.NodeGroupV2(
        nodeGroupName,
        {
          cluster: this.cluster,
          instanceTypes: ng.instanceTypes,
          desiredCapacity: ng.desiredSize,
          minSize: ng.minSize,
          maxSize: ng.maxSize,
          labels: ng.labels || {},
          taints: formattedTaints,
          nodeAssociatePublicIpAddress: false, // Security best practice: private IPs only
          nodeRootVolumeSize: ng.diskSize || 20, // Default to 20GB if not specified
          amiType: ng.amiType || "AL2_x86_64", // Default to Amazon Linux 2
          capacityType: ng.capacityType || "ON_DEMAND", // Default to on-demand instances
          updateConfig: updateConfig, // Add update configuration
          tags: {
            ...(tags || {}),
            Name: nodeGroupName,
            ManagedBy: "pulumi",
          },
          // Configure spot instances if specified
          ...this.getSpotConfiguration(ng),
        },
        { parent: this }
      );
    });
  }

  /**
   * Formats taints into the structure expected by the EKS API.
   *
   * This pure function transforms the input taint array into the required map format
   * without side effects, following functional programming principles.
   *
   * Like a city translator that converts specialized work instructions from one
   * format to another that the city departments can understand and implement.
   * For example, converting "no heavy vehicles on this road" to the specific
   * signage format and code references that the road department needs.
   *
   * @param taints Array of taint configurations
   * @returns Formatted taints object or undefined if no taints
   */
  private formatTaints(
    taints?: { key: string; value: string; effect: string }[]
  ): { [key: string]: { value: string; effect: string } } | undefined {
    if (!taints) return undefined;

    return taints.reduce((acc, taint) => {
      acc[taint.key] = {
        value: taint.value,
        effect: taint.effect,
      };
      return acc;
    }, {} as { [key: string]: { value: string; effect: string } });
  }

  /**
   * Returns configuration for spot instances if specified.
   *
   * Following the DRY principle, this method encapsulates the spot configuration logic
   * in one place, making it easier to maintain and update.
   *
   * This is like having specialized guidelines for temporary workers in the city.
   * Just as seasonal workers need special identification badges and different
   * payment processing, spot instances need special configuration to identify
   * them within the Kubernetes system.
   *
   * Without this special identification, Kubernetes wouldn't be able to recognize
   * these nodes as potentially short-lived, leading to potential service disruptions
   * when they're terminated unexpectedly.
   *
   * @param nodeGroup The node group configuration
   * @returns Object with spot configuration or empty object
   */
  private getSpotConfiguration(nodeGroup: NodeGroupConfig): any {
    // For temporary workers (spot instances), add special identification
    if (nodeGroup.capacityType === "SPOT") {
      return {
        kubeletExtraArgs: "--node-labels=node.kubernetes.io/lifecycle=spot",
      };
    }
    return {};
  }

  /**
   * Creates the IAM role that allows the AWS Load Balancer Controller to manage resources.
   *
   * This method follows the Single Responsibility Principle by focusing solely on
   * the ALB controller role creation.
   *
   * Like establishing the authority and credentials for the city's transportation
   * department to build and manage bridges, tunnels, and highways that connect
   * the city to the outside world. Without this authority, the city would be isolated,
   * with no way for visitors to enter or leave efficiently.
   *
   * If this role isn't created or has insufficient permissions, your cluster won't
   * be able to create external load balancers, making it difficult to expose services
   * to the internet or other AWS services.
   *
   * @param name Base name for the resources
   * @returns The created IAM role
   */
  private createAlbControllerRole(name: string): aws.iam.Role {
    // Extract OIDC provider URL from the cluster
    const oidcProviderUrl = this.getOidcProviderUrl();

    // Get AWS account ID
    const callerIdentity = aws.getCallerIdentity();

    // Create IAM role for ALB controller
    const albControllerRole = new aws.iam.Role(
      `${name}-alb-controller`,
      {
        assumeRolePolicy: pulumi
          .all([oidcProviderUrl, callerIdentity])
          .apply(([providerUrl, identity]) => {
            if (!providerUrl) {
              throw new Error("OIDC Provider URL is undefined or empty");
            }

            return JSON.stringify({
              Version: "2012-10-17",
              Statement: [
                {
                  Effect: "Allow",
                  Principal: {
                    Federated: `arn:aws:iam::${identity.accountId}:oidc-provider/${providerUrl}`,
                  },
                  Action: "sts:AssumeRoleWithWebIdentity",
                  Condition: {
                    StringEquals: {
                      [`${providerUrl}:sub`]:
                        "system:serviceaccount:kube-system:aws-load-balancer-controller",
                    },
                  },
                },
              ],
            });
          }),
      },
      { parent: this }
    );

    // Attach policy to the role - grant the specific permissions needed
    this.attachAlbControllerPolicy(name, albControllerRole);

    return albControllerRole;
  }

  /**
   * Gets the OIDC provider URL from the cluster.
   *
   * Following best practices for handling Pulumi outputs by:
   * 1. Using the correct property paths for EKS cluster outputs
   * 2. Applying proper error handling for undefined values
   * 3. Returning a consistent Output<string> type
   *
   * This is like retrieving the official identity verification service details
   * that the city uses to validate worker credentials. It's a crucial piece of
   * infrastructure that allows for secure delegation of authority.
   *
   * Without a properly configured OIDC provider, your Kubernetes pods wouldn't
   * be able to securely assume IAM roles, similar to city workers being unable
   * to prove their authority when working with external agencies.
   *
   * @returns The OIDC provider URL as a Pulumi Output<string>
   */
  private getOidcProviderUrl(): pulumi.Output<string> {
    // For EKS clusters, the OIDC provider URL can be retrieved from
    // the `oidcIssuer` property
    return this.cluster.oidcIssuer.apply((issuer) => {
      // If the issuer is not available, return empty string
      if (!issuer) {
        return "";
      }

      // Remove the https:// prefix as AWS IAM references OIDC providers without it
      return issuer.replace("https://", "");
    });
  }

  /**
   * Attaches the required policy to the ALB controller role.
   *
   * Following Interface Segregation and Single Responsibility principles,
   * this method focuses only on policy attachment.
   *
   * Like defining the specific authorities and permissions for the transportation
   * department - which roads they can build, what types of infrastructure they
   * can modify, and what resources they can access. Without these specific
   * permissions, the department would have a title but no actual ability to do its job.
   *
   * If this policy isn't attached or has insufficient permissions, the ALB controller
   * would be unable to create, modify, or delete load balancers and related resources,
   * preventing your services from being exposed externally.
   *
   * @param baseName Base name for the policy
   * @param role The role to attach the policy to
   */
  private attachAlbControllerPolicy(
    baseName: string,
    role: aws.iam.Role
  ): void {
    new aws.iam.RolePolicy(
      `${baseName}-alb-controller-policy`,
      {
        role: role.id,
        policy: JSON.stringify({
          Version: "2012-10-17",
          Statement: [
            {
              Effect: "Allow",
              Action: [
                // EC2 permissions - for managing network components
                // Like allowing transportation dept to modify roads, tunnels, and bridges
                "ec2:DescribeVpcs",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeInstances",
                "ec2:DescribeSubnets",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeNatGateways",
                "ec2:CreateSecurityGroup",
                "ec2:CreateTags",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:RevokeSecurityGroupIngress",

                // ELB permissions - for managing load balancers
                // Similar to permissions for building and operating tollbooths and entry points
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:DescribeLoadBalancerAttributes",
                "elasticloadbalancing:DescribeListeners",
                "elasticloadbalancing:DescribeListenerCertificates",
                "elasticloadbalancing:DescribeSSLPolicies",
                "elasticloadbalancing:DescribeRules",
                "elasticloadbalancing:DescribeTargetGroups",
                "elasticloadbalancing:DescribeTargetGroupAttributes",
                "elasticloadbalancing:DescribeTargetHealth",
                "elasticloadbalancing:DescribeTags",

                // Creation permissions - building new infrastructure
                // Like permits to construct new entry points to the city
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:CreateRule",
                "elasticloadbalancing:CreateTargetGroup",

                // Modification permissions - updating existing infrastructure
                // Similar to authority to modify, upgrade, or redirect existing routes
                "elasticloadbalancing:ModifyListener",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:ModifyRule",
                "elasticloadbalancing:ModifyTargetGroup",
                "elasticloadbalancing:ModifyTargetGroupAttributes",
                "elasticloadbalancing:RegisterTargets",
                "elasticloadbalancing:DeregisterTargets",
                "elasticloadbalancing:SetIpAddressType",
                "elasticloadbalancing:SetSecurityGroups",
                "elasticloadbalancing:SetSubnets",

                // Deletion permissions - removing obsolete infrastructure
                // Like permissions to demolish old, unused entry points
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:DeleteTargetGroup",
                "elasticloadbalancing:DeleteListener",
                "elasticloadbalancing:DeleteRule",

                // Tag management - for tracking and categorization
                // Similar to authority to place official city markings and signs
                "elasticloadbalancing:AddTags",
                "elasticloadbalancing:RemoveTags",

                // WAF permissions - for web application firewall integration
                // Like coordinating with security forces to establish checkpoints
                "wafv2:GetWebACL",
                "wafv2:GetWebACLForResource",
                "wafv2:AssociateWebACL",
                "wafv2:DisassociateWebACL",
              ],
              // Resource wildcard "*" means these permissions apply to all relevant resources
              // Like giving jurisdiction throughout the city, not just specific districts
              Resource: "*",
            },
          ],
        }),
      },
      { parent: this }
    );
  }
}
