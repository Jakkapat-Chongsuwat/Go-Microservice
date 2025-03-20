export class ProductEntity {
  id: string;
  name: string;
  price: number;
  quantity: number;

  get inStock(): boolean {
    return this.quantity > 0;
  }

  get displayName(): string {
    return this.name;
  }

  constructor(partial: Partial<ProductEntity>) {
    Object.assign(this, partial);
  }
}
