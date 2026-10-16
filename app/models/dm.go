package models

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	MaxMessage     = 2000
	MaxChannelName = 24

	KindMsg    = "msg"
	KindJoin   = "join"
	KindLeave  = "leave"
	KindInvite = "invite"
	KindCreate = "create"
	KindNames  = "names"

	StateJoined  = "joined"
	StateLeft    = "left"
	StateInvited = "invited"

	Public  = "public"
	Private = "private"
)

var (
	ErrBadMessage   = errors.New("message must be 1-2000 characters")
	ErrSelfDM       = errors.New("you can't message yourself")
	ErrBadChannel   = errors.New("channel names are 1-24 characters: a-z, 0-9, - and _, starting with a letter or digit")
	ErrChannelTaken = errors.New("that channel already exists")
	ErrNotMember    = errors.New("you're not in this chat")
	ErrNotInvited   = errors.New("this channel is private. you need an invite to join")
	ErrAlreadyIn    = errors.New("they're already in this channel")
	ErrSelfInvite   = errors.New("you're already here")

	channelRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,23}$`)
)

type Group struct {
	ID            int64
	Name          string
	Visibility    string
	CreatedBy     string
	Members       int
	LastMessageAt time.Time
	MyState       string
}

type Invite struct {
	Group
	InvitedBy string
}

type Conversation struct {
	ID            int64
	Kind          string
	Name          string
	Other         string
	LastMessageAt time.Time
	LastBody      string
	LastSender    string
	LastKind      string
	LastTarget    string
	Unread        int
}

func (c Conversation) IsGroup() bool { return c.Kind == "group" }

type Message struct {
	ID        int64
	ConvID    int64
	SenderID  int64
	Sender    string
	Body      string
	Kind      string
	Target    string
	CreatedAt time.Time
	Mine      bool
}

func NormalizeChannel(name string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), "#"))
}

func ValidChannel(name string) bool { return channelRe.MatchString(name) }

func pair(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}

func (s *Store) UnreadMessages(ctx context.Context, userID int64) (int, error) {
	var n int
	err := s.DB.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM chat_members cm JOIN messages m ON m.conversation_id = cm.conversation_id
			WHERE cm.user_id = $1 AND cm.state = 'joined' AND m.kind = 'msg' AND m.sender_id <> $1 AND m.id > cm.last_read)
		+ (SELECT count(*) FROM chat_members WHERE user_id = $1 AND state = 'invited')`, userID).Scan(&n)
	return n, err
}

func (s *Store) Inbox(ctx context.Context, userID int64, limit, offset int) ([]Conversation, error) {
	rows, err := s.DB.Query(ctx, `SELECT c.id, c.kind, coalesce(c.name, ''), coalesce(ou.username, ''), c.last_message_at,
			coalesce(lm.body, ''), coalesce(lu.username, ''), coalesce(lm.kind, 'msg'), coalesce(lt.username, ''),
			(SELECT count(*) FROM messages m WHERE m.conversation_id = c.id AND m.kind = 'msg'
				AND m.sender_id <> $1 AND m.id > cm.last_read)
		FROM chat_members cm
		JOIN conversations c ON c.id = cm.conversation_id
		LEFT JOIN users ou ON c.kind = 'direct' AND ou.id = CASE WHEN c.user_a = $1 THEN c.user_b ELSE c.user_a END
		LEFT JOIN LATERAL (SELECT body, sender_id, kind, target_id FROM messages
			WHERE conversation_id = c.id ORDER BY id DESC LIMIT 1) lm ON true
		LEFT JOIN users lu ON lu.id = lm.sender_id
		LEFT JOIN users lt ON lt.id = lm.target_id
		WHERE cm.user_id = $1 AND cm.state = 'joined'
		ORDER BY c.last_message_at DESC, c.id DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Conversation, error) {
		var c Conversation
		err := r.Scan(&c.ID, &c.Kind, &c.Name, &c.Other, &c.LastMessageAt, &c.LastBody, &c.LastSender, &c.LastKind, &c.LastTarget, &c.Unread)
		return c, err
	})
}

func (s *Store) Invites(ctx context.Context, userID int64) ([]Invite, error) {
	rows, err := s.DB.Query(ctx, `SELECT c.id, c.name, c.visibility, coalesce(cu.username, ''),
			(SELECT count(*) FROM chat_members x WHERE x.conversation_id = c.id AND x.state = 'joined'),
			c.last_message_at, cm.state, coalesce(iu.username, '')
		FROM chat_members cm
		JOIN conversations c ON c.id = cm.conversation_id
		LEFT JOIN users cu ON cu.id = c.created_by
		LEFT JOIN users iu ON iu.id = cm.invited_by
		WHERE cm.user_id = $1 AND cm.state = 'invited'
		ORDER BY cm.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Invite, error) {
		var i Invite
		err := r.Scan(&i.ID, &i.Name, &i.Visibility, &i.CreatedBy, &i.Members, &i.LastMessageAt, &i.MyState, &i.InvitedBy)
		return i, err
	})
}

