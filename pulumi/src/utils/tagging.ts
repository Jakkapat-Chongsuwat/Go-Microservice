/**
 * Utility functions for managing AWS resource tags
 */

/**
 * Create standard resource tags
 *
 * @param projectName The name of the project
 * @param environment The deployment environment
 * @param additional Additional tags to add
 * @returns A record of tag key-value pairs
 */
export function createResourceTags(
  projectName: string,
  environment: string,
  additional?: Record<string, string>
): Record<string, string> {
  return {
    Project: projectName,
    Environment: environment,
    ManagedBy: "pulumi",
    ...additional,
  };
}

/**
 * Create name tag for a resource
 *
 * @param name Base resource name
 * @param suffix Optional suffix to append
 * @returns Tag object with a Name property
 */
export function createNameTag(
  name: string,
  suffix?: string
): Record<string, string> {
  return {
    Name: suffix ? `${name}-${suffix}` : name,
  };
}
