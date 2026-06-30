package main

import "GoLiteConfig/internal/router"

func main() {
	r := router.SetupRouter()
	err := r.Run(":8080")
	if err != nil {
		return
	}
}
