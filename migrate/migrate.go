package main

import (
	"fmt"
	"onichan/database"
	"onichan/model"
	"onichan/utils"
)

func main() {
	utils.LoadEnv()
	loadDatabase()
}

func loadDatabase() {
	database.Connect()

	database.Database.AutoMigrate(&model.User{})
	database.Database.AutoMigrate(&model.Notification{})
	database.Database.AutoMigrate(&model.Post{})
	database.Database.AutoMigrate(&model.PostReaction{})
	database.Database.AutoMigrate(&model.Reaction{})
	database.Database.AutoMigrate(&model.Category{})
	database.Database.AutoMigrate(&model.Avatar{})
	database.Database.AutoMigrate(&model.Report{})

	fmt.Println("Migration completed successfully")
}
