import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { OrderRequestDto } from '../dtos/order-request.dto';
import { OrderResponseDto } from '../dtos/order-response.dto';

@Injectable({
  providedIn: 'root',
})
export class OrderService {
  private readonly apiUrl = 'http://localhost:9001/orders';

  constructor(private http: HttpClient) {}

  createOrder(order: OrderRequestDto): Observable<OrderResponseDto> {
    return this.http.post<OrderResponseDto>(this.apiUrl, order);
  }
}
