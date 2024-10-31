package scrape

import (
	"crypto/tls"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/imroc/req/v3"
	"net/url"
	"strings"
	"time"
)

var ledaURL = "http://leda.elfak.ni.ac.rs"

var uueCode = "ek"
var uueURL = "http://leda.elfak.ni.ac.rs/education/Uvod%20u%20elek/"
var uueRoleMention = "<@&1126264183105798225>"
var uueName = "Увод у електронику (II семестар)"

func leda(bot bot.Client, courseCode string, courseURL string, roleMention string, courseName string) {
	reqC := req.C().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := reqC.R().Get(courseURL)
	if err != nil {
		log.Error(err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	htmlNew, err := doc.Find("table[width=\"615\"] > tbody tr:nth-child(2) > td.box").Html()
	if err != nil {
		log.Error(err)
		return
	}

	key := []byte("leda." + courseCode + ".htmlOld")

	db := utils.GetDB()
	htmlOld, err := db.Get(key)
	if err != nil {
		db.Put(key, []byte(""))
	}

	converter := md.NewConverter("", true, &md.Options{
		GetAbsoluteURL: func(selec *goquery.Selection, rawURL string, domain string) string {
			baseURL := courseURL
			if rawURL[0] == '/' {
				baseURL = ledaURL
			}

			path, _ := url.JoinPath(baseURL, rawURL)
			return path
		},
	})

	linesOld := strings.Split(string(htmlOld), "\n")
	linesNew := strings.Split(htmlNew, "\n")

	desc := ""

	for _, sNew := range linesNew {
		matches := false
		for _, sOld := range linesOld {
			if strings.Compare(sNew, sOld) == 0 {
				matches = true
				break
			}
		}
		if !matches && len(sNew) > 0 {
			markdown, err := converter.ConvertString(sNew)
			if err == nil {
				desc += markdown + "\n"
			}
		}
	}

	if len(desc) > 0 {
		if len(desc) > 4096 {
			desc = desc[:4094] + "…"
		}
		_, err := bot.Rest().CreateMessage(
			snowflake.MustParse("1125946814873485392"),
			discord.NewMessageCreateBuilder().
				SetContent(roleMention).
				AddEmbeds(discord.NewEmbedBuilder().
					SetAuthorName("LEDA").
					SetTitle(courseName).
					SetURL(courseURL).
					SetDescription(desc).
					SetColor(0x00ffff).
					SetFooterText("Детектоване промене странице").
					SetTimestamp(time.Now()).
					Build()).
				Build(),
		)
		if err != nil {
			log.Error(err)
			return
		}
	}

	err = db.Put(key, []byte(htmlNew))
	if err != nil {
		log.Error(err)
	}
}

func UUE(bot bot.Client) {
	leda(bot, uueCode, uueURL, uueRoleMention, uueName)
}
