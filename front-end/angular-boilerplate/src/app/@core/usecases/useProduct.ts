import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { ProductEntity } from '../entities/product.entity';
import { ProductService } from '../services/product.service';
import { ProductsResponseDto, ProductsResponseSchema } from '../dtos/product-response.dto';
import { plainToInstance } from 'class-transformer';

@Injectable({
  providedIn: 'root',
})
export class UseProduct {
  private readonly _productService: ProductService = inject(ProductService);

  getProducts(): Observable<ProductEntity[]> {
    return this._productService.getProducts().pipe(
      map((response: any) => {
        const validatedResponse: ProductsResponseDto = ProductsResponseSchema.parse(response);
        const entities = validatedResponse.data.map((productDto) => plainToInstance(ProductEntity, productDto));
        return entities;
      }),
    );
  }
}
