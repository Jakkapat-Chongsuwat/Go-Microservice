// src/app/core/modules/core.module.ts

import { NgModule } from '@angular/core';
import { LoggerModule, NgxLoggerLevel } from 'ngx-logger';
import { environment } from '@env/environment';

@NgModule({
  imports: [
    LoggerModule.forRoot({
      level: environment.production ? NgxLoggerLevel.WARN : NgxLoggerLevel.TRACE,
    }),
  ],
  exports: [LoggerModule],
})
export class CoreModule {}
