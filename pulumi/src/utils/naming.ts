/**
 * Utility functions for resource naming
 */

/**
 * Generate a standardized resource name
 *
 * @param projectName Project name prefix
 * @param environment Environment name (dev, staging, prod)
 * @param resourceType Type of resource (vpc, subnet, etc.)
 * @param suffix Additional suffix for the name
 * @returns Standardized resource name
 */
export function createResourceName(
  projectName: string,
  environment: string,
  resourceType: string,
  suffix?: string
): string {
  const baseName = `${projectName}-${environment}-${resourceType}`;
  return suffix ? `${baseName}-${suffix}` : baseName;
}

/**
 * Create a standard resource name for a component
 *
 * @param stackName Stack name (typically project-environment)
 * @param componentType Component type (vpc, eks, etc.)
 * @param suffix Additional identifier
 * @returns Standardized component name
 */
export function createComponentName(
  stackName: string,
  componentType: string,
  suffix?: string
): string {
  const baseName = `${stackName}-${componentType}`;
  return suffix ? `${baseName}-${suffix}` : baseName;
}
