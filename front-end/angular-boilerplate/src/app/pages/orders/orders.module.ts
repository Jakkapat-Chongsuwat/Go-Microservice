// src/app/pages/orders/orders.module.ts
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { OrdersRoutingModule } from './orders-routing.module';
import { OrderFormComponent } from './components/order-form/order-form.component';

@NgModule({
  imports: [CommonModule, OrdersRoutingModule, OrderFormComponent],
})
export class OrdersModule {}
