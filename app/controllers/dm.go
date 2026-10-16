package controllers

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/app/realtime"
	"maidcafe-engine/web/templates"
)

const sseHeartbeat = 25 * time.Second

func (ctl *Controller) Inbox(c fiber.Ctx) error {
	return ctl.renderInbox(c, fiber.StatusOK, page(c, "Messages"), templates.ChannelDraft{})
}

func (ctl *Controller) renderInbox(c fiber.Ctx, status int, p templates.Page, draft templates.ChannelDraft) error {
	me := middleware.CurrentUser(c)
	pg := pager(c, ctl.Config.Pagination.Users)
	convs, err := ctl.Store.Inbox(c.Context(), me.ID, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, convs); err != nil {
		return err
	}
	convs = models.Trim(&pg, convs)
	invites, err := ctl.Store.Invites(c.Context(), me.ID)
	if err != nil {
		return err
	}
	return render(c, status, templates.InboxPage(p, templates.InboxView{
		Conversations: convs,
		Invites:       invites,
		Pager:         pg,
		Draft:         draft,
		Retention:     ctl.Config.Chat.RetentionHours,
	}))
}

func (ctl *Controller) Channels(c fiber.Ctx) error {
	pg := pager(c, ctl.Config.Pagination.Users)
	groups, err := ctl.Store.PublicGroups(c.Context(), viewerID(c), pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, groups); err != nil {
		return err
	}
	groups = models.Trim(&pg, groups)
	return render(c, fiber.StatusOK, templates.ChannelsPage(page(c, "Channels"), templates.ChannelsView{Groups: groups, Pager: pg}))
}

func (ctl *Controller) CreateChannel(c fiber.Ctx) error {
	me := middleware.CurrentUser(c)
	draft := templates.ChannelDraft{Name: c.FormValue("name"), Visibility: c.FormValue("visibility")}
	g, ev, err := ctl.Store.CreateGroup(c.Context(), me.ID, draft.Name, draft.Visibility)
	if errors.Is(err, models.ErrBadChannel) || errors.Is(err, models.ErrChannelTaken) {
		p := page(c, "Messages")
		p.Error = err.Error()
		return ctl.renderInbox(c, fiber.StatusUnprocessableEntity, p, draft)
	}
	if err != nil {
		return err
	}
	ctl.Hub.Publish(realtime.ConvKey(g.ID), ev)
	return c.Redirect().To(templates.ChannelURL(g.Name))
}

type chatTarget struct {
	convID   int64
	key      string
	base     string
	title    string
	group    *models.Group
	other    *models.User
	member   bool
	canJoin  bool
	invited  bool
	left     bool
	canLeave bool
}

func (ctl *Controller) directTarget(c fiber.Ctx) (*chatTarget, error) {
	me := middleware.CurrentUser(c)
	other, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return nil, err
	}
	if other.ID == me.ID {
		return nil, fiber.NewError(fiber.StatusBadRequest, models.ErrSelfDM.Error())
	}
	id, state, err := ctl.Store.DirectConversation(c.Context(), me.ID, other.ID)
	if err != nil {
		return nil, err
	}
	return &chatTarget{
		convID:   id,
		key:      realtime.ConvKey(id),
		base:     templates.DMURL(other.Username),
		title:    "~" + other.Username,
		other:    other,
		member:   id == 0 || state == models.StateJoined,
		left:     state == models.StateLeft,
		canLeave: state == models.StateJoined,
	}, nil
}

func (ctl *Controller) groupTarget(c fiber.Ctx) (*chatTarget, error) {
	me := middleware.CurrentUser(c)
	g, err := ctl.Store.GroupByName(c.Context(), me.ID, c.Params("channel"))
	if err != nil {
		return nil, err
	}
	return &chatTarget{
		convID:   g.ID,
		key:      realtime.ConvKey(g.ID),
		base:     templates.ChannelURL(g.Name),
		title:    "#" + g.Name,
		group:    g,
		member:   g.MyState == models.StateJoined,
		canJoin:  g.MyState != models.StateJoined && (g.Visibility == models.Public || g.MyState == models.StateInvited),
		invited:  g.MyState == models.StateInvited,
		left:     g.MyState == models.StateLeft,
		canLeave: g.MyState == models.StateJoined,
	}, nil
}

