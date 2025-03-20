import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ProductsRoutingModule } from './products-routing.module';
import { ListComponent } from './list/list.component';

@NgModule({
  imports: [CommonModule, ProductsRoutingModule, ListComponent],
})
export class ProductsModule {}
