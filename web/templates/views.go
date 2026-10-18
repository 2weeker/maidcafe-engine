package templates

import "maidcafe-engine/app/models"

const HoneypotField = "website"

type SiteSettings struct {
	RegistrationOpen bool
	About            bool
	AboutTitle       string
}

var Site = SiteSettings{RegistrationOpen: true}

type CaptchaView struct {
	ID            string
	Width, Height int
}

type AuthForm struct {
	Username string
	Honeypot bool
	Captcha  *CaptchaView
	Locked   string
}

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

type AdminUsersView struct {
	Users []models.AdminUser
	Query string
	Pager models.Pager
}

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

func SearchScopes(loggedIn bool) []string {
	if loggedIn {
		return []string{"all", "notes", "threads", "users", "tags"}
	}
	return []string{"all", "notes", "threads"}
}

func scopeLabel(scope string) string {
	switch scope {
	case "notes":
		return "✎ notes"
	case "threads":
		return ">> threads"
	case "users":
		return "~ users"
	case "tags":
		return "# tags"
	}
	return "* all"
}

func (v SearchView) Empty() bool {
	return len(v.Notes)+len(v.Posts)+len(v.Threads)+len(v.Users)+len(v.Tags) == 0
}

func searchURL(q, scope string) string {
	return withParam(withParam("/search", "q", urlQuery(q)), "in", scope)
}
