export class UserEntity {
  id: string;
  name: string;
  email: string;

  get displayName(): string {
    return this.name.toUpperCase();
  }

  constructor(partial: Partial<UserEntity>) {
    Object.assign(this, partial);
  }
}
