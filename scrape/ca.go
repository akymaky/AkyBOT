package scrape

import (
	"github.com/disgoorg/disgo/bot"
)

func CA(bot bot.Client) {
	atomFeed(
		bot,
		"https://ca.elfak.ni.ac.rs/?feed=atom",
		[]byte("ca.lastUpdated"),
		"CISCO Академија",
		"<@&1144399134594437251>",
		0x00bdec,
	)
}