func (ctl *Controller) Conversation(c fiber.Ctx) error {
	t, err := ctl.directTarget(c)
	if err != nil {
		return err
	}
	return ctl.renderChat(c, t, fiber.StatusOK, "", "")
}

func (ctl *Controller) Channel(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	return ctl.renderChat(c, t, fiber.StatusOK, "", "")
}

func (ctl *Controller) SendMessage(c fiber.Ctx) error {
	me := middleware.CurrentUser(c)
	t, err := ctl.directTarget(c)
	if err != nil {
		return err
	}
	msgs, err := ctl.Store.SendDirect(c.Context(), me.ID, t.other.ID, c.FormValue("body"))
	if err == nil && len(msgs) > 0 {
		t.key = realtime.ConvKey(msgs[0].ConvID)
		ctl.Hub.Publish(t.key, msgs...)
	}
	return ctl.afterSend(c, t, err)
}

func (ctl *Controller) SendChannel(c fiber.Ctx) error {
	me := middleware.CurrentUser(c)
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	m, err := ctl.Store.SendGroup(c.Context(), t.convID, me.ID, c.FormValue("body"))
	if err == nil {
		ctl.Hub.Publish(t.key, m)
	}
	return ctl.afterSend(c, t, err)
}

func (ctl *Controller) afterSend(c fiber.Ctx, t *chatTarget, err error) error {
	viaFetch := c.Get("X-Requested-With") == "fetch"
	switch {
	case errors.Is(err, models.ErrSelfDM):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrNotMember):
		if viaFetch {
			return c.Status(fiber.StatusForbidden).SendString(err.Error())
		}
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, models.ErrBadMessage):
		if viaFetch {
			return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
		}
		return ctl.renderChat(c, t, fiber.StatusUnprocessableEntity, err.Error(), c.FormValue("body"))
	case err != nil:
		return err
	}
	if viaFetch {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect().To(t.base + "#latest")
}

func (ctl *Controller) LeaveDirect(c fiber.Ctx) error {
	t, err := ctl.directTarget(c)
	if err != nil {
		return err
	}
	return ctl.leave(c, t, "/dm")
}

func (ctl *Controller) LeaveChannel(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	return ctl.leave(c, t, t.base)
}

func (ctl *Controller) leave(c fiber.Ctx, t *chatTarget, next string) error {
	if t.convID == 0 {
		return c.Redirect().To("/dm")
	}
	ev, err := ctl.Store.Leave(c.Context(), t.convID, middleware.CurrentUser(c).ID)
	if errors.Is(err, models.ErrNotMember) {
		return c.Redirect().To(next)
	}
	if err != nil {
		return err
	}
	ctl.Hub.Publish(t.key, ev)
	return c.Redirect().To(next)
}

func (ctl *Controller) JoinChannel(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	ev, joined, err := ctl.Store.Join(c.Context(), t.group, middleware.CurrentUser(c).ID)
	if errors.Is(err, models.ErrNotInvited) {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	if err != nil {
		return err
	}
	if joined {
		ctl.Hub.Publish(t.key, ev)
	}
	return c.Redirect().To(t.base + "#latest")
}

func (ctl *Controller) InviteToChannel(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	invitee, err := ctl.Store.UserByUsername(c.Context(), c.FormValue("username"))
	if errors.Is(err, models.ErrNotFound) {
		return ctl.renderChat(c, t, fiber.StatusUnprocessableEntity, "no such user: ~"+models.NormalizeUsername(c.FormValue("username")), "")
	}
	if err != nil {
		return err
	}
	ev, err := ctl.Store.Invite(c.Context(), t.convID, middleware.CurrentUser(c).ID, invitee.ID)
	switch {
	case errors.Is(err, models.ErrNotMember):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, models.ErrAlreadyIn), errors.Is(err, models.ErrSelfInvite):
		return ctl.renderChat(c, t, fiber.StatusUnprocessableEntity, err.Error(), "")
	case err != nil:
		return err
	}
	ctl.Hub.Publish(t.key, ev)
	return c.Redirect().To(t.base + "#latest")
}

func (ctl *Controller) DeclineInvite(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	if err := ctl.Store.Decline(c.Context(), t.convID, middleware.CurrentUser(c).ID); err != nil {
		return err
	}
	return c.Redirect().To("/dm")
}

