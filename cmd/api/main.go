package main

import (
	"context"
	"github.com/Filiphasan/rabbitmqgolang/configs"
	"time"
)

func main() {
	// Start the server
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs.LoadConfig()
	//appConfig := configs.GetConfig()

}
