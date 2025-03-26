import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { Logger } from '@app/@core/services/misc/logger.service';
import { OrderRequestDto } from '@core/dtos/order-request.dto';
import { HotToastService } from '@ngxpert/hot-toast';
import { UseOrder } from '@app/@core/usecases/useOrder';

@Component({
  selector: 'app-order-form',
  templateUrl: './order-form.component.html',
  styleUrls: ['./order-form.component.scss'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
})
export class OrderFormComponent implements OnInit {
  orderForm!: FormGroup;
  private readonly logger = new Logger('OrderFormComponent');

  private readonly fb = inject(FormBuilder);
  private readonly useOrder = inject(UseOrder);
  private readonly route = inject(ActivatedRoute);
  private readonly _toast = inject(HotToastService);

  FORM_ITEMS_KEY = 'items';
  productIdFromRoute: string | null = null;

  ngOnInit(): void {
    this.orderForm = this.fb.group({
      userId: ['', Validators.required],
      items: this.fb.array([this.createItem()], Validators.required),
    });
    this.logger.info('OrderFormComponent initialized');

    this.route.paramMap.subscribe((params) => {
      this.productIdFromRoute = params.get('productId');
      if (this.productIdFromRoute) {
        const itemsArray = this.orderForm.get(this.FORM_ITEMS_KEY) as FormArray;
        if (itemsArray && itemsArray.length > 0) {
          itemsArray.at(0).patchValue({ productId: this.productIdFromRoute });
        }
        this.logger.info('Product ID from route:', this.productIdFromRoute);
      }
    });
  }

  createItem(): FormGroup {
    return this.fb.group({
      productId: ['', Validators.required],
      quantity: [1, [Validators.required, Validators.min(1)]],
    });
  }

  get items(): FormArray {
    return this.orderForm.get(this.FORM_ITEMS_KEY) as FormArray;
  }

  addItem(): void {
    this.items.push(this.createItem());
  }

  submitOrder(): void {
    if (this.orderForm.valid) {
      const order: OrderRequestDto = this.orderForm.value;
      this.useOrder.createOrder(order).subscribe({
        next: (response) => {
          this.logger.info('Order created successfully', response);
        },
        error: (err) => {
          this.logger.error('Error creating order', err);
        },
      });
    } else {
      this.logger.warn('Order form is invalid');
    }
  }
}
