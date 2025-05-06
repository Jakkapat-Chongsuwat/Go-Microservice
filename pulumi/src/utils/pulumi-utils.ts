import * as pulumi from "@pulumi/pulumi";

/**
 * Utility functions for Pulumi operations
 */

/**
 * Create a mapping from a pulumi output to another type
 *
 * @param output The pulumi output to transform
 * @param fn The transformation function
 * @returns A new pulumi output with the transformed value
 */
export function mapOutput<T, U>(
  output: pulumi.Output<T>,
  fn: (value: T) => U
): pulumi.Output<U> {
  return output.apply(fn);
}

/**
 * Safely get a property from a potentially undefined object
 *
 * @param output The pulumi output that may be undefined
 * @param defaultValue The default value to use if undefined
 * @returns A new pulumi output with the property value or default
 */
export function getOutputOrDefault<T>(
  output: pulumi.Output<T | undefined>,
  defaultValue: T
): pulumi.Output<T> {
  return output.apply((value) => (value === undefined ? defaultValue : value));
}

/**
 * Combine multiple pulumi outputs into a single output
 *
 * @param outputs Object containing pulumi outputs
 * @returns A single pulumi output containing all values
 */
export function combineOutputs<T extends { [key: string]: pulumi.Input<any> }>(
  outputs: T
): pulumi.Output<{ [K in keyof T]: pulumi.Unwrap<T[K]> }> {
  // Use pulumi.all to properly combine the outputs with the correct typing
  return pulumi.all(outputs) as pulumi.Output<{
    [K in keyof T]: pulumi.Unwrap<T[K]>;
  }>;
}

/**
 * Helper function to handle conditional resource creation
 *
 * @param condition Boolean condition to evaluate
 * @param createFn Function to create the resource
 * @returns The created resource or undefined
 */
export function createResourceIf<T>(
  condition: boolean,
  createFn: () => T
): T | undefined {
  return condition ? createFn() : undefined;
}
