package routes

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/controllers"
	"maidcafe-engine/app/middleware"
)

func Register(app *fiber.App, ctl *controllers.Controller) {
	app.Get("/", ctl.Home)
	app.Post("/notes", middleware.RequirePoster, ctl.CreateNote)
	app.Get("/n/:id<int>", ctl.Note)
	app.Post("/n/:id<int>/reply", middleware.RequirePoster, ctl.ReplyNote)
	app.Get("/n/:id<int>/edit", middleware.RequirePoster, ctl.EditNoteForm)
	app.Post("/n/:id<int>/edit", middleware.RequirePoster, ctl.EditNote)
	app.Get("/n/:id<int>/delete", middleware.RequireUser, ctl.DeleteNoteForm)
	app.Post("/n/:id<int>/delete", middleware.RequireUser, ctl.DeleteNote)
	app.Post("/n/:id<int>/like", middleware.RequirePoster, ctl.LikeNote)
	app.Post("/n/:id<int>/repost", middleware.RequirePoster, ctl.RepostNote)

	app.Get("/register", middleware.RequireGuest, ctl.RegisterForm)
	app.Post("/register", middleware.RequireGuest, ctl.Register)
	app.Get("/login", middleware.RequireGuest, ctl.LoginForm)
	app.Post("/login", middleware.RequireGuest, ctl.Login)
	app.Post("/logout", ctl.Logout)
	app.Post("/theme", ctl.SetTheme)
	app.Get("/captcha/:file", ctl.CaptchaImage)
	app.Get("/about", ctl.About)

	app.Get("/u/:name", ctl.Profile)
	app.Get("/u/:name/followers", middleware.RequireUser, ctl.Followers)
	app.Get("/u/:name/following", middleware.RequireUser, ctl.FollowingList)
	app.Post("/u/:name/follow", middleware.RequirePoster, ctl.Follow)
	app.Post("/u/:name/role", middleware.RequireAdmin, ctl.SetRole)
	app.Post("/u/:name/ban", middleware.RequireMod, ctl.Ban)
	app.Post("/u/:name/unban", middleware.RequireMod, ctl.Unban)
	app.Get("/users", middleware.RequireUser, ctl.Users)
	app.Get("/tags", middleware.RequireUser, ctl.Tags)
	app.Post("/tags", middleware.RequirePoster, ctl.CreateTag)
	app.Get("/tags/:name", middleware.RequireUser, ctl.TagPage)
	app.Get("/search", ctl.Search)
	app.Get("/settings", middleware.RequireUser, ctl.Settings)
	app.Post("/settings", middleware.RequireUser, ctl.SaveSettings)
	app.Post("/settings/delete", middleware.RequireUser, ctl.DeleteAccount)
	app.Post("/settings/undelete", middleware.RequireUser, ctl.CancelDeletion)

	app.Get("/boards", ctl.BoardsPage)
	app.Get("/b/:slug", ctl.Board)
	app.Post("/b/:slug", middleware.RequirePoster, ctl.CreateThread)
	app.Get("/b/:slug/archive", ctl.Archive)
	app.Get("/t/:id<int>", ctl.Thread)
	app.Post("/t/:id<int>", middleware.RequirePoster, ctl.Reply)
	app.Get("/t/:id<int>/p/:no<int>/edit", middleware.RequirePoster, ctl.EditPostForm)
	app.Post("/t/:id<int>/p/:no<int>/edit", middleware.RequirePoster, ctl.EditPost)
	app.Get("/t/:id<int>/p/:no<int>/delete", middleware.RequireUser, ctl.DeletePostForm)
	app.Post("/t/:id<int>/p/:no<int>/delete", middleware.RequireUser, ctl.DeletePost)
	app.Post("/t/:id<int>/p/:no<int>/restore", middleware.RequireMod, ctl.RestorePost)
	app.Post("/t/:id<int>/lock", middleware.RequireMod, ctl.LockThread)
	app.Post("/t/:id<int>/pin", middleware.RequireMod, ctl.PinThread)
	app.Get("/t/:id<int>/delete", middleware.RequireMod, ctl.DeleteThreadForm)
	app.Post("/t/:id<int>/delete", middleware.RequireMod, ctl.DeleteThread)

	app.Get("/admin", middleware.RequireMod, ctl.AdminHome)
	app.Get("/admin/users", middleware.RequireMod, ctl.AdminUsers)
	app.Get("/admin/log", middleware.RequireMod, ctl.AdminLog)
	app.Get("/admin/boards", middleware.RequireAdmin, ctl.AdminBoards)
	app.Post("/admin/boards", middleware.RequireAdmin, ctl.CreateBoard)
	app.Post("/admin/boards/:slug", middleware.RequireAdmin, ctl.UpdateBoard)
	app.Get("/admin/config", middleware.RequireAdmin, ctl.AdminConfig)
}
