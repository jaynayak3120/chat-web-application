package database

import (
	model "chat-app/internal/Models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	CreateRefreshToken(user_id string, refreshTokenString string) error

	GetRefreshToken(refreshTokenString string) (bool, error)

	UpdateRefreshToken(refreshTokenString string, user_id string) error

	DeleteRefreshToken() (string, error)

	CreateUser(user model.User) (string, error)

	GetAllUsers() ([]model.User, error)

	GetAUser(userName string, pass string) (model.User, error)

	UpdateUserPassword(Id string, password string) error

	UpdateUserDetails(Id string, user model.User) error

	DeleteUser(Id string) error

	GetAUserv2(Id string) (model.User, error)

	CreateChatRoom(chatRoom model.ChatRoom) (string, error)

	DeleteChatRoom(Id string) error

	GetAllChatRoom() ([]model.ChatRoom, error)

	GetChatRoom(Id string) (model.ChatRoom, error)

	CreateMessage(message model.Message) (string, error)

	GetMessagesForChatRoom(chatRoomId string) ([]model.Message, error)

	GetMessagesforIndividualChat(senderReceiver map[string]string) ([]model.Message, error)

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	dbname     = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	dbInstance *service
)

func PrintEnv() {
	fmt.Println("port>>>", port)
	fmt.Println("host: ", host)
	fmt.Println("dbname<<<<", dbname)
}

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname))
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}
	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) CreateRefreshToken(user_id string, refreshTokenString string) error {
	result, err := s.db.Exec("INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES(?, ?, ?)", &user_id, &refreshTokenString,
		time.Now().Add(time.Hour*24))
	if err != nil {
		return err
	}
	refreshToken_id, _ := result.LastInsertId()
	fmt.Println("Refresh-token is inserted in the system with Id:", refreshToken_id)
	return nil
}

func (s *service) GetRefreshToken(refreshTokenString string) (bool, error) {
	var isValid bool
	err := s.db.QueryRow("SELECT is_valid FROM refresh_tokens WHERE token = ? AND expires_at > ?", &refreshTokenString,
		time.Now()).Scan(&isValid)
	if err != nil {
		return false, err
	}
	if !isValid {
		return isValid, fmt.Errorf("refresh token is not valid")
	}
	return isValid, nil
}

func (s *service) UpdateRefreshToken(refreshTokenString string, user_id string) error {
	_, err := s.db.Exec("Update refresh_tokens SET is_valid = ? WHERE token = ? AND user_id = ?", false, &refreshTokenString, &user_id)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteRefreshToken() (string, error) {
	result, err := s.db.Exec("DELETE FROM refresh_tokens WHERE is_valid = false")
	if err != nil {
		return "", err
	}

	totalDeletedRows, _ := result.RowsAffected()
	fmt.Println("Total Deleted refresh tokens:", totalDeletedRows)

	return fmt.Sprint("Invalid Refresh Tokens Deleted: ", totalDeletedRows), nil
}

func (s *service) CreateUser(user model.User) (string, error) {

	result, err := s.db.Exec("INSERT INTO user (username, password_hash, Name, email, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)",
		&user.UserName, &user.Password, &user.Name, &user.Email, time.Now(), time.Now())
	if err != nil {
		return "", err
	}
	userID, _ := result.LastInsertId()

	return "User is inserted with ID: " + fmt.Sprintf("%d", userID), nil
}

func (s *service) GetAllUsers() ([]model.User, error) {
	var users = []model.User{}
	rows, err := s.db.Query("SELECT username, Name, email FROM user")
	if err != nil {
		return users, err
	}
	defer rows.Close()
	fmt.Println("rows:", rows)

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.UserName, &user.Name, &user.Email); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	fmt.Println("users:", users)

	return users, nil
}

func (s *service) GetAUser(userName string, pass string) (model.User, error) {
	var user model.User

	err := s.db.QueryRow("SELECT id, username, password_hash, Name, email, created_at, updated_at FROM user WHERE username = ? AND password_hash = ?",
		userName, pass).Scan(&user.Id, &user.UserName, &user.Password, &user.Name, &user.Email, &user.Created_at, &user.Upated_at)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) GetAUserv2(Id string) (model.User, error) {
	var user model.User

	err := s.db.QueryRow("SELECT id, username, Name, password_hash, email, created_at, updated_at FROM user WHERE id = ?",
		Id).Scan(&user.Id, &user.UserName, &user.Name, &user.Password, &user.Email, &user.Created_at, &user.Upated_at)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) UpdateUserPassword(Id string, password string) error {
	result, err := s.db.Exec("UPDATE user SET password_hash = ?, updated_at = ? WHERE id = ?",
		password, time.Now(), Id)
	if err != nil {
		return err
	}

	fmt.Printf("result: %v\n", result)
	return nil
}

func (s *service) UpdateUserDetails(Id string, user model.User) error {
	result, err := s.db.Exec("UPDATE user SET email = ?, Name = ?, updated_at = ? WHERE id = ?",
		&user.Email, &user.Name, time.Now(), Id)
	if err != nil {
		return err
	}

	fmt.Println("result: ", result)
	return nil
}

