package scrape

import (
	"encoding/binary"
	"github.com/PuerkitoBio/goquery"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/imroc/req/v3"
	"strings"
	"time"
)

var sipKey = []byte("sip.lastUpdated")

type Post struct {
	title     string
	desc      string
	date      string
	link      string
	important bool
}

func CreatePost(s *goquery.Selection, important bool) *Post {
	link, _ := s.Find("a").Attr("href")

	return &Post{
		title:     s.Find("h4").Text(),
		desc:      s.Find("p:not(.date)").Text(),
		link:      link,
		date:      s.Find(".date").Text(),
		important: important,
	}
}

func (p *Post) Embed() discord.Embed {
	author := "Најновије вести"
	if p.important {
		author = "Важна обавештења"
	}

	return discord.NewEmbedBuilder().
		SetAuthor("СИП | "+author, "", "").
		SetTitle(p.title).
		SetDescription(p.desc).
		SetURL(p.link).
		SetFooterText("Нова вест").
		SetTimestamp(p.Timestamp()).
		SetThumbnail("https://i.imgur.com/dyu12dZ.png").
		SetColor(0x5bbc2e).
		Build()
}

func rsToUsDate(date string) string {
	newDate := strings.Trim(date, " \t\n\r")
	newDate = strings.ReplaceAll(newDate, "Јан", "Jan")
	newDate = strings.ReplaceAll(newDate, "Феб", "Feb")
	newDate = strings.ReplaceAll(newDate, "Мар", "Mar")
	newDate = strings.ReplaceAll(newDate, "Апр", "Apr")
	newDate = strings.ReplaceAll(newDate, "Мај", "May")
	newDate = strings.ReplaceAll(newDate, "Јун", "Jun")
	newDate = strings.ReplaceAll(newDate, "Јул", "Jul")
	newDate = strings.ReplaceAll(newDate, "Авг", "Aug")
	newDate = strings.ReplaceAll(newDate, "Сеп", "Sep")
	newDate = strings.ReplaceAll(newDate, "Окт", "Oct")
	newDate = strings.ReplaceAll(newDate, "Нов", "Nov")
	newDate = strings.ReplaceAll(newDate, "Дец", "Dec")

	newDate = strings.ReplaceAll(newDate, "Пон", "Mon")
	newDate = strings.ReplaceAll(newDate, "Уто", "Tue")
	newDate = strings.ReplaceAll(newDate, "Сре", "Wed")
	newDate = strings.ReplaceAll(newDate, "Чет", "Thu")
	newDate = strings.ReplaceAll(newDate, "Пет", "Fri")
	newDate = strings.ReplaceAll(newDate, "Суб", "Sat")
	newDate = strings.ReplaceAll(newDate, "Нед", "Sun")

	newDate = strings.ReplaceAll(newDate, "у", "at")

	return newDate
}

func (p *Post) Timestamp() time.Time {
	location, _ := time.LoadLocation("Europe/Belgrade")
	t, _ := time.ParseInLocation("Mon, 2. Jan, 2006. at 15:04", rsToUsDate(p.date), location)
	return t
}

func postsEqual(p *Post, l []byte) bool {
	return strings.Compare(p.String(), string(l)) == 0
}

func (p *Post) String() string {
	return p.title + " | " + p.date
}

func SIP(bot bot.Client) {
	db := utils.GetDB()
	data, err := db.Get(sipKey)
	if err != nil {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0)
		db.Put(sipKey, data)
	}

	lastUpdated := int64(binary.LittleEndian.Uint64(data))

	resp, err := req.Get("https://sip.elfak.ni.ac.rs")
	if err != nil {
		log.Error(err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	boxes := doc.Find(".news-box")
	all := goquery.NewDocumentFromNode(boxes.Get(0))
	important := goquery.NewDocumentFromNode(boxes.Get(1))

	var posts []*Post

	all.Find("ul > li").Each(func(i int, s *goquery.Selection) {
		post := CreatePost(s, false)

		if post.Timestamp().Unix() > lastUpdated {
			posts = append([]*Post{post}, posts...)
		}
	})

	important.Find("ul > li").Each(func(i int, s *goquery.Selection) {
		for _, post := range posts {
			if post.title == s.Find("h4").Text() && post.date == s.Find(".date").Text() {
				post.important = true
			}
		}
	})

	if len(posts) > 0 {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, uint64(time.Now().Unix()))
		db.Put(sipKey, data)

		var embeds []discord.Embed

		roleMentions := "<@&1127379612377284628>"
		roleModified := false

		for _, post := range posts {
			embeds = append(embeds, post.Embed())
			if !roleModified && post.important {
				roleModified = true
				roleMentions += " <@&1127379610380800011>"
			}
		}

		_, err := bot.Rest().CreateMessage(
			snowflake.MustParse("1144390838756069429"),
			discord.NewMessageCreateBuilder().
				SetContent(roleMentions).
				AddEmbeds(embeds...).
				Build(),
		)
		if err != nil {
			log.Error(err)
		}
	}
}
