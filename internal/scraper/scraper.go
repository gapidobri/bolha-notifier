package scraper

import (
	"encoding/json"
	"fmt"
	"github.com/gapidobri/bolha-notifier/internal/pkg/types"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func ParsePage(pageUrl string) ([]types.Post, error) {
	r, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	r.Header.Add("User-Agent", randomUserAgent())

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			log.WithError(err).Warn("failed to close response body")
		}
	}()

	root, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	nodes := getPostNodes(root)

	if len(nodes) == 0 {
		return nil, ErrNoPosts{}
	}

	var posts []types.Post

	for _, node := range nodes {
		var post types.Post

		optionsStr := getAttribute(node, "data-options")
		if optionsStr == nil {
			log.Error("id not found")
			continue
		}

		var options struct {
			Id int `json:"id"`
		}

		err = json.Unmarshal([]byte(*optionsStr), &options)
		if err != nil {
			log.Errorf("failed to unmarshal options: %v", err)
			continue
		}

		post.Id = options.Id

		url := getAttribute(node, "data-href")
		if url == nil {
			log.Error("url not found")
			continue
		}

		post.URL = "https://bolha.com" + *url

		titleNode := getElementByClassName(node, "entity-title")
		if titleNode == nil {
			log.Error("title not found")
			continue
		}

		post.Title = innerText(titleNode)

		thumbnailNode := getElementByClassName(node, "entity-thumbnail")
		if thumbnailNode == nil {
			log.Error("thumbnail not found")
			continue
		}

		for c := range node.Descendants() {
			if c.Type == html.ElementNode && c.Data == "img" {
				src, ok := lo.Find(c.Attr, func(a html.Attribute) bool {
					return a.Key == "data-src"
				})
				if ok {
					post.Thumbnail = "https:" + regexp.MustCompile(".*jpg").FindString(src.Val)
					break
				}
			}
		}

		priceNode := getElementByClassName(node, "price")
		if priceNode == nil {
			log.Error("price not found")
			continue
		}

		priceStr := innerText(priceNode)
		if priceStr != "Cena po dogovoru" {
			priceStr = strings.ReplaceAll(strings.Split(priceStr, " ")[0], ".", "")
			priceStr = strings.ReplaceAll(priceStr, ",", ".")
			priceStr = strings.TrimSpace(priceStr)
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				log.Errorf("failed to parse price: %v", err)
				continue
			}
			post.Price = &price
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func getPostNodes(n *html.Node) []*html.Node {
	var posts []*html.Node

	for c := range n.Descendants() {
		if c.Type == html.ElementNode && c.Data == "li" && lo.ContainsBy(c.Attr, func(attr html.Attribute) bool {
			return strings.Contains(attr.Val, "EntityList-item--Regular")
		}) {
			posts = append(posts, c)
		}
	}

	return posts
}

func getElementByClassName(n *html.Node, className string) *html.Node {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && lo.ContainsBy(c.Attr, func(attr html.Attribute) bool {
			return strings.Contains(attr.Key, "class") && strings.Contains(attr.Val, className)
		}) {
			return c
		}
	}
	return nil
}

func innerText(n *html.Node) string {
	var texts []string

	for c := range n.Descendants() {
		if c.Type == html.TextNode {
			texts = append(texts, c.Data)
		}
	}

	return strings.TrimSpace(strings.Join(texts, " "))
}

func getAttribute(n *html.Node, key string) *string {
	attr, ok := lo.Find(n.Attr, func(a html.Attribute) bool {
		return strings.Contains(a.Key, key)
	})
	if !ok {
		return nil
	}
	return lo.ToPtr(strings.TrimSpace(attr.Val))
}
