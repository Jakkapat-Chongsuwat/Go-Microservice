import * as aws from "@pulumi/aws";

export class AvailabilityZoneUtils {
    /**
     * Get availability zone suffixes or default to a, b, c
     */
    public static getAvailabilityZoneSuffixes(zones?: string[]): string[] {
        return zones || ["a", "b", "c"];
    }

    /**
     * Get full availability zone names including region
     */
    public static getAvailabilityZoneNames(zones?: string[]): string[] {
        const suffixes = this.getAvailabilityZoneSuffixes(zones);
        const region = aws.config.region;
        return suffixes.map(suffix => `${region}${suffix}`);
    }
}