const groupCols = `c.id, c.name, c.visibility, coalesce(cu.username, ''),
	(SELECT count(*) FROM chat_members x WHERE x.conversation_id = c.id AND x.state = 'joined'),
	c.last_message_at,
	coalesce((SELECT state FROM chat_members me WHERE me.conversation_id = c.id AND me.user_id = $1), '')
	FROM conversations c LEFT JOIN users cu ON cu.id = c.created_by`

func (s *Store) queryGroups(ctx context.Context, sql string, args ...any) ([]Group, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Group, error) {
		var g Group
		err := r.Scan(&g.ID, &g.Name, &g.Visibility, &g.CreatedBy, &g.Members, &g.LastMessageAt, &g.MyState)
		return g, err
	})
}

func (s *Store) GroupByName(ctx context.Context, viewer int64, name string) (*Group, error) {
	gs, err := s.queryGroups(ctx, `SELECT `+groupCols+` WHERE c.kind = 'group' AND c.name = $2`, viewer, NormalizeChannel(name))
	if err != nil {
		return nil, err
	}
	if len(gs) == 0 {
		return nil, ErrNotFound
	}
	return &gs[0], nil
}

func (s *Store) PublicGroups(ctx context.Context, viewer int64, limit, offset int) ([]Group, error) {
	return s.queryGroups(ctx, `SELECT `+groupCols+` WHERE c.kind = 'group' AND c.visibility = 'public'
		ORDER BY c.last_message_at DESC, c.id DESC LIMIT $2 OFFSET $3`, viewer, limit, offset)
}

func (s *Store) Members(ctx context.Context, convID int64) ([]string, error) {
	rows, err := s.DB.Query(ctx, `SELECT u.username FROM chat_members cm JOIN users u ON u.id = cm.user_id
		WHERE cm.conversation_id = $1 AND cm.state = 'joined' ORDER BY u.username`, convID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (s *Store) DirectConversation(ctx context.Context, userID, otherID int64) (int64, string, error) {
	a, b := pair(userID, otherID)
	var id int64
	var state string
	err := s.DB.QueryRow(ctx, `SELECT c.id, coalesce(cm.state, '') FROM conversations c
		LEFT JOIN chat_members cm ON cm.conversation_id = c.id AND cm.user_id = $3
		WHERE c.kind = 'direct' AND c.user_a = $1 AND c.user_b = $2`, a, b, userID).Scan(&id, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", nil
	}
	return id, state, err
}

const messageCols = `m.id, m.conversation_id, m.sender_id, u.username, m.body, m.kind, coalesce(tu.username, ''),
	m.created_at, m.sender_id = $2
	FROM messages m JOIN users u ON u.id = m.sender_id LEFT JOIN users tu ON tu.id = m.target_id`

func (s *Store) queryMessages(ctx context.Context, sql string, args ...any) ([]Message, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Message, error) {
		var m Message
		err := r.Scan(&m.ID, &m.ConvID, &m.SenderID, &m.Sender, &m.Body, &m.Kind, &m.Target, &m.CreatedAt, &m.Mine)
		return m, err
	})
}