func (s *service) DeleteUser(Id string) error {
	result, err := s.db.Exec("DELETE FROM user WHERE Id = ?", Id)
	if err != nil {
		return err
	}

	totalDeletedRows, _ := result.RowsAffected()
	fmt.Println("Total Deleted Rows:", totalDeletedRows)

	if totalDeletedRows > 0 {
		return nil
	}
	return fmt.Errorf("No Chatroom exists with Id: " + Id)
}

// Chatroom
func (s *service) CreateChatRoom(chatRoom model.ChatRoom) (string, error) {
	result, err := s.db.Exec("INSERT INTO chatroom (name, description, created_at, updated_at) VALUES(?, ?, ?, ?)",
		&chatRoom.Name, &chatRoom.Description, time.Now(), time.Now())
	if err != nil {
		return "", err
	}
	chatRoomID, _ := result.LastInsertId()

	return "ChatRoom is created with ID: " + fmt.Sprintf("%d", chatRoomID), nil
}

func (s *service) DeleteChatRoom(Id string) error {
	result, err := s.db.Exec("DELETE FROM chatroom WHERE chatroomid = ?", Id)
	if err != nil {
		return err
	}

	totalDeletedRows, _ := result.RowsAffected()
	fmt.Println("Total Deleted Rows:", totalDeletedRows)

	if totalDeletedRows > 0 {
		return nil
	} else {
		return fmt.Errorf("No Chatroom exists with Id: " + Id)
	}
}

func (s *service) GetChatRoom(Id string) (model.ChatRoom, error) {
	var chatRoom model.ChatRoom
	err := s.db.QueryRow("SELECT chatRoomId, Name, description, created_at, updated_at FROM chatroom WHERE chatRoomId = ?",
		&Id).Scan(&chatRoom.ChatRoomId, &chatRoom.Name, &chatRoom.Description, &chatRoom.Created_at, &chatRoom.Upated_at)
	if err != nil {
		return chatRoom, err
	}
	return chatRoom, nil
}

func (s *service) GetAllChatRoom() ([]model.ChatRoom, error) {
	var chatRooms []model.ChatRoom
	rows, err := s.db.Query("SELECT chatroomId, name, description, created_at, updated_at FROM chatroom")
	if err != nil {
		return chatRooms, err
	}
	defer rows.Close()
	for rows.Next() {
		var chatRoom model.ChatRoom
		if err := rows.Scan(&chatRoom.ChatRoomId, &chatRoom.Name, &chatRoom.Description, &chatRoom.Created_at, &chatRoom.Upated_at); err != nil {
			return chatRooms, err
		}
		chatRooms = append(chatRooms, chatRoom)
	}
	return chatRooms, nil
}

func (s *service) CreateMessage(message model.Message) (string, error) {
	if message.Receiver_Id != "" && message.ChatRoomId == "" {
		result, err := s.db.Exec("INSERT INTO message (sender_id, receiver_id, content, created_at) VALUES(?, ?, ?, ?)",
			&message.Sender_Id, &message.Receiver_Id, &message.Content, time.Now())

		if err != nil {
			return "", err
		}
		messageID, _ := result.LastInsertId()

		return "Message is created with ID: " + fmt.Sprintf("%d", messageID), nil
	}
	if message.ChatRoomId != "" && message.Receiver_Id == "" {
		result, err := s.db.Exec("INSERT INTO message (chatroomid, sender_id, content, created_at) VALUES(?, ?, ?, ?)",
			&message.ChatRoomId, &message.Sender_Id, &message.Content, time.Now())

		if err != nil {
			return "", err
		}
		messageID, _ := result.LastInsertId()

		return "Message is created with ID: " + fmt.Sprintf("%d", messageID), nil
	}
	return "", fmt.Errorf("there cannot have both chatroomid and receiver_id in the body")
}

func (s *service) GetMessagesForChatRoom(chatRoomId string) ([]model.Message, error) {
	var messages []model.Message
	rows, err := s.db.Query("SELECT messageid, chatroomid, sender_id, content, created_at FROM message WHERE chatroomid = ? ORDER BY created_at",
		chatRoomId)
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message model.Message
		err := rows.Scan(&message.MessageId, &message.ChatRoomId, &message.Sender_Id, &message.Content,
			&message.Created_at)
		if err != nil {
			return messages, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (s *service) GetMessagesforIndividualChat(senderReceiver map[string]string) ([]model.Message, error) {
	var messages []model.Message
	rows, err := s.db.Query("SELECT messageid, sender_id, receiver_id, content, created_at FROM message WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) ORDER BY created_at",
		senderReceiver["sender_id"], senderReceiver["receiver_id"], senderReceiver["receiver_id"], senderReceiver["sender_id"])
	if err != nil {
		return messages, err
	}

	for rows.Next() {
		var message model.Message
		err := rows.Scan(&message.MessageId, &message.Sender_Id, &message.Receiver_Id, &message.Content,
			&message.Created_at)
		if err != nil {
			return messages, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dbname)
	return s.db.Close()
}
