import { Injectable } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { Observable, Subject, timer, throwError } from 'rxjs';
import { catchError, retryWhen, switchMap } from 'rxjs/operators';
import { Logger } from './misc/logger.service';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private WS_URL = 'ws://localhost:20052/ws';
  private socket$: WebSocketSubject<any>;
  private notificationsSubject = new Subject<any>();
  private readonly logger = new Logger('NotificationService');

  constructor() {
    this.connect();
  }

  /**
   * Connect to the WebSocket endpoint and manage reconnection.
   */
  private connect(): void {
    this.logger.info('Connecting to WebSocket...');
    this.socket$ = webSocket({
      url: this.WS_URL,
      deserializer: (msg) => JSON.parse(msg.data),
      openObserver: {
        next: () => {
          this.logger.info('WebSocket connected successfully');
        },
      },
    });

    this.socket$
      .pipe(
        retryWhen((errors) =>
          errors.pipe(
            switchMap((err) => {
              console.error('WebSocket error, attempting reconnection in 3 seconds...', err);
              return timer(3000);
            }),
          ),
        ),
        catchError((err) => {
          console.error('WebSocket error:', err);
          return throwError(() => err);
        }),
      )
      .subscribe({
        next: (message) => {
          this.notificationsSubject.next(message);
        },
        error: (error) => {
          console.error('WebSocket subscription error:', error);
        },
        complete: () => {
          console.warn('WebSocket connection closed.');
        },
      });
  }

  getNotifications(): Observable<any> {
    return this.notificationsSubject.asObservable();
  }
}
