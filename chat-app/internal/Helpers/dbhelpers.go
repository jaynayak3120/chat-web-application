package dbHelper

import (
	model "chat-app/internal/Models"
	"fmt"
)

func CreateUser(userData model.User) string {
	fmt.Println("Hello Users!")

	return "User has been created!"
}
