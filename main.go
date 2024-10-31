package main

import (
	"context"
	"github.com/akymaky/akybot/config"
	"github.com/akymaky/akybot/scrape"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.LoadConfig()
	utils.InitDB()
	db := utils.GetDB()
	defer db.Close()

	s := make(chan os.Signal, 1)

	client, err := disgo.New(config.Get().DiscordToken,
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentsAll)),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
		),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			log.Info("Logged in as %s#%s.\n", e.User.Username, e.User.Discriminator)
			//go radio.PlayRadio(e.Client())
			//go radio.UpdateNowPlaying(e.Client())
			go scrape.Init(e.Client())
		}),
		//bot.WithEventListenerFunc(func(e *events.GuildVoiceLeave) {
		//	if e.Member.User.ID != e.Client().ID() {
		//		return
		//	}
		//	time.Sleep(time.Second * 10)
		//	go radio.PlayRadio(e.Client())
		//}),
	)

	if err != nil {
		panic(err)
	}

	if err = client.OpenGateway(context.TODO()); err != nil {
		panic(err)
	}

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
