package domain

type User struct {
	ID    int    `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
}

// UserRepository interface defines the contract for user data operations
type UserRepository interface {
	FindAll() ([]User, error)
	FindByID(id int) (User, error)
	Create(user User) (User, error)
	Update(user User) (User, error)
	Delete(id int) error
}
