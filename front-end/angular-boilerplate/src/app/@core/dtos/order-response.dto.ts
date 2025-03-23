// src/app/pages/orders/dtos/order-response.dto.ts
import { z } from 'zod';

export const OrderItemResponseSchema = z.object({
  productId: z.string().uuid(),
  quantity: z.number().min(1),
});

export const OrderResponseSchema = z.object({
  orderId: z.string().uuid(),
  userId: z.string().uuid(),
  status: z.enum(['CREATED', 'CONFIRMED', 'SHIPPED', 'CANCELLED']),
  items: z.array(OrderItemResponseSchema),
});

export type OrderResponseDto = z.infer<typeof OrderResponseSchema>;
