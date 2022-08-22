package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/echovl/orderflo-dev/db/mysql"
	"github.com/echovl/orderflo-dev/layerhub"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")

	db, err := mysql.New(&mysql.Config{
		DSN: dsn,
	})
	if err != nil {
		log.Fatal(err)

	}

	// err = createAdmins(db)
	// if err != nil {
	// 	log.Fatalf("creating admins: %s", err)
	// }

	err = createPublicFonts(db)
	if err != nil {
		log.Fatalf("creating fonts: %s", err)
	}

	err = createPublicFrames(db)
	if err != nil {
		log.Fatalf("creating frames: %s", err)
	}
}

func createAdmins(db layerhub.DB) error {
	admins := []*layerhub.User{
		{
			ID:            layerhub.UniqueID("user"),
			FirstName:     "Alonso",
			LastName:      "Villegas",
			Email:         "alonso.villegas@backium.co",
			PasswordHash:  "$2a$10$4MTHXX/xuCK2bFpfOgrvL.ELkhTUpmJRVDr1my1i9JaIaPI2t4Sre", // Test@123!
			EmailVerified: true,
			Role:          layerhub.UserRoleOwner,
			Source:        layerhub.AuthSourceEmail,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, admin := range admins {
		err := db.PutUser(context.TODO(), admin)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPublicFonts(db layerhub.DB) error {
	fb, err := os.ReadFile("./scripts/seed/fonts.json")
	if err != nil {
		return err
	}

	fonts := struct {
		Fonts []layerhub.Font `json:"fonts"`
	}{}
	err = json.Unmarshal(fb, &fonts)
	if err != nil {
		return err
	}

	for i := range fonts.Fonts {
		fonts.Fonts[i].Public = true
	}

	return db.BatchCreateFonts(context.TODO(), fonts.Fonts)
}

func createPublicFrames(db layerhub.DB) error {
	publicFrames := []*layerhub.Frame{
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Presentation (4:3)",
			Width:   1024,
			Height:  768,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-presentation-4x3.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Presentation (16:9)",
			Width:   1920,
			Height:  1080,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-presentation-16x9.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Social Media Story",
			Width:   1080,
			Height:  1920,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-social-media-story.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Instagram Post",
			Width:   1080,
			Height:  1080,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-social-media-post.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Facebook Post",
			Width:   1200,
			Height:  1200,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-social-media-post.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Facebook Cover / Page Cover",
			Width:   1702,
			Height:  630,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-facebook-event-cover.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Facebook Event Cover",
			Width:   1920,
			Height:  1080,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-facebook-event-cover.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "YouTube Channel Art",
			Width:   2560,
			Height:  1440,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-youtube-channel-art.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "YouTube Thumbnail",
			Width:   1280,
			Height:  720,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-youtube-thumbnail.svg",
		},
		{
			ID:      layerhub.UniqueID("frame"),
			Name:    "Twitter Post",
			Width:   1200,
			Height:  675,
			Unit:    "px",
			Public:  true,
			Preview: "https://ik.imagekit.io/scenify/social-twittter-post.svg",
		},
	}

	for _, frame := range publicFrames {
		err := db.PutFrame(context.TODO(), frame)
		if err != nil {
			return err
		}
	}

	return nil
}
