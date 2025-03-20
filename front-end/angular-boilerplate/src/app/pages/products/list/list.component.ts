import { Component, OnInit, inject } from '@angular/core';
import { UseUser } from '@core/usecases/useUser';
import { UserEntity } from '@core/entities/user.entity';
import { HotToastService } from '@ngneat/hot-toast';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss'],
})
export class ListComponent implements OnInit {
  users: UserEntity[] = [];
  isLoading = true;

  private readonly _useUser = inject(UseUser);
  private readonly _toast = inject(HotToastService);

  ngOnInit() {
    this._useUser.getUsers().subscribe({
      next: (users) => {
        this.users = users;
        this.isLoading = false;
      },
      error: (error) => {
        console.error(error);
        this.isLoading = false;
      },
    });
  }

  userClicked() {
    this._toast.show('Products clicked');
  }
}
