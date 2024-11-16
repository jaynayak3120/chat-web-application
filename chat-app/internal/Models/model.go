package model

type UserAuth struct {
	UserName string
	Password string
}

type User struct {
	Id         string `json:"id"`
	Name       string
	UserName   string `json:"username"`
	Password   string `json:"password_hash"`
	Email      string `json:"email"`
	Created_at string `json:"created_at"`
	Upated_at  string `json:"updated_at"`
}

type ChatRoom struct {
	ChatRoomId  string
	Name        string
	Description string
	Created_at  string
	Upated_at   string
}
