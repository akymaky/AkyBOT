package scrape

import (
	"encoding/binary"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/akymaky/akybot/config"
	"github.com/akymaky/akybot/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/imroc/req/v3"
	"io"
	"strconv"
	"strings"
	"time"
)

type Formats int

const (
	Moodle   Formats = 0
	HTML             = 1
	Plain            = 2
	Markdown         = 4
)

type moodlePost struct {
	ID            int                `json:"id"`
	Subject       string             `json:"subject"`
	Message       string             `json:"message"`
	MessageFormat Formats            `json:"messageformat"`
	Attachments   []moodleAttachment `json:"attachments"`
	DiscussionID  int                `json:"discussion"`
	TimeCreated   int64              `json:"created"`
	TimeModified  int64              `json:"modified"`
}

func (p *moodlePost) URL() string {
	return config.Get().Moodle.Url +
		"/mod/forum/discuss.php?d=" +
		strconv.Itoa(p.DiscussionID) +
		"#p" +
		strconv.Itoa(p.ID)
}

type moodleDiscussion struct {
	ID            int                `json:"discussion"`
	Name          string             `json:"name"`
	TimeCreated   int64              `json:"created"`
	TimeModified  int64              `json:"modified"`
	TimeLastPost  int64              `json:"timemodified"`
	PostID        int                `json:"id"`
	Subject       string             `json:"subject"`
	Message       string             `json:"message"`
	MessageFormat Formats            `json:"messageformat"`
	NumReplies    int                `json:"numreplies"`
	Attachments   []moodleAttachment `json:"attachments"`
}

func (d *moodleDiscussion) URL() string {
	return config.Get().Moodle.Url +
		"/mod/forum/discuss.php?d=" +
		strconv.Itoa(d.ID)
}

func (d *moodleDiscussion) Post() *moodlePost {
	return &moodlePost{
		ID:            d.PostID,
		Subject:       d.Subject,
		Message:       d.Message,
		MessageFormat: d.MessageFormat,
		Attachments:   d.Attachments,
		DiscussionID:  d.ID,
		TimeCreated:   d.TimeCreated,
		TimeModified:  d.TimeModified,
	}
}

type moodleAttachment struct {
	Name         string `json:"filename"`
	URL          string `json:"fileurl"`
	TimeModified int64  `json:"timemodified"`
}

type moodleGetDiscussionPosts struct {
	Posts []moodlePost `json:"posts"`
}

type moodleGetForumDiscussions struct {
	Discussions []moodleDiscussion `json:"discussions"`
}

func (p *moodlePost) CreateFiles() []*discord.File {
	var files []*discord.File

	for _, f := range p.Attachments {
		files = append(files, &discord.File{
			Name:   f.Name,
			Reader: DownloadMoodleFile(f.URL),
		})
	}

	return files
}

func (p *moodlePost) CreateEmbed(author string, footer string, updateTime int64) discord.Embed {
	converter := md.NewConverter("", true, nil)
	desc, err := converter.ConvertString(p.Message)
	if err != nil {
		log.Error(err)
	}

	if len(desc) > 4096 {
		desc = desc[:4094] + "…"
	}

	return discord.NewEmbedBuilder().
		SetAuthorName(author).
		SetColor(0xf27f22).
		SetTitle(p.Subject).
		SetURL(p.URL()).
		SetDescription(desc).
		SetFooterText(footer).
		SetTimestamp(time.Unix(updateTime, 0)).
		Build()
}

func (d *moodleDiscussion) CreateFiles() []*discord.File {
	return d.Post().CreateFiles()
}

func (d *moodleDiscussion) CreateEmbed(author string, footer string, updateTime int64) discord.Embed {
	return d.Post().CreateEmbed(author, footer, updateTime)
}

func wsClient() *req.Client {
	return noVerifyClient().
		AddCommonQueryParam("wstoken", config.Get().Moodle.Token).
		AddCommonQueryParam("moodlewsrestformat", "json").
		SetBaseURL(config.Get().Moodle.Token + "/webservice/rest/server.php")
}

