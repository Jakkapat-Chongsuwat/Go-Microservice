import { Type } from 'class-transformer';

export class OrderEntity {
  userId: string;

  @Type(() => OrderItem)
  items: OrderItem[];

  constructor(partial: Partial<OrderEntity>) {
    Object.assign(this, partial);
  }

  get totalItems(): number {
    return this.items?.reduce((acc, item) => acc + item.quantity, 0) || 0;
  }
}

export class OrderItem {
  productId: string;
  quantity: number;

  constructor(partial: Partial<OrderItem>) {
    Object.assign(this, partial);
  }
}
