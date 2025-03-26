import { z } from 'zod';

export const NotificationSchema = z.object({
  id: z.string(),
  type: z.string(),
  message: z.string(),
  created_at: z.string().optional(),
});

export type NotificationDto = z.infer<typeof NotificationSchema>;
