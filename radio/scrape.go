package radio

import (
	"bytes"
	"encoding/json"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/generaltso/vibrant"
	"github.com/imroc/req/v3"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var lastMessageKey = []byte("radio.last.message-id")
var lastTrackKey = []byte("radio.last.track")
var lastShowKey = []byte("radio.last.show")

func scrapeBBC(client bot.Client) {
	tracks := new(TracksJson)

	rms, err := req.Get("https://rms.api.bbc.co.uk/v2/services/bbc_radio_one/tracks/latest/playable")

	if err != nil {
		log.Error(err)
		return
	}

	err = rms.Unmarshal(tracks)
	if err != nil {
		log.Error(err)
		return
	}

	resp, err := http.Get("https://www.bbc.co.uk/sounds/play/live:bbc_radio_one")
	if err != nil {
		log.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}
	html := buf.String()

	jsonString := strings.Split(strings.Split(html, "window.__PRELOADED_STATE__ = ")[1], "};")[0] + "}"

	programmes := new(ProgrammesJSON)

	err = json.Unmarshal([]byte(jsonString), &programmes)
	if err != nil {
		log.Error(err)
		return
	}

	currentShow := &programmes.Programmes.Current
	currentShow.Timestamp, err = time.Parse("2006-01-02T15:04:05Z", currentShow.End)
	if err != nil {
		currentShow.Timestamp = time.Now()
	}

	var currentTrack *Track

	if len(tracks.Tracks) > 0 && tracks.Tracks[0].Availability.Label == "Now Playing" {
		currentTrack = &tracks.Tracks[0]
		currentTrack.Timestamp = time.Now()
	}

	var lastMessage snowflake.ID
	db := utils.GetDB()
	lastMessageID, _ := db.Get(lastMessageKey)
	lastShowBytes, _ := db.Get(lastShowKey)
	lastTrackBytes, _ := db.Get(lastTrackKey)

	var lastShow *Show
	if lastShowBytes != nil {
		lastShow = new(Show)
		_ = json.Unmarshal(lastShowBytes, &lastShow)
	}

	var lastTrack *Track
	if lastTrackBytes != nil {
		lastTrack = new(Track)
		_ = json.Unmarshal(lastTrackBytes, &lastTrack)
	}

	lastMessage, _ = snowflake.Parse(string(lastMessageID))

	if lastMessage != snowflake.ID(0) {
		msgBuilder := discord.NewMessageUpdateBuilder()

		if lastShow != nil && (currentShow == nil || lastShow.ID != currentShow.ID) {
			msgBuilder.AddEmbeds(lastShow.Embed("Ended"))
		}

		if lastTrack != nil && (currentTrack == nil || lastTrack.ID != currentTrack.ID) {
			msgBuilder.AddEmbeds(lastTrack.Embed("Played"))
		}

		if msgBuilder.Embeds != nil && len(*msgBuilder.Embeds) > 0 {
			_, err := client.Rest().UpdateMessage(channelID, lastMessage, msgBuilder.Build())
			if err != nil {
				log.Error(err)
			}
		} else if currentTrack != nil && lastTrack == nil {
			err := client.Rest().DeleteMessage(channelID, lastMessage)
			if err != nil {
				log.Error(err)
			}
		}
	}

	msgBuilder := discord.NewMessageCreateBuilder()

	if currentShow != nil {
		msgBuilder.AddEmbeds(currentShow.Embed("Until"))
	}

	if currentTrack != nil && (lastTrack == nil || currentTrack.ID != lastTrack.ID) {
		msgBuilder.AddEmbeds(currentTrack.Embed("Now Playing"))
	}

	if currentShow != nil && (lastShow == nil || currentShow.ID != lastShow.ID) ||
		currentTrack == nil && lastTrack != nil ||
		currentTrack != nil && lastTrack == nil ||
		currentTrack != nil && lastTrack != nil && currentTrack.ID != lastTrack.ID {
		npMessage, err := client.Rest().CreateMessage(channelID, msgBuilder.Build())

		if err != nil {
			log.Error(err)
			return
		}

		lastMessageID = []byte(npMessage.ID.String())
		_ = db.Put(lastMessageKey, lastMessageID)
	}

	if currentShow == nil || lastShow == nil || currentShow.ID != lastShow.ID {
		lastShowBytes, _ = json.Marshal(currentShow)
		_ = db.Put(lastShowKey, lastShowBytes)
	}

	if currentTrack == nil || lastTrack == nil || currentTrack.ID != lastTrack.ID {
		lastTrackBytes, _ = json.Marshal(currentTrack)
		_ = db.Put(lastTrackKey, lastTrackBytes)
	}
}

func (track *Track) Embed(footer string) discord.Embed {
	imageURL := ""

	if track.ImageURL != "" && track.ImageURL != "https://ichef.bbci.co.uk/images/ic/{recipe}/p0bqcdzf.jpg" {
		imageURL = strings.ReplaceAll(track.ImageURL, "{recipe}", "raw")
	}

	return discord.NewEmbedBuilder().
		SetColor(getColor(imageURL)).
		SetTitle(track.Titles.Primary).
		SetDescription(track.Titles.Secondary).
		SetImage(imageURL).
		SetFooterText(footer).
		SetTimestamp(track.Timestamp).
		Build()
}

func (show *Show) Embed(footer string) discord.Embed {
	title := "**ON AIR**"
	if footer == "Ended" {
		title = "**Previously**"
	}

	return discord.NewEmbedBuilder().
		SetColor(0x000001).
		SetAuthorName("BBC Radio 1").
		SetAuthorIcon("https://i.imgur.com/Q4DvfO8.png").
		SetTitle(title).
		AddField(
			"**"+show.Titles.Primary+"**",
			"**"+show.Titles.Secondary+"**\n"+show.Synopses.Short,
			false,
		).
		SetImage(strings.ReplaceAll(show.ImageURL, "{recipe}", "raw")).
		SetFooterText(footer).
		SetTimestamp(show.Timestamp).
		Build()
}

func getColor(imageURL string) int {
	if imageURL == "" {
		return 0
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		log.Error(err)
		return 0
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Error(err)
		return 0
	}

	palette, err := vibrant.NewPaletteFromImage(img)

	swatches := palette.ExtractAwesome()

	vib, ok := swatches["Vibrant"]
	if ok {
		return swatchToInt(*vib)
	}

	muted, ok := swatches["Muted"]
	if ok {
		return swatchToInt(*muted)
	}

	return 0
}

func swatchToInt(swatch vibrant.Swatch) int {
	hexString := strings.Replace(swatch.Color.RGBHex(), "#", "", -1)
	hex, err := strconv.ParseInt(hexString, 16, 64)
	if err == nil {
		return int(hex)
	}

	return 0
}
