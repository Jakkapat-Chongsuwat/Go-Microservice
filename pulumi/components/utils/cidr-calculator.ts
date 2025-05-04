export class CidrCalculator {
    /**
     * Calculate subnet CIDR block based on the VPC CIDR
     * @param vpcCidr The CIDR block of the VPC
     * @param subnetIndex The index of the subnet to calculate
     * @param totalSubnets Total number of subnets to create
     */
    public static calculateSubnetCidr(vpcCidr: string, subnetIndex: number, totalSubnets: number): string {
        const [baseIp, baseCidr] = vpcCidr.split('/');
        
        const baseCidrNum = parseInt(baseCidr, 10);
        const newCidrNum = baseCidrNum + Math.ceil(Math.log2(totalSubnets * 2));
        
        // Calculate the subnet size
        const subnetSize = 2 ** (32 - newCidrNum);
        
        // Split the base IP into octets
        const octets = baseIp.split('.').map(octet => parseInt(octet, 10));
        
        // Convert the base IP to a number
        let ipNum = (octets[0] << 24) + (octets[1] << 16) + (octets[2] << 8) + octets[3];
        
        // Calculate the new IP for this subnet
        ipNum += subnetSize * subnetIndex;
        
        // Convert back to IP address format
        const newOctets = [
            (ipNum >>> 24) & 0xFF,
            (ipNum >>> 16) & 0xFF,
            (ipNum >>> 8) & 0xFF,
            ipNum & 0xFF
        ];
        
        return `${newOctets.join('.')}/${newCidrNum}`;
    }
}
