package scrape

import "github.com/disgoorg/disgo/bot"

func ELFAK(bot bot.Client) {
	atomFeedJoomla(
		bot,
		"https://elfak.ni.ac.rs/informacije/novosti-i-obavestenja?format=feed&type=atom",
		[]byte("elfak.news.lastUpdated"),
		"Елфак | Новости и обавештења",
		"<@&1144399071470161993>",
		0x164d7d,
	)
}
