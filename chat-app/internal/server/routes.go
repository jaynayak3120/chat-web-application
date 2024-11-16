package server

import (
	jwtauth "chat-app/internal/Authentication"
	model "chat-app/internal/Models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", s.HelloWorldHandler)
	r.HandleFunc("/health", s.healthHandler)
	r.HandleFunc("/login", s.authenticateUser).Methods("POST")

	r.HandleFunc("/createuser", s.createUser).Methods("POST")
	r.HandleFunc("/getAllUsers", s.getAllUsers).Methods("GET")
	r.HandleFunc("/getAUser/{id}", s.GetAUser).Methods("GET")
	r.HandleFunc("/updateUserPassword/{id}&{password}", s.updateUserPassword).Methods("PUT")
	r.HandleFunc("/updateUserDetails/{id}", s.updateUserDetails).Methods("PUT")
	r.HandleFunc("/deleteUser/{id}", s.deleteUser).Methods("DELETE")

	r.HandleFunc("/chatroom", s.createChatRoom).Methods("POST")
	r.HandleFunc("/chatrooms", s.getAllChatRooms).Methods("GET")
	r.HandleFunc("/chatroom/{id}", s.getChatRoom).Methods("GET")
	r.HandleFunc("/chatroom/{id}", s.deleteChatRoom).Methods("DELETE")

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) authenticateUser(w http.ResponseWriter, r *http.Request) {
	var userCreds model.UserAuth

	_ = json.NewDecoder(r.Body).Decode(&userCreds)

	_, err := s.db.GetAUser(userCreds.UserName, userCreds.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Authentication failed, Invalid Credentials")
	} else {
		tokenString, err := jwtauth.CreateToken(userCreds.UserName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenString)
		}
	}
}

func verifyToken(r *http.Request) (bool, string) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return false, "Missing authorization Token"
	}

	tokenString = tokenString[len("Bearer "):]
	err := jwtauth.VerifyToken(tokenString)
	if err != nil {
		return false, "Invalid token"
	}
	return true, ""
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	fmt.Println("r.Body:", r.Body)
	var userData model.User
	_ = json.NewDecoder(r.Body).Decode(&userData)

	fmt.Println("userData:", userData)

	userCreation, err := s.db.CreateUser(userData)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	_, _ = w.Write([]byte(userCreation))
}

func (s *Server) getAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	usersData, err := s.db.GetAllUsers()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	json.NewEncoder(w).Encode(usersData)
}

func (s *Server) GetAUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	user, err := s.db.GetAUserv2(params["id"])
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (s *Server) updateUserPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	err := s.db.UpdateUserPassword(params["id"], params["password"])
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	updateUser, err := s.db.GetAUserv2(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, err)
		return
	}

	json.NewEncoder(w).Encode(&updateUser)
}

func (s *Server) updateUserDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	err := s.db.UpdateUserDetails(params["id"], user)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	updateUser, err := s.db.GetAUserv2(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, err)
		return
	}

	json.NewEncoder(w).Encode(&updateUser)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	err := s.db.DeleteUser(params["id"])
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	json.NewEncoder(w).Encode("User with Id " + params["id"] + " is deleted.")
}

// ChatRoom
func (s *Server) createChatRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	var chatroom model.ChatRoom
	_ = json.NewDecoder(r.Body).Decode(&chatroom)

	response, err := s.db.CreateChatRoom(chatroom)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, response)
}

func (s *Server) deleteChatRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	err := s.db.DeleteChatRoom(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, "ChatRoom with Id "+params["id"]+" is deleted.")
}

func (s *Server) getAllChatRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	chatRooms, err := s.db.GetAllChatRoom()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	json.NewEncoder(w).Encode(chatRooms)
}

func (s *Server) getChatRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isValid, errMessage := verifyToken(r)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errMessage)
		return
	}

	params := mux.Vars(r)
	chatRoom, err := s.db.GetChatRoom(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	json.NewEncoder(w).Encode(chatRoom)
}
