package scrape

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/imroc/req/v3"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

func noVerifyClient() *req.Client {
	return req.C().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

func interval(bot bot.Client) {
	SIP(bot)
	EK(bot)
	FIZ(bot)
	OET(bot)
	UUE(bot)
	ELFAK(bot)
	CA(bot)
	CS(bot)
	MIKRO(bot)
}

func Init(bot bot.Client) {
	interval(bot)
	for range time.Tick(time.Minute * 5) {
		interval(bot)
	}
}

func atomFeed(bot bot.Client, feedURL string, dbKey []byte, authorName string, messageContent string, embedColor int) {
	reqC := req.C().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := reqC.R().Get(feedURL)
	if err != nil {
		log.Error(err)
		return
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	db := utils.GetDB()
	data, err := db.Get(dbKey)
	if err != nil {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0)
		db.Put(dbKey, data)
	}

	lastUpdated := int64(binary.LittleEndian.Uint64(data))

	if feed.UpdatedParsed.Unix() <= lastUpdated {
		return
	}

	var posts []*gofeed.Item

	for _, item := range feed.Items {
		posts = append([]*gofeed.Item{item}, posts...)
	}

	for _, post := range posts {
		if post.UpdatedParsed.Unix() <= lastUpdated {
			continue
		}

		html := post.Content
		if len(html) == 0 {
			html = post.Description
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Error(err)
			return
		}

		var files []*discord.File

		doc.Find("img").Each(func(i int, selection *goquery.Selection) {
			imgUrl, exists := selection.Attr("src")
			if exists {
				resp, err := reqC.R().Get(imgUrl)
				if err != nil {
					log.Error(err)
					return
				}
				imgPath := strings.Split(imgUrl, "/")
				files = append(files, &discord.File{
					Name:   imgPath[len(imgPath)-1],
					Reader: resp.Body,
				})
			}
			selection.RemoveFiltered("img")
		}).Remove()

		converter := md.NewConverter("", true, nil)
		desc := converter.Convert(doc.Selection)

		if len(desc) > 1000 {
			desc = desc[:1000] + "…"
		}

		footer := "Нова објава"
		if post.PublishedParsed.Unix() != post.UpdatedParsed.Unix() {
			footer = "Измењена објава"
		}

		img := ""
		if post.Image != nil {
			img = post.Image.URL
		}

		_, err = bot.Rest().CreateMessage(
			snowflake.MustParse("1144390838756069429"),
			discord.NewMessageCreateBuilder().
				SetContent(messageContent).
				AddEmbeds(discord.NewEmbedBuilder().
					SetAuthorName(authorName).
					SetColor(embedColor).
					SetTitle(post.Title).
					SetURL(post.Link).
					SetDescription(desc).
					SetFooterText(footer).
					SetTimestamp(*post.UpdatedParsed).
					SetImage(img).
					Build()).
				AddFiles(files...).
				Build(),
		)
		if err != nil {
			log.Error(err)
			return
		}
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(feed.UpdatedParsed.Unix()))
	db.Put(dbKey, data)
}

func atomFeedJoomla(bot bot.Client, feedURL string, dbKey []byte, authorName string, messageContent string, embedColor int) {
	reqC := req.C().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := reqC.R().Get(feedURL)
	if err != nil {
		log.Error(err)
		return
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	db := utils.GetDB()
	data, err := db.Get(dbKey)
	if err != nil {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0)
		db.Put(dbKey, data)
	}

	lastUpdated := int64(binary.LittleEndian.Uint64(data))

	var posts []*gofeed.Item

	for _, item := range feed.Items {
		posts = append([]*gofeed.Item{item}, posts...)
	}

	maxTimestamp := lastUpdated

	for _, post := range posts {
		if post.UpdatedParsed.Unix() <= lastUpdated {
			continue
		}

		if post.UpdatedParsed.Unix() > maxTimestamp {
			maxTimestamp = post.UpdatedParsed.Unix()
		}

		resp, err := reqC.R().Get(post.Link)
		if err != nil {
			log.Error(err)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}

		contentWrap := doc.Find(".content-wrap")

		var files []*discord.File

		contentWrap.Find("img").Each(func(i int, selection *goquery.Selection) {
			imgUrl, exists := selection.Attr("src")
			if exists {
				resp, err := reqC.R().Get("https://" + md.DomainFromURL(feedURL) + imgUrl)
				if err != nil {
					log.Error(err)
					return
				}
				imgPath := strings.Split(imgUrl, "/")
				files = append(files, &discord.File{
					Name:   imgPath[len(imgPath)-1],
					Reader: resp.Body,
				})
			}
			selection.RemoveFiltered("img")
		}).Remove()
		contentWrap.Find("h2[itemprop=\"name\"]").Remove()

		converter := md.NewConverter(md.DomainFromURL(feedURL), true, nil)
		desc := converter.Convert(contentWrap)

		if len(desc) > 1000 {
			desc = desc[:1000] + "…"
		}

		footer := "Нова објава"
		if post.PublishedParsed.Unix() != post.UpdatedParsed.Unix() {
			footer = "Измењена објава"
		}

		img := ""
		if post.Image != nil {
			img = post.Image.URL
		}

		_, err = bot.Rest().CreateMessage(
			snowflake.MustParse("1144390838756069429"),
			discord.NewMessageCreateBuilder().
				SetContent(messageContent).
				AddEmbeds(discord.NewEmbedBuilder().
					SetAuthorName(authorName).
					SetColor(embedColor).
					SetTitle(post.Title).
					SetURL(post.Link).
					SetDescription(desc).
					SetFooterText(footer).
					SetTimestamp(*post.UpdatedParsed).
					SetImage(img).
					Build()).
				AddFiles(files...).
				Build(),
		)
		if err != nil {
			log.Error(err)
			return
		}
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(maxTimestamp))
	db.Put(dbKey, data)
}
