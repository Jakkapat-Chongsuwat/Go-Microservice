// pulumi/src/components/utils/cidr-calculator.ts

/**
 * CidrCalculator helps organize IP address space for network infrastructure,
 * similar to how urban planners divide land into zones and neighborhoods.
 */
export class CidrCalculator {
  /**
   * Calculate subnet CIDR block based on the VPC CIDR
   *
   * This is like dividing a large city (VPC) into distinct neighborhoods (subnets).
   *
   * In cloud infrastructure:
   * - The VPC is like the entire city limits - the total addressable space
   * - Subnets are like neighborhoods where specific buildings (services) will be located
   * - For EKS clusters, proper subnet planning is critical because:
   *   1. Different node groups may need to be in different logical areas
   *   2. Some services need public access (like load balancers in public subnets)
   *   3. Database and backend services should be in private subnets for security
   *   4. Control plane to worker node communication depends on proper networking
   *
   * Just as a city planner ensures neighborhoods have the right size and don't overlap,
   * this function ensures our network segments are properly sized and distinct.
   *
   * @param vpcCidr The CIDR block of the VPC (like the city boundaries)
   * @param subnetIndex The index of the subnet to calculate (which neighborhood to create)
   * @param totalSubnets Total number of subnets to create (how many neighborhoods needed)
   * @returns A calculated CIDR notation for the specific subnet
   */
  public static calculateSubnetCidr(
    vpcCidr: string,
    subnetIndex: number,
    totalSubnets: number
  ): string {
    // Split the CIDR into base IP address and network prefix length
    // This is like identifying the starting point and total size of the city
    const [baseIp, baseCidr] = vpcCidr.split("/");

    // Convert prefix to number (e.g., "10.0.0.0/16" -> 16)
    // This determines how many bits identify the network vs. hosts
    const baseCidrNum = parseInt(baseCidr, 10);

    // Calculate new CIDR prefix for subnets
    // We need to allocate enough bits to accommodate all our subnets
    // Like deciding how to divide the city into equal-sized neighborhoods
    const newCidrNum = baseCidrNum + Math.ceil(Math.log2(totalSubnets * 2));

    // Calculate the subnet size (number of IP addresses)
    // This determines how many "buildings" can fit in each "neighborhood"
    const subnetSize = 2 ** (32 - newCidrNum);

    // Split the base IP into octets (e.g., "10.0.0.0" -> [10, 0, 0, 0])
    // Breaking down the city's starting coordinates
    const octets = baseIp.split(".").map((octet) => parseInt(octet, 10));

    // Convert the base IP to a single number for easier math
    // This gives us a numeric representation of our starting point
    let ipNum =
      (octets[0] << 24) + (octets[1] << 16) + (octets[2] << 8) + octets[3];

    // Calculate the new IP for this subnet
    // This determines the starting point for this specific neighborhood
    ipNum += subnetSize * subnetIndex;

    // Convert back to IP address format (e.g., 167837952 -> "10.0.16.0")
    // Turning our numeric calculations back into a readable address
    const newOctets = [
      (ipNum >>> 24) & 0xff,
      (ipNum >>> 16) & 0xff,
      (ipNum >>> 8) & 0xff,
      ipNum & 0xff,
    ];

    // Return the complete CIDR notation for this subnet
    // The full "address" of this neighborhood with its size
    return `${newOctets.join(".")}/${newCidrNum}`;
  }
}
