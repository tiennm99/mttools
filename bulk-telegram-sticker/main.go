package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

type Sticker struct {
	Original string `json:"original"`
	Webp     string `json:"webp"`
	Emoji    string `json:"emoji"`
}

type Config struct {
	Title     string    `json:"title"`
	ShortName string    `json:"short_name"`
	Stickers  []Sticker `json:"stickers"`
}

func loadEnv() (string, string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", "", fmt.Errorf("load .env: %w", err)
	}

	apiIDStr := os.Getenv("API_ID")
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil || apiID == 0 {
		return "", "", "", fmt.Errorf("invalid API_ID: %s", apiIDStr)
	}

	apiHash := strings.TrimSpace(os.Getenv("API_HASH"))
	if apiHash == "" {
		return "", "", "", fmt.Errorf("API_HASH not set")
	}

	phone := strings.TrimSpace(os.Getenv("PHONE"))
	if phone == "" {
		return "", "", "", fmt.Errorf("PHONE not set")
	}

	return apiHash, phone, fmt.Sprintf("%d", apiID), nil
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if config.Title == "" || config.ShortName == "" || len(config.Stickers) == 0 {
		return nil, fmt.Errorf("invalid config: missing title, short_name or stickers")
	}

	return &config, nil
}

func main() {
	apiHash, phone, apiIDStr, err := loadEnv()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("API_ID: %s\n", apiIDStr)
	fmt.Printf("API_HASH: %s\n...\n", apiHash[:8])
	fmt.Printf("PHONE: %s\n", phone)

	client, err := gotgproto.NewClient(
		0, // appId will be ignored for phone auth
		apiHash,
		gotgproto.ClientTypePhone(phone),
		&gotgproto.ClientOpts{
			Session:  sessionMaker.SimpleSession(),
			InMemory: true,
		},
	)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}
	defer client.Stop()

	ctx := client.CreateContext()

	configPath := filepath.Join("output", "sticker_config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Run 'go run prepare.go' first!")
		return
	}

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Authenticating...")
	if err := client.Idle(); err != nil {
		fmt.Printf("Auth failed: %v\n", err)
		return
	}

	stickerBot, err := ctx.ResolveUsername("@Stickers")
	if err != nil {
		fmt.Printf("Resolve @Stickers: %v\n", err)
		return
	}

	fmt.Printf("Creating pack '%s' (%s)\n", config.Title, config.ShortName)

	// /newpack
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: "/newpack"})
	if err != nil {
		fmt.Printf("Send /newpack: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Title
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: config.Title})
	if err != nil {
		fmt.Printf("Send title: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Short name
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: config.ShortName})
	if err != nil {
		fmt.Printf("Send short_name: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Username (short_name)
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: config.ShortName})
	if err != nil {
		fmt.Printf("Send username: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	outputDir := "output"
	for _, sticker := range config.Stickers {
		stickerPath := filepath.Join(outputDir, sticker.Webp)
		if _, err := os.Stat(stickerPath); os.IsNotExist(err) {
			fmt.Printf("Warning: %s not found\n", stickerPath)
			continue
		}

		// Upload file
		uploader := message.NewUploader(ctx.Raw)
		inputFile, err := uploader.FromPath(ctx, stickerPath)
		if err != nil {
			fmt.Printf("Upload %s: %v\n", sticker.Webp, err)
			continue
		}

		media := &tg.InputMediaUploadedDocument{
			File:     inputFile,
			MimeType: "image/webp",
		}
		_, err = ctx.SendMedia(stickerBot.GetID(), &tg.MessagesSendMediaRequest{
			Media:   media,
			Message: sticker.Emoji,
		})
		if err != nil {
			fmt.Printf("Send sticker %s: %v\n", sticker.Webp, err)
			continue
		}
		time.Sleep(2 * time.Second)

		// Send emoji again
		time.Sleep(1 * time.Second)
		_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: sticker.Emoji})
		if err != nil {
			fmt.Printf("Send emoji for %s: %v\n", sticker.Webp, err)
		}
		time.Sleep(1 * time.Second)
	}

	// /publish
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: "/publish"})
	if err != nil {
		fmt.Printf("Send /publish: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Yes
	_, err = ctx.SendMessage(stickerBot.GetID(), &tg.MessagesSendMessageRequest{Message: "Yes"})
	if err != nil {
		fmt.Printf("Send Yes: %v\n", err)
		return
	}

	fmt.Printf("Sticker pack '%s' created successfully!\n", config.ShortName)
}
