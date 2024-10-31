package radio

import "time"

type TracksJson struct {
	Schema string  `json:"$schema"`
	Total  int64   `json:"total"`
	Limit  int64   `json:"limit"`
	Offset int64   `json:"offset"`
	Tracks []Track `json:"data"`
}

type Track struct {
	Type           string        `json:"type"`
	ID             string        `json:"id"`
	Urn            string        `json:"urn"`
	Network        Network       `json:"network"`
	Titles         Titles        `json:"titles"`
	Synopses       interface{}   `json:"synopses"`
	ImageURL       string        `json:"image_url"`
	Duration       interface{}   `json:"duration"`
	Progress       interface{}   `json:"progress"`
	Container      interface{}   `json:"container"`
	Download       interface{}   `json:"download"`
	Availability   Availability  `json:"availability"`
	Release        Release       `json:"release"`
	Guidance       interface{}   `json:"guidance"`
	Activities     []interface{} `json:"activities"`
	Uris           []interface{} `json:"uris"`
	PlayContext    interface{}   `json:"play_context"`
	Recommendation interface{}   `json:"recommendation"`
	Timestamp      time.Time     `json:"timestamp"`
}

type Availability struct {
	From  interface{} `json:"from"`
	To    interface{} `json:"to"`
	Label string      `json:"label"`
}

type Release struct {
	Date  interface{} `json:"date"`
	Label interface{} `json:"label"`
}

type ProgrammesJSON struct {
	Programmes Programmes `json:"programmes"`
}

type Container struct {
	Type       ContainerType `json:"type"`
	ID         string        `json:"id"`
	Urn        string        `json:"urn"`
	Title      string        `json:"title"`
	Activities []interface{} `json:"activities"`
	Synopses   *Synopses     `json:"synopses,omitempty"`
}

type Synopses struct {
	Short  string `json:"short"`
	Medium string `json:"medium"`
	Long   string `json:"long"`
}

type DurationClass struct {
	Value int64  `json:"value"`
	Label string `json:"label"`
}

type Network struct {
	ID         PlaybackIdentifier `json:"id"`
	Key        Key                `json:"key"`
	ShortTitle ShortTitle         `json:"short_title"`
	LogoURL    string             `json:"logo_url"`
}

type Titles struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Tertiary  string `json:"tertiary"`
}

type Programmes struct {
	Current                  Show   `json:"current"`
	Next                     Show   `json:"next"`
	BroadcastPollingEnabled  bool   `json:"broadcastPollingEnabled"`
	BroadcastPollingInterval int64  `json:"broadcastPollingInterval"`
	BroadcastPollingFeed     string `json:"broadcastPollingFeed"`
}

type Show struct {
	ID           string             `json:"id"`
	Type         DatumType          `json:"type"`
	Urn          string             `json:"urn"`
	Network      Network            `json:"network"`
	Titles       Titles             `json:"titles"`
	Synopses     Synopses           `json:"synopses"`
	ImageURL     string             `json:"image_url"`
	Duration     int64              `json:"duration"`
	Container    Container          `json:"container"`
	Start        string             `json:"start"`
	End          string             `json:"end"`
	ServiceID    PlaybackIdentifier `json:"service_id"`
	PlayableItem interface{}        `json:"playable_item"`
	Timestamp    time.Time          `json:"timestamp"`
}

type PlaybackIdentifier string

const (
	BbcRadioOne PlaybackIdentifier = "bbc_radio_one"
)

type ContainerType string

const (
	Brand  ContainerType = "brand"
	Series ContainerType = "series"
)

type Key string

const (
	Radio1 Key = "radio1"
)

type ShortTitle string

const (
	ShortTitleRadio1 ShortTitle = "Radio 1"
)

type DatumType string

const (
	BroadcastSummary DatumType = "broadcast_summary"
	DisplayItem      DatumType = "display_item"
	PlayableItem     DatumType = "playable_item"
	SegmentItem      DatumType = "segment_item"
)

type Label string

const (
	Latest Label = "Latest"
	Link   Label = "Link"
)

type UrisType string

const (
	TypeLatest UrisType = "latest"
	TypeLink   UrisType = "link"
)
