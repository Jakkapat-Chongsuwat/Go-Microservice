import { z } from 'zod';

export const UserResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  email: z.string().email(),
});

export type UserResponseDto = z.infer<typeof UserResponseSchema>;
