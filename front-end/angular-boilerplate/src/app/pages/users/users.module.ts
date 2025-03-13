import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { UsersRoutingModule } from './users-routing.module';
import { ListComponent } from './list/list.component';

@NgModule({
  imports: [CommonModule, UsersRoutingModule, ListComponent],
})
export class UsersModule {}
