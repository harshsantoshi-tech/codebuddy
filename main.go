package main

import (
	"codebuddy/api"
	"codebuddy/config"
	"fmt"
	"log"
)

func main(){
	//load config
	cfg := config.Load()

	fmt.Println("Config loaded and server started on port : ", cfg.Port)

	//start server

	router := api.SetupRouter(cfg)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}