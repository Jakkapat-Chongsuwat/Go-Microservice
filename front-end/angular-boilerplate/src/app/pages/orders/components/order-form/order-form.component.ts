import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { Logger } from '@app/@core/services/misc/logger.service';
import { OrderService } from '@app/@core/services/order.service';
import { OrderRequestDto } from '@app/@core/dtos/order-request.dto';
import { HotToastService } from '@ngxpert/hot-toast';

@Component({
  selector: 'app-order-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './order-form.component.html',
  styleUrls: ['./order-form.component.scss'],
})
export class OrderFormComponent implements OnInit {
  orderForm!: FormGroup;
  private readonly logger = new Logger('OrderFormComponent');

  private readonly fb = inject(FormBuilder);
  private readonly orderService = inject(OrderService);
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
      this.orderService.createOrder(order).subscribe({
        next: (response) => {
          this.logger.info('Order created successfully', response);
          this._toast.show('Order created successfully');
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
