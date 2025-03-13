// user-service\internal\adapters\repository\gorm_user_repository.go

package repository

import (
	"context"
	"user-service/internal/adapters/columns"
	"user-service/internal/adapters/models"
	"user-service/internal/domain"
	"user-service/internal/usecases"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormUserRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

var _ usecases.UserRepository = (*GormUserRepository)(nil)

func NewGormUserRepo(db *gorm.DB, logger *zap.Logger) *GormUserRepository {
	return &GormUserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GormUserRepository) Save(ctx context.Context, user *domain.User) error {
	dbUser := models.GormDBUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: columns.ColumnID}},
			DoUpdates: clause.AssignmentColumns([]string{columns.ColumnUsername, columns.ColumnEmail}),
		}).
		Create(&dbUser).Error

	if err != nil {
		r.logger.Error("GORM failed to save user",
			zap.String("id", user.ID), zap.Error(err))
		return err
	}

	r.logger.Info("GORM saved user", zap.String("id", user.ID))
	return nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var dbUser models.GormDBUser

	err := r.db.WithContext(ctx).
		First(&dbUser, columns.ColumnID+" = ?", id).
		Error
	if err != nil {
		r.logger.Warn("GORM find by ID failed", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &domain.User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
	}, nil
}

func (r *GormUserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	var dbUsers []models.GormDBUser

	err := r.db.WithContext(ctx).
		Find(&dbUsers).
		Error
	if err != nil {
		r.logger.Warn("GORM find all failed", zap.Error(err))
		return nil, err
	}

	users := make([]*domain.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		users = append(users, &domain.User{
			ID:       dbUser.ID,
			Username: dbUser.Username,
			Email:    dbUser.Email,
		})
	}

	return users, nil
}
