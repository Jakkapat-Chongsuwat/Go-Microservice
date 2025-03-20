import { z } from 'zod';

export const MetaSchema = z.object({
  total: z.number(),
});

export const BaseResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    data: dataSchema,
    meta: MetaSchema,
  });
