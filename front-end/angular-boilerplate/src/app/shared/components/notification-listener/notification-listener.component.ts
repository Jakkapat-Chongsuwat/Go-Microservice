import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { Subscription } from 'rxjs';
import { HotToastService } from '@ngxpert/hot-toast';
import { UseNotification } from '@app/@core/usecases/useNotification';

@Component({
  selector: 'app-notification-listener',
  template: '',
  standalone: true,
})
export class NotificationListenerComponent implements OnInit, OnDestroy {
  private subscription = new Subscription();
  private useNotification = inject(UseNotification);
  private toast = inject(HotToastService);

  ngOnInit(): void {
    this.subscription.add(
      this.useNotification.getNotifications().subscribe((notification) => {
        console.log('Global notification received:', notification);
        this.toast.show(`Notification: ${notification.message}`);
      }),
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }
}
