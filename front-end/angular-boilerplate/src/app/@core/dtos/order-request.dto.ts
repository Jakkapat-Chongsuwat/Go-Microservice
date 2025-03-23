// src/app/pages/orders/dtos/order-request.dto.ts
import { z } from 'zod';

export const OrderItemRequestSchema = z.object({
  productId: z.string().uuid(),
  quantity: z.number().min(1),
});

export const OrderRequestSchema = z.object({
  userId: z.string().uuid(),
  items: z.array(OrderItemRequestSchema),
});

export type OrderRequestDto = z.infer<typeof OrderRequestSchema>;
