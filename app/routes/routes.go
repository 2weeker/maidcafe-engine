package routes

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/controllers"
	"maidcafe-engine/app/middleware"
)

func Register(app *fiber.App, ctl *controllers.Controller) {
	app.Get("/", ctl.Home)
	app.Post("/notes", middleware.RequireUser, ctl.CreateNote)
	app.Get("/n/:id<int>", ctl.Note)
	app.Post("/n/:id<int>/reply", middleware.RequireUser, ctl.ReplyNote)
	app.Post("/n/:id<int>/like", middleware.RequireUser, ctl.LikeNote)
	app.Post("/n/:id<int>/repost", middleware.RequireUser, ctl.RepostNote)

	app.Get("/register", ctl.RegisterForm)
	app.Post("/register", ctl.Register)
	app.Get("/login", ctl.LoginForm)
	app.Post("/login", ctl.Login)
	app.Post("/logout", ctl.Logout)
	app.Post("/theme", ctl.SetTheme)

	app.Get("/u/:name", ctl.Profile)
	app.Get("/u/:name/followers", ctl.Followers)
	app.Get("/u/:name/following", ctl.FollowingList)
	app.Post("/u/:name/follow", middleware.RequireUser, ctl.Follow)
	app.Get("/users", ctl.Users)
	app.Get("/dm", middleware.RequireUser, ctl.Inbox)
	app.Get("/dm/channels", middleware.RequireUser, ctl.Channels)
	app.Post("/dm/channels", middleware.RequireUser, ctl.CreateChannel)
	app.Get("/dm/c/:channel", middleware.RequireUser, ctl.Channel)
	app.Post("/dm/c/:channel", middleware.RequireUser, ctl.SendChannel)
	app.Get("/dm/c/:channel/events", middleware.RequireUser, ctl.ChannelEvents)
	app.Post("/dm/c/:channel/join", middleware.RequireUser, ctl.JoinChannel)
	app.Post("/dm/c/:channel/leave", middleware.RequireUser, ctl.LeaveChannel)
	app.Post("/dm/c/:channel/invite", middleware.RequireUser, ctl.InviteToChannel)
	app.Post("/dm/c/:channel/decline", middleware.RequireUser, ctl.DeclineInvite)
	app.Get("/dm/:name", middleware.RequireUser, ctl.Conversation)
	app.Post("/dm/:name", middleware.RequireUser, ctl.SendMessage)
	app.Get("/dm/:name/events", middleware.RequireUser, ctl.DirectEvents)
	app.Post("/dm/:name/leave", middleware.RequireUser, ctl.LeaveDirect)
	app.Get("/tags", ctl.Tags)
	app.Post("/tags", middleware.RequireUser, ctl.CreateTag)
	app.Get("/tags/:name", ctl.TagPage)
	app.Get("/search", ctl.Search)
	app.Get("/settings", middleware.RequireUser, ctl.Settings)
	app.Post("/settings", middleware.RequireUser, ctl.SaveSettings)

	app.Get("/boards", ctl.BoardsPage)
	app.Get("/b/:slug", ctl.Board)
	app.Post("/b/:slug", middleware.RequireUser, ctl.CreateThread)
	app.Get("/b/:slug/archive", ctl.Archive)
	app.Get("/t/:id<int>", ctl.Thread)
	app.Post("/t/:id<int>", middleware.RequireUser, ctl.Reply)
}
