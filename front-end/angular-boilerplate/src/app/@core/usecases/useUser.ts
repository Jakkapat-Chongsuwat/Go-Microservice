import { Injectable, inject } from '@angular/core';
import { map } from 'rxjs/operators';
import { UserService } from '@core/services/user.service';
import { UserEntity } from '@core/entities/user.entity';
import { UserResponseSchema } from '@core/dtos/user-response.dto';
import { plainToInstance } from 'class-transformer';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class UseUser {
  private readonly _userService: UserService = inject(UserService);

  getUsers(): Observable<UserEntity[]> {
    return this._userService.getUsers().pipe(
      map((response: any[]) =>
        response.map((user) => {
          const validatedUser = UserResponseSchema.parse(user);
          return plainToInstance(UserEntity, validatedUser);
        }),
      ),
    );
  }
}
