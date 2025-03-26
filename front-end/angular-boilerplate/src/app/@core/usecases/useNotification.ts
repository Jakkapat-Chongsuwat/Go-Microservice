import { Injectable } from '@angular/core';
import { NotificationService } from '@core/services/notification.service';
import { NotificationEntity } from '@core/entities/notification.entity';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { plainToInstance } from 'class-transformer';
import { NotificationDto, NotificationSchema } from '../dtos/notification-receive.dto';

@Injectable({
  providedIn: 'root',
})
export class UseNotification {
  constructor(private notificationService: NotificationService) {}

  /**
   * Returns an observable stream of NotificationEntity items.
   * Raw notifications from the WebSocket are validated and transformed.
   */
  getNotifications(): Observable<NotificationEntity> {
    return this.notificationService.getNotifications().pipe(
      map((rawNotif: any) => {
        const validated: NotificationDto = NotificationSchema.parse(rawNotif);
        const entity = plainToInstance(NotificationEntity, {
          id: validated.id,
          type: validated.type,
          message: validated.message,
          createdAt: validated.created_at,
        });
        return entity;
      }),
    );
  }
}
