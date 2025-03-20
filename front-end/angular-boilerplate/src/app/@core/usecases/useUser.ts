import { Injectable, inject } from '@angular/core';
import { map } from 'rxjs/operators';
import { UserService } from '@core/services/user.service';
import { UsersResponseSchema, UsersResponseDto } from '@core/dtos/user-response.dto';
import { plainToInstance } from 'class-transformer';
import { UserEntity } from '@core/entities/user.entity';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class UseUser {
  private readonly _userService: UserService = inject(UserService);

  getUsers(): Observable<UserEntity[]> {
    return this._userService.getUsers().pipe(
      map((response: any) => {
        const validatedResponse: UsersResponseDto = UsersResponseSchema.parse(response);
        const entities = validatedResponse.data.map((userDto) => plainToInstance(UserEntity, userDto));
        return entities;
      }),
    );
  }
}
