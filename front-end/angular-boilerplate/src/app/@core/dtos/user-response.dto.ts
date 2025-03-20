import { z } from 'zod';
import { BaseResponseSchema } from './base.dto';

export const UserSchema = z.object({
  id: z.string(),
  username: z.string(),
  email: z.string().email(),
});

export const UsersResponseSchema = BaseResponseSchema(z.array(UserSchema));

export type UsersResponseDto = z.infer<typeof UsersResponseSchema>;
