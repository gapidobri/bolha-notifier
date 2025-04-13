package internal

import (
	"errors"
	"fmt"
	"github.com/gapidobri/bolha-notifier/internal/notifier"
	"github.com/gapidobri/bolha-notifier/internal/pkg/config"
	"github.com/gapidobri/bolha-notifier/internal/pkg/types"
	"github.com/gapidobri/bolha-notifier/internal/scraper"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

var existing []int

func Run() {
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.WithError(err).Fatal("failed to decode config")
	}

	val := validator.New(validator.WithRequiredStructEnabled())
	if err := val.Struct(&cfg); err != nil {
		var validateErrors validator.ValidationErrors
		if errors.As(err, &validateErrors) {
			for _, e := range validateErrors {
				log.Errorf("'%s' failed on '%s'", e.Field(), e.Tag())
			}
		}
		os.Exit(1)
	}

	nt, err := notifier.NewDiscordNotifier(cfg.WebhookUrl)
	if err != nil {
		log.WithError(err).Fatal("failed to create webhook client")
	}

	for _, url := range cfg.Urls {
		err = prefetch(url)
		if err != nil {
			log.WithError(err).WithField("url", url).Error("failed to prefetch url")
		}
	}

	log.Infof("watching %d urls", len(cfg.Urls))

	for range time.Tick(time.Duration(cfg.CheckInterval) * time.Second) {
		log.Info("running check")
		for _, url := range cfg.Urls {
			newPosts, err := getNew(url)
			if err != nil {
				log.WithError(err).Error("failed to fetch new posts")
				continue
			}

			if len(newPosts) == 0 {
				continue
			}

			err = nt.SendPosts(newPosts)
			if err != nil {
				log.WithError(err).Error("failed to send new posts")
			}
		}
	}
}

func prefetch(url string) error {
	posts, err := scraper.ParsePage(fmt.Sprintf("%s&sort=new", url))
	switch {
	case err == nil:
		break
	case errors.Is(err, scraper.ErrNoPosts{}):
		return nil
	default:
		return fmt.Errorf("error parsing page: %w", err)
	}

	for _, post := range posts {
		existing = append(existing, post.Id)
	}

	return nil
}

func getNew(url string) ([]types.Post, error) {
	page := 1
	var newPosts []types.Post

	for {
		posts, err := scraper.ParsePage(fmt.Sprintf("%s&sort=new&page=%d", url, page))
		switch {
		case err == nil:
			break
		case errors.Is(err, scraper.ErrNoPosts{}):
			return newPosts, nil
		default:
			return nil, fmt.Errorf("error parsing page: %w", err)
		}

		for _, post := range posts {
			if lo.Contains(existing, post.Id) {
				return newPosts, nil
			}

			newPosts = append(newPosts, post)
			existing = append(existing, post.Id)
		}

		page++
	}
}
