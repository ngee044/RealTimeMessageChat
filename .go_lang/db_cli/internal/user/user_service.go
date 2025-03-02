package user

import (
	"gorm.io/gorm"

	"db_cli/internal/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(name, status string) (*models.User, error) {
	user := &models.User{
		Name:   name,
		Status: status,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	user := &models.User{}
	if err := s.db.First(user, id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user by ID
func (s *UserService) UpdateUser(id uint, name, status string) (*models.User, error) {
	user := &models.User{}
	if err := s.db.First(user, id).Error; err != nil {
		return nil, err
	}

	user.Name = name
	user.Status = status

	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id uint) error {
	user := &models.User{}
	if err := s.db.First(user, id).Error; err != nil {
		return err
	}

	if err := s.db.Delete(user).Error; err != nil {
		return err
	}

	return nil
}

// GetAllUsers - retrieves all users
func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
