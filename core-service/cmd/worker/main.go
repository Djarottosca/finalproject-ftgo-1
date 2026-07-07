package main

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/bootstrap"

func main() {
	app := bootstrap.NewApp()
	app.RunWorker()
}
