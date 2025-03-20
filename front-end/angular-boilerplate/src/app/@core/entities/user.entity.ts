export class UserEntity {
  id: string;
  username: string;
  email: string;
  phone?: string = 'N/A';
  location?: {
    city?: string;
    country?: string;
  } = { city: 'Unknown', country: 'Unknown' };
  picture?: {
    thumbnail?: string;
    medium?: string;
    large?: string;
  } = { thumbnail: 'default-thumbnail.png' };

  get displayName(): string {
    return this.username.toUpperCase();
  }

  constructor(partial: Partial<UserEntity>) {
    Object.assign(this, partial);
  }
}
