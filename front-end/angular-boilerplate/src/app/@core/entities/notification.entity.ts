export class NotificationEntity {
  id: string;
  type: string;
  message: string;
  createdAt: Date;

  constructor(partial: Partial<NotificationEntity>) {
    Object.assign(this, partial);
    if (!this.createdAt) {
      this.createdAt = new Date();
    } else if (typeof this.createdAt === 'string') {
      this.createdAt = new Date(this.createdAt);
    }
  }
}