func DownloadMoodleFile(fileURL string) io.Reader {
	moodle := noVerifyClient().R()
	moodle.
		AddQueryParam("token", config.Get().Moodle.Token)

	resp, err := moodle.Get(fileURL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return resp.Body
}

func csGetDiscussionPosts(discussionId int) []moodlePost {
	var result moodleGetDiscussionPosts

	moodle := wsClient().R().
		AddQueryParam("wsfunction", "mod_forum_get_forum_discussion_posts").
		AddQueryParam("discussionid", strconv.Itoa(discussionId)).
		AddQueryParams("sortdirection", "ASC").
		SetSuccessResult(&result)

	resp, err := moodle.Get("")
	if err != nil {
		log.Error(err)
		return make([]moodlePost, 0)
	}

	if !resp.IsSuccessState() {
		return make([]moodlePost, 0)
	}

	return result.Posts
}

func csGetDiscussions(bot bot.Client, forumId string, courseName string, roleMention string, channelId string) {
	key := []byte("cs." + forumId + ".lastUpdated")
	db := utils.GetDB()
	data, err := db.Get(key)
	if err != nil {
		data = make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0)
		db.Put(key, data)
	}

	lastUpdated := int64(binary.LittleEndian.Uint64(data))

	var result moodleGetForumDiscussions

	moodle := wsClient().R()
	moodle.
		AddQueryParam("wsfunction", "mod_forum_get_forum_discussions").
		AddQueryParam("perpage", "3").
		AddQueryParam("forumid", forumId).
		SetSuccessResult(&result)

	resp, err := moodle.Get("")
	if err != nil {
		log.Error(err)
		return
	}

	if !resp.IsSuccessState() {
		return
	}

	var discussions []moodleDiscussion

	for _, discussion := range result.Discussions {
		discussions = append([]moodleDiscussion{discussion}, discussions...)
	}

	var embeds []discord.Embed
	var files []*discord.File

	for _, discussion := range discussions {
		if discussion.TimeModified <= lastUpdated {
			continue
		}

		footer := "Измењена објава"
		updateTime := discussion.TimeModified

		if discussion.TimeCreated > lastUpdated {
			footer = "Нова објава"
			updateTime = discussion.TimeModified
		}

		embeds = append(embeds, discussion.CreateEmbed(courseName, footer, updateTime))
		files = append(files, discussion.CreateFiles()...)

		if discussion.TimeLastPost > lastUpdated {
			posts := csGetDiscussionPosts(discussion.ID)

			for _, post := range posts {
				if post.TimeCreated <= discussion.TimeCreated || post.TimeModified <= lastUpdated {
					continue
				}

				footer := "Измењен одговор"
				updateTime := post.TimeModified
				if post.TimeCreated > lastUpdated {
					footer = "Нов одговор"
					updateTime = post.TimeCreated
				}

				embeds = append(embeds, post.CreateEmbed(courseName, footer, updateTime))
				files = append(files, post.CreateFiles()...)
			}
		}
	}

	if len(embeds) == 0 {
		return
	}

	_, err = bot.Rest().CreateMessage(
		snowflake.MustParse(channelId),
		discord.NewMessageCreateBuilder().
			SetContent(roleMention).
			AddEmbeds(embeds...).
			AddFiles(files...).
			Build(),
	)

	if err != nil {
		log.Error(err)
		return
	}

	data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(time.Now().Unix()))
	db.Put(key, data)
}

func CS(bot bot.Client) {
	moodle := config.Get().Moodle

	for _, course := range moodle.Courses {
		embedAuthor := moodle.Embed.Author
		if len(course.Embed.Author) > 0 {
			embedAuthor = course.Embed.Author
		}

		channelId := moodle.ChannelId
		if course.ChannelId != 0 {
			channelId = course.ChannelId
		}

		courseName := strings.ReplaceAll(embedAuthor, "%course_name%", course.Name)

		csGetDiscussions(bot, strconv.Itoa(course.ForumId), courseName, "<@&"+strconv.FormatInt(course.RoleId, 10)+">", strconv.FormatInt(channelId, 10))
	}
}
