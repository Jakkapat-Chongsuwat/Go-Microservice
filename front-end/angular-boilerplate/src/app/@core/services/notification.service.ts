import { Injectable, OnDestroy } from '@angular/core';
import { Observable, Subject } from 'rxjs';
import { Logger } from './misc/logger.service';
import { environment } from '@env/environment';

@Injectable({
  providedIn: 'root',
})
export class NotificationService implements OnDestroy {
  private WS_URL = environment.websocketUrl;
  private ws: WebSocket | null = null;
  private notificationsSubject = new Subject<any>();
  private readonly logger = new Logger('NotificationService');
  private reconnectionAttempts = 0;
  private readonly MAX_RECONNECT_ATTEMPTS = 5;
  private reconnectTimeout: any = null;

  constructor() {
    // Log the exact URL we're trying to connect to
    console.log('WebSocket URL:', this.WS_URL);

    // Wait a bit for Angular to fully initialize
    setTimeout(() => this.connect(), 1000);
  }

  private connect(): void {
    // Close existing connection if any
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.logger.info(`Connecting to WebSocket at ${this.WS_URL}...`);

    if (!this.WS_URL || this.WS_URL === '') {
      this.logger.error('WebSocket URL is not defined in environment');
      return;
    }

    try {
      // Create a native WebSocket connection
      this.ws = new WebSocket(this.WS_URL);

      // Set up event handlers
      this.ws.onopen = () => {
        this.logger.info('WebSocket connected successfully');
        this.reconnectionAttempts = 0;
      };

      this.ws.onmessage = (event) => {
        try {
          const data = typeof event.data === 'string' ? JSON.parse(event.data) : event.data;
          this.logger.debug('Received WebSocket message:', data);
          this.notificationsSubject.next(data);
        } catch (error) {
          this.logger.error('Error parsing WebSocket message:', error);
          this.notificationsSubject.next(event.data);
        }
      };

      this.ws.onerror = (error) => {
        this.logger.error('WebSocket error:', error);
        // Log more details for debugging
        console.error('WebSocket error details:', {
          url: this.WS_URL,
          readyState: this.ws?.readyState,
          error: error
        });
      };

      this.ws.onclose = (event) => {
        this.logger.warn(`WebSocket connection closed with code ${event.code}`);
        this.reconnect();
      };
    } catch (error) {
      this.logger.error('Error creating WebSocket:', error);
      this.reconnect();
    }
  }

  private reconnect(): void {
    this.reconnectionAttempts++;
    if (this.reconnectionAttempts <= this.MAX_RECONNECT_ATTEMPTS) {
      const delay = Math.min(this.reconnectionAttempts * 2000, 10000);
      this.logger.info(`Attempting reconnection ${this.reconnectionAttempts}/${this.MAX_RECONNECT_ATTEMPTS} in ${delay}ms`);

      this.reconnectTimeout = setTimeout(() => {
        this.connect();
      }, delay);
    } else {
      this.logger.error(`Failed to reconnect after ${this.MAX_RECONNECT_ATTEMPTS} attempts`);
    }
  }

  getNotifications(): Observable<any> {
    return this.notificationsSubject.asObservable();
  }

  sendMessage(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const data = typeof message === 'string' ? message : JSON.stringify(message);
      this.ws.send(data);
    } else {
      this.logger.error('Cannot send message - WebSocket is not connected');
    }
  }

  ngOnDestroy() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.notificationsSubject.complete();
  }
}
