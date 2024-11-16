package main

import (
	"chat-app/internal/server"
	"fmt"
	"os"
)

func main() {
	fmt.Println("START>>>>")
	//database.PrintEnv()
	server := server.NewServer()
	fmt.Println("Server is Listning on Port:", os.Getenv("PORT"))
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("error:", err)
		panic(fmt.Sprintf("cannot start server: %s", err))
		//log.Fatal(err)
	}
	fmt.Println("Server is Listning on Port:", os.Getenv("PORT"))
}
