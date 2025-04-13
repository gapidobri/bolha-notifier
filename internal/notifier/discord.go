package notifier

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/gapidobri/bolha-notifier/internal/pkg/types"
)

type DiscordNotifier struct {
	client webhook.Client
}

func NewDiscordNotifier(url string) (*DiscordNotifier, error) {
	client, err := webhook.NewWithURL(url)
	if err != nil {
		return nil, fmt.Errorf("cannot create webhook client: %w", err)
	}

	return &DiscordNotifier{
		client: client,
	}, nil
}

func (n *DiscordNotifier) SendPosts(posts []types.Post) error {
	var embeds []discord.Embed

	for _, post := range posts {
		var fields []discord.EmbedField

		if post.Price != nil {
			fields = append(fields, discord.EmbedField{
				Name:  "Price",
				Value: fmt.Sprintf("%.2f €", *post.Price),
			})
		}

		embeds = append(embeds, discord.Embed{
			Title: post.Title,
			Type:  discord.EmbedTypeRich,
			URL:   post.URL,
			Image: &discord.EmbedResource{
				URL: post.Thumbnail,
			},
			Fields: fields,
		})
	}

	_, err := n.client.CreateMessage(discord.WebhookMessageCreate{
		Content: "New posts on bolha.com",
		Embeds:  embeds,
	})

	if err != nil {
		return fmt.Errorf("error sending posts: %w", err)
	}

	return nil
}
