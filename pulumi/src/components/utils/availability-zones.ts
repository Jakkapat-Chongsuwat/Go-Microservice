import * as aws from "@pulumi/aws";

/**
 * AvailabilityZoneUtils provides helper methods for working with AWS Availability Zones.
 *
 * In cloud infrastructure, availability zones are like independent data centers within a region.
 * Using multiple AZs is crucial for high availability - similar to how critical infrastructure
 * in a city (hospitals, fire stations) is distributed across different areas to ensure
 * service continues even if one area experiences problems.
 */
export class AvailabilityZoneUtils {
  /**
   * Get availability zone suffixes or default to a, b, c
   *
   * This method provides the letter suffixes that identify specific availability zones.
   *
   * For EKS deployments, spreading across multiple AZs is essential because:
   * - It protects against datacenter-level failures
   * - It reduces latency for users in different geographic areas
   * - It allows the cluster to maintain operations even if an entire AZ goes down
   *
   * Think of it as ensuring your city has essential services available in multiple
   * neighborhoods, so if flooding affects the east side, the west and north sides
   * can still provide services to all citizens.
   *
   * @param zones Optional list of zone suffixes to use
   * @returns An array of availability zone suffixes, defaulting to ["a", "b", "c"]
   */
  public static getAvailabilityZoneSuffixes(zones?: string[]): string[] {
    return zones || ["a", "b", "c"];
  }

  /**
   * Get full availability zone names including region
   *
   * This method combines the AWS region with zone suffixes to create complete AZ identifiers.
   *
   * In an EKS cluster, these full AZ names are used to:
   * - Place worker nodes in different physical locations
   * - Configure auto-scaling groups to span multiple zones
   * - Set up networking components (subnets) in each zone
   *
   * This is similar to providing complete addresses that include both the city (region)
   * and specific neighborhood (zone), allowing resources to be precisely located
   * within AWS's global infrastructure.
   *
   * @param zones Optional list of zone suffixes to use
   * @returns An array of complete availability zone names (e.g., ["us-east-1a", "us-east-1b", "us-east-1c"])
   */
  public static getAvailabilityZoneNames(zones?: string[]): string[] {
    const suffixes = this.getAvailabilityZoneSuffixes(zones);
    const region = aws.config.region;
    return suffixes.map((suffix) => `${region}${suffix}`);
  }
}
