import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { OrderRequestDto } from '../dtos/order-request.dto';
import { OrderResponseDto } from '../dtos/order-response.dto';
import { environment } from 'src/environments/environment';

@Injectable({
  providedIn: 'root',
})
export class OrderService {
  private readonly apiUrl = environment.apiEndpoints.orders || `${environment.apiUrl}/orders`;

  constructor(private http: HttpClient) {}

  createOrder(order: OrderRequestDto): Observable<OrderResponseDto> {
    return this.http.post<OrderResponseDto>(this.apiUrl, order);
  }
}
