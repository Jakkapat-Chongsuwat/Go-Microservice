import { Injectable } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { Observable, Subject, timer, throwError, EMPTY } from 'rxjs';
import { catchError, retryWhen, switchMap, tap, delay } from 'rxjs/operators';
import { Logger } from './misc/logger.service';
import { environment } from '@env/environment';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private WS_URL = environment.websocketUrl;
  private socket$: WebSocketSubject<any> | null = null;
  private notificationsSubject = new Subject<any>();
  private readonly logger = new Logger('NotificationService');
  private reconnectionAttempts = 0;
  private readonly MAX_RECONNECT_ATTEMPTS = 5;

  constructor() {
    this.connect();
  }

  /**
   * Connect to the WebSocket endpoint and manage reconnection.
   */
  private connect(): void {
    if (this.socket$ !== null) {
      this.socket$.complete();
      this.socket$ = null;
    }

    this.logger.info(`Connecting to WebSocket at ${this.WS_URL}...`);

    // Check if URL is valid
    if (!this.WS_URL || this.WS_URL === '') {
      this.logger.error('WebSocket URL is not defined in environment');
      return;
    }

    this.socket$ = webSocket({
      url: this.WS_URL,
      // Use a simple pass-through deserializer to handle both JSON and text
      deserializer: (e) => {
        try {
          // Try to parse as JSON first
          return JSON.parse(e.data);
        } catch (err) {
          // If not valid JSON, return as text
          return e.data;
        }
      },
      openObserver: {
        next: () => {
          this.logger.info('WebSocket connected successfully');
          this.reconnectionAttempts = 0; // Reset counter on successful connection
        },
      },
      closeObserver: {
        next: (event) => {
          this.logger.warn(`WebSocket connection closed with code ${event.code}`);
          this.reconnect();
        }
      }
    });

    this.socket$
      .pipe(
        catchError((err) => {
          this.logger.error('WebSocket error:', err);
          this.reconnect();
          return EMPTY;
        })
      )
      .subscribe({
        next: (message) => {
          this.logger.debug('Received WebSocket message:', message);
          this.notificationsSubject.next(message);
        },
        error: (error) => {
          this.logger.error('WebSocket subscription error:', error);
        },
        complete: () => {
          this.logger.warn('WebSocket connection completed.');
        },
      });
  }

  private reconnect(): void {
    this.reconnectionAttempts++;
    if (this.reconnectionAttempts <= this.MAX_RECONNECT_ATTEMPTS) {
      const delay = Math.min(this.reconnectionAttempts * 2000, 10000);
      this.logger.info(`Attempting reconnection ${this.reconnectionAttempts}/${this.MAX_RECONNECT_ATTEMPTS} in ${delay}ms`);

      setTimeout(() => {
        this.connect();
      }, delay);
    } else {
      this.logger.error(`Failed to reconnect after ${this.MAX_RECONNECT_ATTEMPTS} attempts`);
    }
  }

  getNotifications(): Observable<any> {
    return this.notificationsSubject.asObservable();
  }

  // Method to manually send a message (optional)
  sendMessage(message: any): void {
    if (this.socket$ && !this.socket$.closed) {
      // Let RxJS handle the serialization as per version
      this.socket$.next(message);
    } else {
      this.logger.error('Cannot send message - WebSocket is not connected');
    }
  }

  // Properly close the connection when the service is destroyed
  ngOnDestroy() {
    if (this.socket$) {
      this.socket$.complete();
    }
  }
}
