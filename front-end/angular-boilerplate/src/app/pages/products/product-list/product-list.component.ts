import { Component, OnDestroy, OnInit, inject } from '@angular/core';
import { Subscription } from 'rxjs';
import { ProductEntity } from '@app/@core/entities/product.entity';
import { UseProduct } from '@app/@core/usecases/useProduct';
import { HotToastService } from '@ngxpert/hot-toast';
import { Logger } from '@app/@core/services/misc/logger.service';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-product-list',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './product-list.component.html',
  styleUrls: ['./product-list.component.scss'],
})
export class ProductListComponent implements OnInit, OnDestroy {
  products: ProductEntity[] = [];
  isLoading = true;
  error: string | null = null;
  private subscription = new Subscription();

  private readonly logger = new Logger('ProductListComponent');

  private readonly _useProduct = inject(UseProduct);
  private readonly _toast = inject(HotToastService);
  private readonly router = inject(Router);

  ngOnInit(): void {
    this.logger.info('Initializing ProductListComponent');

    const sub = this._useProduct.getProducts().subscribe({
      next: (products) => {
        this.logger.debug('Products retrieved', products);
        this.products = products;
        this.isLoading = false;
      },
      error: (error) => {
        this.logger.error('Error retrieving products', error);
        this.error = error;
        this.isLoading = false;
      },
    });
    this.subscription.add(sub);
  }

  ngOnDestroy(): void {
    this.logger.info('Destroying ProductListComponent');
    this.subscription.unsubscribe();
  }

  productClicked(product: ProductEntity): void {
    this.logger.info('User clicked on a product', product);
    this._toast.show('Product clicked');
    this.router.navigate(['/orders/new', product.id]);
  }

  trackById(index: number, product: ProductEntity): any {
    return product.id;
  }
}
