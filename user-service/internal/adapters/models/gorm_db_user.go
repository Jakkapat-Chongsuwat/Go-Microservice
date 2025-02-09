package models

type GormDBUser struct {
	ID       string `gorm:"column:id;primaryKey"`
	Username string `gorm:"column:username"`
	Email    string `gorm:"column:email"`
}

func (GormDBUser) TableName() string {
	return "users"
}
