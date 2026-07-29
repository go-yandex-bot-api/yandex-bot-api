// Package main provides an example.
package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	yabotapi "github.com/go-yandex-bot-api/yandex-bot-api"
	"github.com/go-yandex-bot-api/yandex-bot-api/pkg/router"
)

func main() {
	bot, err := yabotapi.NewBot("YOUR_TOKEN_HERE", yabotapi.WithDebug(true))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
	ctx := context.Background()

	updatesChannel, err := bot.Updates.GetUpdatesChannel(ctx, yabotapi.NewUpdateConfig(0))
	if err != nil {
		log.Fatal("Failed to start polling:", err)
	}

	r := router.NewRouter(bot)

	// Example of sending an image from disk
	r.HandleCommand("image", func(c *router.Context) error {
		u := c.Update
		// Use the test.png image provided in the root of the example
		req := yabotapi.SendImageRequest{
			ChatID:   u.GetChatID(),
			FilePath: "test.png",
			Text:     "Look, this is an image sent from the disk!",
		}
		if req.ChatID == "" {
			req.Login = u.GetFromLogin()
		}

		_, sendErr := bot.Files.SendImage(ctx, req)
		return sendErr
	})

	// Example of sending a gallery of images
	r.HandleCommand("gallery", func(c *router.Context) error {
		u := c.Update
		req := yabotapi.SendGalleryRequest{
			ChatID:    u.GetChatID(),
			FilePaths: []string{"test.png", "test.png"},
			Text:      "Look, here is a gallery of two images!",
		}
		if req.ChatID == "" {
			req.Login = u.GetFromLogin()
		}

		_, sendErr := bot.Files.SendGallery(ctx, req)
		return sendErr
	})

	// Example of sending a document from disk
	r.HandleCommand("doc", func(c *router.Context) error {
		u := c.Update
		createDummyFile("report.txt")
		defer func() { _ = os.Remove("report.txt") }()

		req := yabotapi.SendFileRequest{
			ChatID:   u.GetChatID(),
			FilePath: "report.txt",
			Text:     "And here is a text document.",
		}
		if req.ChatID == "" {
			req.Login = u.GetFromLogin()
		}

		_, err := bot.Files.SendFile(c.Ctx, req)
		if err != nil {
			log.Println("Error sending document:", err)
			return err
		}
		return nil
	})

	// Intercept any incoming files/images from the user and download them
	r.HandleDefault(func(c *router.Context) error {
		u := c.Update
		// Check if the user attached a file
		//nolint:nestif // Example structure is straightforward enough
		if u.File != nil && u.File.ID != "" {
			log.Printf("User sent a file! FileID: %s", u.File.ID)

			// Download the file as an io.ReadCloser (streaming)
			stream, err := bot.Files.GetFileByID(c.Ctx, u.File.ID)
			if err != nil {
				log.Println("Error downloading file:", err)
				return err
			}
			defer func() { _ = stream.Close() }()

			// Save to disk
			out, err := os.Create("downloaded_" + u.File.Name)
			if err != nil {
				log.Println("Error creating file on disk:", err)
				return err
			}
			defer func() { _ = out.Close() }()

			if _, err := io.Copy(out, stream); err != nil {
				log.Println("Error saving file:", err)
				return err
			}

			log.Printf("File saved to: %s", out.Name())
			msg := yabotapi.NewReply(&u, "I saved your file!")
			if _, err := bot.Messages.SendText(c.Ctx, msg); err != nil {
				return err
			}
		} else {
			_, _ = bot.Messages.SendText(c.Ctx,
				yabotapi.NewReply(&u, "Send me the /image or /doc command, or just upload any file!"))
		}
		return nil
	})

	log.Println("Bot is running. Send /image, /doc, or attach a file.")
	r.Start(ctx, updatesChannel)
}

// createDummyFile creates an empty file for demonstration purposes.
func createDummyFile(name string) {
	f, _ := os.Create(filepath.Clean(name))
	_, _ = f.WriteString("hello world")
	_ = f.Close()
}