func (s *Store) Messages(ctx context.Context, convID, viewer int64, limit, offset int) ([]Message, error) {
	return s.queryMessages(ctx, `SELECT `+messageCols+`
		WHERE m.conversation_id = $1 ORDER BY m.id DESC LIMIT $3 OFFSET $4`, convID, viewer, limit, offset)
}

func (s *Store) MessagesAfter(ctx context.Context, convID, viewer, afterID int64, limit int) ([]Message, error) {
	return s.queryMessages(ctx, `SELECT `+messageCols+`
		WHERE m.conversation_id = $1 AND m.id > $3 ORDER BY m.id LIMIT $4`, convID, viewer, afterID, limit)
}

func (s *Store) MarkRead(ctx context.Context, convID, userID int64) error {
	_, err := s.DB.Exec(ctx, `UPDATE chat_members
		SET last_read = coalesce((SELECT max(id) FROM messages WHERE conversation_id = $1), 0)
		WHERE conversation_id = $1 AND user_id = $2`, convID, userID)
	return err
}

func insertMessage(ctx context.Context, tx pgx.Tx, convID, senderID int64, kind, body string, target *int64) (Message, error) {
	m := Message{ConvID: convID, SenderID: senderID, Kind: kind, Body: body}
	err := tx.QueryRow(ctx, `WITH ins AS (
			INSERT INTO messages (conversation_id, sender_id, body, kind, target_id) VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at)
		SELECT ins.id, ins.created_at, (SELECT username FROM users WHERE id = $2),
			coalesce((SELECT username FROM users WHERE id = $5), '') FROM ins`,
		convID, senderID, body, kind, target).Scan(&m.ID, &m.CreatedAt, &m.Sender, &m.Target)
	if err != nil {
		return m, err
	}
	_, err = tx.Exec(ctx, `UPDATE conversations SET last_message_at = now() WHERE id = $1`, convID)
	return m, err
}

func memberState(ctx context.Context, tx pgx.Tx, convID, userID int64) (string, error) {
	var state string
	err := tx.QueryRow(ctx, `SELECT state FROM chat_members WHERE conversation_id = $1 AND user_id = $2 FOR UPDATE`,
		convID, userID).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return state, err
}

func setState(ctx context.Context, tx pgx.Tx, convID, userID int64, state string, invitedBy *int64) error {
	_, err := tx.Exec(ctx, `INSERT INTO chat_members (conversation_id, user_id, state, invited_by) VALUES ($1, $2, $3, $4)
		ON CONFLICT (conversation_id, user_id) DO UPDATE SET state = $3, invited_by = coalesce($4, chat_members.invited_by), updated_at = now()`,
		convID, userID, state, invitedBy)
	return err
}

func (s *Store) SendDirect(ctx context.Context, userID, otherID int64, body string) ([]Message, error) {
	if userID == otherID {
		return nil, ErrSelfDM
	}
	body, ok := cleanText(body, MaxMessage)
	if !ok {
		return nil, ErrBadMessage
	}
	a, b := pair(userID, otherID)
	var out []Message
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO conversations (kind, user_a, user_b) VALUES ('direct', $1, $2)
			ON CONFLICT (user_a, user_b) DO UPDATE SET last_message_at = now() RETURNING id`, a, b).Scan(&id); err != nil {
			return err
		}
		for _, uid := range []int64{userID, otherID} {
			state, err := memberState(ctx, tx, id, uid)
			if err != nil {
				return err
			}
			if state == StateJoined {
				continue
			}
			if err := setState(ctx, tx, id, uid, StateJoined, nil); err != nil {
				return err
			}
			if state == "" {
				continue
			}
			ev, err := insertMessage(ctx, tx, id, uid, KindJoin, "", nil)
			if err != nil {
				return err
			}
			out = append(out, ev)
		}
		m, err := insertMessage(ctx, tx, id, userID, KindMsg, body, nil)
		if err != nil {
			return err
		}
		out = append(out, m)
		_, err = tx.Exec(ctx, `UPDATE chat_members SET last_read = $3 WHERE conversation_id = $1 AND user_id = $2`, id, userID, m.ID)
		return err
	})
	return out, err
}

func (s *Store) SendGroup(ctx context.Context, convID, userID int64, body string) (Message, error) {
	body, ok := cleanText(body, MaxMessage)
	if !ok {
		return Message{}, ErrBadMessage
	}
	var m Message
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		state, err := memberState(ctx, tx, convID, userID)
		if err != nil {
			return err
		}
		if state != StateJoined {
			return ErrNotMember
		}
		if m, err = insertMessage(ctx, tx, convID, userID, KindMsg, body, nil); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE chat_members SET last_read = $3 WHERE conversation_id = $1 AND user_id = $2`, convID, userID, m.ID)
		return err
	})
	return m, err
}

