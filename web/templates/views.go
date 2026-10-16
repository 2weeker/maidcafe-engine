package templates

import (
	"strconv"
	"time"

	"maidcafe-engine/app/models"
)

type HomeView struct {
	Boards     []models.Board
	MoreBoards bool
	Items      []models.FeedItem
	Sort       models.Sort
	Pager      models.Pager
	Draft      string
	Picker     TagPicker
	Users      []models.UserCard
	Tags       []models.Tag
}

type ProfileView struct {
	User  models.User
	Card  models.UserCard
	Items []models.FeedItem
	Sort  models.Sort
	Pager models.Pager
}

type UsersView struct {
	Users []models.UserCard
	Sort  models.UserSort
	Pager models.Pager
}

type FollowListView struct {
	User  models.User
	Kind  string
	Users []models.UserCard
	Pager models.Pager
}

type TagsView struct {
	Tags   []models.Tag
	Sort   models.TagSort
	Pager  models.Pager
	MaxLen int
	Draft  string
}

type TagView struct {
	Tag          models.Tag
	Show         string
	Sort         models.Sort
	Notes        []models.FeedItem
	Threads      []models.Thread
	Pager        models.Pager
	PostsPerPage int
}

type ChannelDraft struct {
	Name       string
	Visibility string
}

type InboxView struct {
	Conversations []models.Conversation
	Invites       []models.Invite
	Pager         models.Pager
	Draft         ChannelDraft
	Retention     int
}

type ChannelsView struct {
	Groups []models.Group
	Pager  models.Pager
}

type ChatView struct {
	Title     string
	Base      string
	Group     *models.Group
	Other     *models.User
	Members   []string
	Member    bool
	CanJoin   bool
	Invited   bool
	Left      bool
	CanLeave  bool
	Messages  []models.Message
	Pager     models.Pager
	Draft     string
	Retention int
}

func ChannelURL(name string) string { return "/dm/c/" + name }

func convURL(c models.Conversation) string {
	if c.IsGroup() {
		return ChannelURL(c.Name)
	}
	return DMURL(c.Other)
}

func lastMessageID(msgs []models.Message) string {
	if len(msgs) == 0 {
		return "0"
	}
	return strconv.FormatInt(msgs[len(msgs)-1].ID, 10)
}

func msgID(id int64) string    { return "m" + strconv.FormatInt(id, 10) }
func day(t time.Time) string   { return t.UTC().Format("2006-01-02") }
func clock(t time.Time) string { return t.UTC().Format("15:04") }

type SearchView struct {
	Query        string
	Scope        string
	Each         int
	Notes        []models.FeedItem
	Posts        []models.SearchPost
	Threads      []models.Thread
	Users        []models.UserCard
	Tags         []models.Tag
	Pager        models.Pager
	PostsPerPage int
}

func (v SearchView) Empty() bool {
	return len(v.Notes)+len(v.Posts)+len(v.Threads)+len(v.Users)+len(v.Tags) == 0
}

func searchURL(q, scope string) string {
	return withParam(withParam("/search", "q", urlQuery(q)), "in", scope)
}