func (ctl *Controller) renderChat(c fiber.Ctx, t *chatTarget, status int, errMsg, draft string) error {
	me := middleware.CurrentUser(c)
	v := templates.ChatView{
		Title:     t.title,
		Base:      t.base,
		Group:     t.group,
		Other:     t.other,
		Member:    t.member,
		CanJoin:   t.canJoin,
		Invited:   t.invited,
		Left:      t.left,
		CanLeave:  t.canLeave,
		Draft:     draft,
		Retention: ctl.Config.Chat.RetentionHours,
	}
	if t.member && t.convID != 0 {
		pg := pager(c, ctl.Config.Pagination.Messages)
		msgs, err := ctl.Store.Messages(c.Context(), t.convID, me.ID, pg.Limit(), pg.Offset())
		if err != nil {
			return err
		}
		if err := pastEnd(pg, msgs); err != nil {
			return err
		}
		msgs = models.Trim(&pg, msgs)
		slices.Reverse(msgs)
		v.Messages, v.Pager = msgs, pg
		if pg.Page == 1 {
			if err := ctl.Store.MarkRead(c.Context(), t.convID, me.ID); err != nil {
				return err
			}
		}
		if t.group != nil {
			if v.Members, err = ctl.Store.Members(c.Context(), t.convID); err != nil {
				return err
			}
		}
	} else {
		v.Pager = models.NewPager(1, 1)
	}
	p := page(c, "dm "+t.title)
	p.Error = errMsg
	return render(c, status, templates.ChatPage(p, v))
}

func (ctl *Controller) DirectEvents(c fiber.Ctx) error {
	t, err := ctl.directTarget(c)
	if err != nil {
		return err
	}
	return ctl.events(c, t)
}

func (ctl *Controller) ChannelEvents(c fiber.Ctx) error {
	t, err := ctl.groupTarget(c)
	if err != nil {
		return err
	}
	return ctl.events(c, t)
}

func (ctl *Controller) events(c fiber.Ctx, t *chatTarget) error {
	if !t.member {
		return fiber.NewError(fiber.StatusForbidden, models.ErrNotMember.Error())
	}
	me := middleware.CurrentUser(c)
	after := fiber.Query[int64](c, "after", 0)
	var lastSeen int64
	if _, err := fmt.Sscan(c.Get("Last-Event-ID"), &lastSeen); err == nil {
		after = max(after, lastSeen)
	}

	store, hub, meID, convID := ctl.Store, ctl.Hub, me.ID, t.convID
	key := t.key
	if convID == 0 {
		key = "pending:" + realtime.ConvKey(meID) + ":" + realtime.ConvKey(t.other.ID)
	}
	events, unsubscribe := hub.Subscribe(key)

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set("X-Accel-Buffering", "no")
	return c.SendStreamWriter(func(w *bufio.Writer) {
		defer unsubscribe()
		ctx := context.Background()

		fmt.Fprint(w, "retry: 3000\n\n")
		if convID != 0 {
			backlog, err := store.MessagesAfter(ctx, convID, meID, after, 200)
			if err != nil {
				log.Printf("chat events backlog: %v", err)
			}
			for _, m := range backlog {
				if writeMessageEvent(ctx, w, m) != nil {
					return
				}
			}
			if len(backlog) > 0 {
				store.MarkRead(ctx, convID, meID)
			}
		}
		if w.Flush() != nil {
			return
		}

		heartbeat := time.NewTicker(sseHeartbeat)
		defer heartbeat.Stop()
		for {
			select {
			case m := <-events:
				m.Mine = m.SenderID == meID
				if writeMessageEvent(ctx, w, m) != nil || w.Flush() != nil {
					return
				}
				if m.Kind == models.KindLeave && m.Mine {
					return
				}
				if !m.Mine && m.Kind == models.KindMsg {
					store.MarkRead(ctx, m.ConvID, meID)
				}
			case <-heartbeat.C:
				fmt.Fprint(w, ": ping\n\n")
				if w.Flush() != nil {
					return
				}
			case <-hub.Done():
				return
			}
		}
	})
}

func writeMessageEvent(ctx context.Context, w *bufio.Writer, m models.Message) error {
	var buf bytes.Buffer
	if err := templates.ChatLine(m).Render(ctx, &buf); err != nil {
		return err
	}
	fmt.Fprintf(w, "id: %d\nevent: msg\n", m.ID)
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}
