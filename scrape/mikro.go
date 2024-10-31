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
	"strings"
	"time"
)

var ekCode = "ek"
var ekURL = "https://mikro.elfak.ni.ac.rs/predmeti/elektronske-komponente/"
var ekRoleMention = "<@&1126264197634863177>"
var ekName = "Електронске компоненте (II семестар)"

var fizCode = "fiz"
var fizURL = "https://mikro.elfak.ni.ac.rs/predmeti/fizika/"
var fizRoleMention = "<@&1126264245273755729>"
var fizName = "Физика (I семестар)"

func mikro(bot bot.Client, courseCode string, courseURL string, roleMention string, courseName string) {
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

	htmlNew, err := doc.Find(".entry-content").Html()
	if err != nil {
		log.Error(err)
		return
	}

	key := []byte("mikro." + courseCode + ".htmlOld")

	db := utils.GetDB()
	htmlOld, err := db.Get(key)
	if err != nil {
		db.Put(key, []byte(""))
	}

	converter := md.NewConverter("", true, nil)

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
					SetAuthorName("Микроелектроника").
					SetTitle(courseName).
					SetURL(courseURL).
					SetDescription(desc).
					SetColor(0x00b9d6).
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

func FIZ(bot bot.Client) {
	mikro(bot, fizCode, fizURL, fizRoleMention, fizName)
}

func EK(bot bot.Client) {
	mikro(bot, ekCode, ekURL, ekRoleMention, ekName)
}

func MIKRO(bot bot.Client) {
	atomFeed(
		bot,
		"https://mikro.elfak.ni.ac.rs/?feed=atom",
		[]byte("mikro.news.lastUpdated"),
		"Микроелектроника | Новости",
		"<@&1144400479384768553>",
		0x00b9d6,
	)
}