func (s *Store) Leave(ctx context.Context, convID, userID int64) (Message, error) {
	var m Message
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		state, err := memberState(ctx, tx, convID, userID)
		if err != nil {
			return err
		}
		if state != StateJoined {
			return ErrNotMember
		}
		if err := setState(ctx, tx, convID, userID, StateLeft, nil); err != nil {
			return err
		}
		m, err = insertMessage(ctx, tx, convID, userID, KindLeave, "", nil)
		return err
	})
	return m, err
}

func (s *Store) CreateGroup(ctx context.Context, userID int64, name, visibility string) (*Group, Message, error) {
	name = NormalizeChannel(name)
	if !ValidChannel(name) {
		return nil, Message{}, ErrBadChannel
	}
	if visibility != Public {
		visibility = Private
	}
	var id int64
	var m Message
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO conversations (kind, name, visibility, created_by) VALUES ('group', $1, $2, $3) RETURNING id`,
			name, visibility, userID).Scan(&id); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrChannelTaken
			}
			return err
		}
		if err := setState(ctx, tx, id, userID, StateJoined, nil); err != nil {
			return err
		}
		var err error
		m, err = insertMessage(ctx, tx, id, userID, KindCreate, visibility, nil)
		return err
	})
	if err != nil {
		return nil, Message{}, err
	}
	g, err := s.GroupByName(ctx, userID, name)
	return g, m, err
}

func (s *Store) Join(ctx context.Context, g *Group, userID int64) (Message, bool, error) {
	var m Message
	joined := false
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		state, err := memberState(ctx, tx, g.ID, userID)
		if err != nil {
			return err
		}
		if state == StateJoined {
			return nil
		}
		if g.Visibility == Private && state != StateInvited {
			return ErrNotInvited
		}
		if err := setState(ctx, tx, g.ID, userID, StateJoined, nil); err != nil {
			return err
		}
		if m, err = insertMessage(ctx, tx, g.ID, userID, KindJoin, "", nil); err != nil {
			return err
		}
		joined = true
		_, err = tx.Exec(ctx, `UPDATE chat_members SET last_read = $3 WHERE conversation_id = $1 AND user_id = $2`, g.ID, userID, m.ID)
		return err
	})
	return m, joined, err
}

func (s *Store) Invite(ctx context.Context, convID, inviterID, inviteeID int64) (Message, error) {
	if inviterID == inviteeID {
		return Message{}, ErrSelfInvite
	}
	var m Message
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		state, err := memberState(ctx, tx, convID, inviterID)
		if err != nil {
			return err
		}
		if state != StateJoined {
			return ErrNotMember
		}
		theirs, err := memberState(ctx, tx, convID, inviteeID)
		if err != nil {
			return err
		}
		if theirs == StateJoined {
			return ErrAlreadyIn
		}
		if err := setState(ctx, tx, convID, inviteeID, StateInvited, &inviterID); err != nil {
			return err
		}
		m, err = insertMessage(ctx, tx, convID, inviterID, KindInvite, "", &inviteeID)
		return err
	})
	return m, err
}

func (s *Store) PurgeMessages(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := s.DB.Exec(ctx, `DELETE FROM messages WHERE created_at < now() - $1::interval`, olderThan)
	return tag.RowsAffected(), err
}

func (s *Store) Decline(ctx context.Context, convID, userID int64) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM chat_members WHERE conversation_id = $1 AND user_id = $2 AND state = 'invited'`, convID, userID)
	return err
}
