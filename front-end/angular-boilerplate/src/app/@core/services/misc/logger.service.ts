import { NGXLogger, NgxLoggerLevel } from 'ngx-logger';

export enum LogLevel {
  Off = 0,
  Error,
  Warning,
  Info,
  Debug,
}

/**
 * A Logger class that wraps NGXLogger to provide structured logging with contextual metadata.
 *
 * Usage:
 *   1. In your AppComponent (or initializer), set the NGXLogger instance:
 *        Logger.setNGXLogger(injectedNgxLogger);
 *
 *   2. Create loggers with an optional source and log:
 *        const log = new Logger('MyComponent');
 *        log.debug('Something happened');
 *
 *   3. You can disable logging for a specific channel by using:
 *        Logger.disableChannel('MyComponent');
 */
export class Logger {
  private static ngxLogger: NGXLogger;

  // A list of logger channels that are disabled.
  private static disabledChannels: Set<string> = new Set();

  static setNGXLogger(logger: NGXLogger): void {
    Logger.ngxLogger = logger;
  }

  static enableProductionMode(): void {
    if (Logger.ngxLogger) {
      Logger.ngxLogger.updateConfig({
        level: NgxLoggerLevel.WARN,
        serverLogLevel: NgxLoggerLevel.ERROR,
      });
    }
  }

  /**
   * Disable logging for a specific source/channel.
   */
  static disableChannel(channel: string): void {
    Logger.disabledChannels.add(channel);
  }

  /**
   * Enable logging for a specific source/channel.
   */
  static enableChannel(channel: string): void {
    Logger.disabledChannels.delete(channel);
  }

  /**
   * Check if logging is disabled for the given channel.
   */
  private static isChannelDisabled(channel?: string): boolean {
    return channel ? Logger.disabledChannels.has(channel) : false;
  }

  constructor(
    private readonly _source?: string,
    private _context?: any,
  ) {}

  setContext(context: any): void {
    this._context = context;
  }

  debug(...objects: any[]): void {
    this.log('DEBUG', objects);
  }

  info(...objects: any[]): void {
    this.log('INFO', objects);
  }

  warn(...objects: any[]): void {
    this.log('WARN', objects);
  }

  error(...objects: any[]): void {
    this.log('ERROR', objects);
  }

  private log(level: string, objects: any[]): void {
    // If this channel is disabled, don't log anything.
    if (Logger.isChannelDisabled(this._source)) {
      return;
    }

    const structuredLog = {
      timestamp: new Date().toISOString(),
      level: level,
      source: this._source,
      message: objects,
      context: this._context,
    };

    const logMessage = JSON.stringify(structuredLog);

    if (Logger.ngxLogger) {
      switch (level) {
        case 'DEBUG':
          Logger.ngxLogger.debug(logMessage);
          break;
        case 'INFO':
          Logger.ngxLogger.info(logMessage);
          break;
        case 'WARN':
          Logger.ngxLogger.warn(logMessage);
          break;
        case 'ERROR':
          Logger.ngxLogger.error(logMessage);
          break;
        default:
          Logger.ngxLogger.debug(logMessage);
      }
    } else {
      console.log(logMessage);
    }
  }
}
