package scrape

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/imroc/req/v3"
	"github.com/mmcdole/gofeed"
)

var oetFeedURL = "https://eth.elfak.ni.ac.rs/?feed=atom"
var oetKey = []byte("eth.oet.lastUpdated")
var oet1RoleMention = "<@&1125063606036340867>"
var oet2RoleMention = "<@&1126263921653854218>"

func OET(bot bot.Client) {
	reqC := req.C().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := reqC.R().Get(oetFeedURL)
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
	data, err := db.Get(oetKey)
	if err != nil {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0)
		db.Put(oetKey, data)
	}

	lastUpdated := int64(binary.LittleEndian.Uint64(data))

	if feed.UpdatedParsed.Unix() <= lastUpdated {
		return
	}

	var posts []*gofeed.Item

	for _, item := range feed.Items {
		posts = append([]*gofeed.Item{item}, posts...)
	}

	messageContent := ""
	oet1Mentioned := false
	oet2Mentioned := false

	var embeds []discord.Embed

	for _, post := range posts {
		if post.UpdatedParsed.Unix() <= lastUpdated {
			continue
		}
		if post.Categories[0] == "Obaveštenja OE1" && !oet1Mentioned {
			messageContent += oet1RoleMention + " "
			oet1Mentioned = true
		}
		if (post.Categories[0] == "Obaveštenja OE2" || post.Categories[0] == "Obaveštenja Lab. praktikum-OET") && !oet2Mentioned {
			messageContent += oet2RoleMention + " "
			oet2Mentioned = true
		}

		converter := md.NewConverter("", true, nil)
		desc, err := converter.ConvertString(post.Content)
		if err != nil {
			log.Error(err)
		}

		if len(desc) > 4096 {
			desc = desc[:4094] + "…"
		}

		footer := "Нова објава"
		if post.PublishedParsed.Unix() != post.UpdatedParsed.Unix() {
			footer = "Измењена објава"
		}

		embeds = append(embeds, discord.NewEmbedBuilder().
			SetAuthor("Теоријска електротехника", "", "").
			SetColor(0x3d5c74).
			SetTitle(post.Title).
			SetURL(post.Link).
			SetDescription(desc).
			SetFooterText(footer).
			SetTimestamp(*post.UpdatedParsed).
			Build())
	}

	_, err = bot.Rest().CreateMessage(
		snowflake.MustParse("1125946814873485392"),
		discord.NewMessageCreateBuilder().
			SetContent(messageContent).
			AddEmbeds(embeds...).
			Build(),
	)
	if err != nil {
		log.Error(err)
		return
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(feed.UpdatedParsed.Unix()))
	db.Put(oetKey, data)
}
