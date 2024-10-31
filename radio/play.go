package radio

import (
	"context"
	"github.com/akymaky/akybot/ffmpeg"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"time"
)

var (
	bbcRadio1StreamURL = "https://a.files.bbci.co.uk/media/live/manifesto/audio/simulcast/hls/nonuk/sbr_low/ak/bbc_radio_one.m3u8"
	guildID            = snowflake.MustParse("621752932517543967")
	channelID          = snowflake.MustParse("784107094269886487")
)

func PlayRadio(client bot.Client) {
	conn := client.VoiceManager().CreateConn(guildID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := conn.Open(ctx, channelID, false, true); err != nil {
		panic("error connecting to voice channel: " + err.Error())
	}

	if err := conn.SetSpeaking(ctx, voice.SpeakingFlagMicrophone); err != nil {
		panic("error setting speaking flag: " + err.Error())
	}

	provider, err := ffmpeg.New(bbcRadio1StreamURL)
	if err != nil {
		panic(err)
	}

	conn.SetOpusFrameProvider(provider)
}

func UpdateNowPlaying(client bot.Client) {
	scrapeBBC(client)
	for range time.Tick(time.Second * 30) {
		scrapeBBC(client)
	}
}
