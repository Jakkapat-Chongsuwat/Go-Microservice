import { Injectable, inject } from '@angular/core';
import { OrderService } from '@core/services/order.service';
import { OrderRequestDto } from '@core/dtos/order-request.dto';
import { OrderResponseDto } from '@core/dtos/order-response.dto';
import { plainToInstance } from 'class-transformer';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { OrderEntity } from '@core/entities/order.entity';

@Injectable({
  providedIn: 'root',
})
export class UseOrder {
  private readonly _orderService = inject(OrderService);

  /**
   * Creates an order by sending a request to the API.
   * Transforms the response into a domain entity.
   */
  createOrder(order: OrderRequestDto): Observable<OrderEntity> {
    return this._orderService.createOrder(order).pipe(
      map((response: OrderResponseDto) => {
        return plainToInstance(OrderEntity, {
          id: response.orderId,
          userId: response.userId,
          status: response.status,
          items: response.items,
        });
      }),
    );
  }
}
