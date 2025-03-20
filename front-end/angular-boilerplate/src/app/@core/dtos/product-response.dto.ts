import { z } from 'zod';
import { BaseResponseSchema } from './base.dto';

export const ProductSchema = z.object({
  id: z.string(),
  name: z.string(),
  price: z.number(),
  quantity: z.number(),
});

export const ProductsResponseSchema = BaseResponseSchema(z.array(ProductSchema));

export type ProductsResponseDto = z.infer<typeof ProductsResponseSchema>;
