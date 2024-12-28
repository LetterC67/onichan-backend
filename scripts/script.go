package main

import (
	"fmt"
	"onichan/database"
	"onichan/model"
	"onichan/utils"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func populateAvatar() {
	avatarUrl := []string{
		"https://www.svgrepo.com/show/411668/find.svg",
		"https://www.svgrepo.com/show/411671/father.svg",
		"https://www.svgrepo.com/show/411706/multiply.svg",
		"https://www.svgrepo.com/show/411707/mother.svg",
		"https://www.svgrepo.com/show/411716/pirate.svg",
		"https://www.svgrepo.com/show/411725/purr.svg",
		"https://www.svgrepo.com/show/411745/smile.svg",
		"https://www.svgrepo.com/show/411765/treat.svg",
		"https://www.svgrepo.com/show/411785/be.svg",
		"https://www.svgrepo.com/show/411768/tweet.svg",
		"https://www.svgrepo.com/show/411757/swim.svg",
		"https://www.svgrepo.com/show/411786/believe.svg",
		"https://www.svgrepo.com/show/411787/birth.svg",
		"https://www.svgrepo.com/show/411699/live.svg",
		"https://www.svgrepo.com/show/411724/pucker.svg",
		"https://www.svgrepo.com/show/411713/pick.svg",
		"https://www.svgrepo.com/show/411720/prick.svg",
		"https://www.svgrepo.com/show/411693/indulge.svg",
		"https://www.svgrepo.com/show/411701/make.svg",
		"https://www.svgrepo.com/show/411683/hallucinate.svg",
		"https://www.svgrepo.com/show/411663/eat.svg",
		"https://www.svgrepo.com/show/411759/throw.svg",
		"https://www.svgrepo.com/show/411744/slice.svg",
		"https://www.svgrepo.com/show/411748/solve.svg",
	}

	for _, url := range avatarUrl {
		avatar := model.Avatar{
			AvatarURL: url,
		}

		if err := database.Database.Create(&avatar).Error; err != nil {
			fmt.Println("Error creating avatar")
		}
	}

	fmt.Println("Avatar populated successfully")
}

func populateCategory() {
	categories := []struct {
		Name string
		Icon string
	}{
		{"technology", "https://www.svgrepo.com/show/411737/search.svg"},
		{"health", "https://www.svgrepo.com/show/411703/medicate.svg"},
		{"finance", "https://www.svgrepo.com/show/411636/chart.svg"},
		{"paranormal", "https://www.svgrepo.com/show/411688/haunt.svg"},
		{"cooking", "https://www.svgrepo.com/show/411677/fry.svg"},
		{"nature", "https://www.svgrepo.com/show/411747/snow.svg"},
		{"education", "https://www.svgrepo.com/show/411788/bookmark.svg"},
		{"travel", "https://www.svgrepo.com/show/411659/drive.svg"},
		{"music", "https://www.svgrepo.com/show/411715/play.svg"},
		{"gaming", "https://www.svgrepo.com/show/411676/game.svg"},
		{"photography", "https://www.svgrepo.com/show/411711/photograph.svg"},
		{"sport", "https://www.svgrepo.com/show/411694/kick.svg"},
		{"reading", "https://www.svgrepo.com/show/411726/read.svg"},
		{"movie", "https://www.svgrepo.com/show/411669/film.svg"},
		{"animal", "https://www.svgrepo.com/show/411725/purr.svg"},
		{"art", "https://www.svgrepo.com/show/411712/paint.svg"},
		{"fashion", "https://www.svgrepo.com/show/411654/dress.svg"},
		{"science", "https://www.svgrepo.com/show/411702/measure.svg"},
		{"advice", "https://www.svgrepo.com/show/411690/help.svg"},
		{"news", "https://www.svgrepo.com/show/411644/cover.svg"},
	}

	for _, category := range categories {
		cat := model.Category{
			Name:     category.Name,
			ImageURL: &category.Icon,
		}

		if err := database.Database.Create(&cat).Error; err != nil {
			fmt.Println("Error creating category")
		}
	}

	fmt.Println("Category populated successfully")
}

func populateReaction() {
	type Reaction struct {
		Name  string
		Emoji string
	}

	reactions := []Reaction{
		{Name: "like", Emoji: "👍"},
		{Name: "dislike", Emoji: "👎"},
		{Name: "love", Emoji: "❤️"},
		{Name: "haha", Emoji: "😂"},
		{Name: "wow", Emoji: "😮"},
		{Name: "sad", Emoji: "😢"},
		{Name: "angry", Emoji: "😡"},
		{Name: "congrats", Emoji: "🎉"},
	}

	for _, reaction := range reactions {
		r := model.Reaction{
			Name:  reaction.Name,
			Emoji: reaction.Emoji,
		}

		if err := database.Database.Create(&r).Error; err != nil {
			fmt.Println("Error creating reaction")
		}
	}

	fmt.Println("Reaction populated successfully")
}

func createAdmin(username string, email string, password string) {
	salt := utils.GetToken(32)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.DefaultCost)

	if err != nil {
		fmt.Println("Error hashing password")
		return
	}

	randomAvatar := utils.GetRandomAvatar()

	user := model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Salt:         salt,
		AvatarURL:    &randomAvatar,
		Role:         "admin",
	}

	if err := database.Database.Create(&user).Error; err != nil {
		fmt.Println("Error creating user")
		return
	}

	fmt.Println("Admin created successfully")
}

func auto() {
	populateAvatar()
	populateReaction()
	populateCategory()
	createAdmin("admin", "loremipsum@admin.com", "@dmin123")
}

func main() {
	if len(os.Args) == 0 {
		fmt.Println("Please provide a command")
		os.Exit(1)
	}

	utils.LoadEnv()
	database.Connect()

	if os.Args[1] == "populate_avatar" {
		populateAvatar()
		os.Exit(0)
	}

	if os.Args[1] == "populate_reaction" {
		populateReaction()
		os.Exit(0)
	}

	if os.Args[1] == "populate_category" {
		populateCategory()
		os.Exit(0)
	}

	if os.Args[1] == "create_admin" {
		if len(os.Args) < 5 {
			fmt.Println("Please provide username, email, and password")
			os.Exit(1)
		}
		createAdmin(os.Args[2], os.Args[3], os.Args[4])
		os.Exit(0)
	}

	if os.Args[1] == "auto" {
		auto()
		os.Exit(0)
	}

	fmt.Println("Script completed successfully")
}
