package repo

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int
	Username     string
	PasswordHash string
}

type UserRepository struct {
	users  map[int]User
	nextID int
}

func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users:  make(map[int]User),
		nextID: 2,
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	repo.users[1] = User{
		Id:           1,
		Username:     "admin",
		PasswordHash: string(hash),
	}

	repo.users[2] = User{
		Id:           2,
		Username:     "user2",
		PasswordHash: string(hash),
	}

	return repo
}

// Получить пользователя по ID
func (r *UserRepository) GetByID(id int) (User, error) {
	user, ok := r.users[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

// Получить пользователя по username
func (r *UserRepository) GetByUsername(username string) (User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}
