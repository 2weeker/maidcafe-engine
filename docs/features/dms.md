# Messages: DMs and channels

Private chat between two people, or group channels, shown as a real-time IRC-style chat. This is
the one place the site uses JavaScript, and it degrades to plain forms without it.

## Two kinds of chat

| | Direct (`/dm/:name`) | Channel (`/dm/c/:channel`) |
| --- | --- | --- |
| Who | exactly two people | any number of members |
| Starts | `message` on a profile; created on the first message | `new channel` form on `/dm` |
| Finding it | your inbox | public ones are listed on `/dm/channels`; private ones only via invite |
| Joining | automatic | `join` button (public), or accept an invite (private) |

Channel names are 1-24 characters, `a-z 0-9 - _`, unique, shown as `#name`.

## Joining and leaving

Every chat has a `leave` button beside the composer. Joins and leaves appear in the log, IRC
style, and are pushed live to everyone in the chat:

```
[20:49] --> bob has joined
[20:50] <-- bob has left
[20:51] -!- carol invited dave
[20:52] *** alice created the channel (public)
```

- **Direct chats:** leaving hides the chat from your inbox and tells the other person. If either
  of you messages again, you're both back in, with a `-->` line.
- **Public channels:** anyone can join from the channel list or the channel page, and rejoin
  after leaving.
- **Private channels:** a member sends an invite from the `invite` button (a username). The
  invitee sees it at the top of their inbox with `join` / `decline`, and the `[dm N]` badge
  counts pending invites. Declining removes the invite; joining later needs a fresh one.
- Non-members can't read a channel, send to it, or open its live stream (403). The page shows
  why (`you're not in #x`, `#x is private`, `you left #x`) and a join button when joining is
  possible.
- A member list sits on the right of a channel (hidden on narrow screens) and updates live as
  people come and go.

## Messages vanish

`chat.retention_hours` in `config.yaml` (default 24; `0` keeps everything) sets how long
messages live. A sweep every minute deletes anything older, in direct chats and channels alike,
including join/leave lines. Chats and memberships stay; only the messages go. The inbox and
every chat window say so (`messages vanish after 24h`).

## The chat window

```
<< messages
┌────────────────────────────────────────────────────────┬──────────────┐
│ #tea  public :: 3 members :: messages vanish after 24h │ [live]       │
│ -- start of #tea --                                    │ 3 in #tea    │
│ [20:49] *** alice created the channel (public)         │ @alice       │
│ [20:49] --> bob has joined                             │ @bob         │
│ [20:49] <bob> tea time                                 │ @carol       │
│ [20:50] <alice> finally                                │              │
├────────────────────────────────────────────────────────┴──────────────┤
│ > [ say something (enter sends)        ] [send]      [invite] [leave] │
└───────────────────────────────────────────────────────────────────────┘
```

- Lines are left-aligned in the usual IRC way: `[HH:MM] <nick> message`, UTC, with
  `-- day changed --` separators. Your own nick is dimmed.
- Messages are 1-2000 characters and use the same [formatting](formatting.md) as notes.
- The composer is a single short line (Shift+Enter for a newline) with `send`, `invite` (channels
  only) and `leave` beside it.
- The newest `pagination.messages` (default 50) are shown; `-- older messages --` pages back.
  Older pages are read-only. Opening the latest page marks the chat read.

## Real time (the only JavaScript on the site)

`/static/js/chat.js` (about 120 lines, no dependencies) loads on the latest page of a
conversation only:

- **Receiving:** it opens an `EventSource` to `/dm/:name/events`. The server pushes each new
  message, pre-rendered as an IRC line, to both participants the moment it's sent. Lines are
  appended without reloading; the log keeps your scroll position unless you're already at the
  bottom.
- **Sending:** Enter (or the button) posts the form with `fetch`; Shift+Enter adds a newline.
  The box clears on success; errors appear in the log as `-!- …` lines. Your own message arrives
  through the same stream as everyone else's, so there's one way lines get drawn.
- **Status:** the bar shows `[live]` while connected and `[reconnecting]` while the browser
  retries (every 3 seconds). On reconnect the browser sends the last message id it saw, and the
  server replays everything newer, so nothing is missed.
- **Read state:** while the stream is open, incoming messages are marked read immediately, so the
  `[dm N]` badge doesn't count messages you're watching arrive.

**Without JavaScript** the same page still works: the form posts and the page reloads at the
newest message, and the bar shows a `[refresh]` link instead of `[live]`.

### How the stream works

- **Protocol:** Server-Sent Events, plain HTTP: `text/event-stream`, one `msg` event per message
  with the message id as the event id, and a comment ping every 25 seconds to keep proxies from
  closing idle connections.
- **Fan-out:** an in-process hub (`app/realtime`) keyed by conversation. Sending a message or a
  join/leave/invite event publishes it to every open stream for that chat (every member, any
  number of tabs). Leaving ends your own stream. A stream
  that falls more than 32 messages behind drops the extra ones instead of slowing the sender;
  they're stored, so they appear on the next page load or reconnect.
- **Shutdown:** stopping the app closes every stream first, so restarts aren't held up by open
  chats.
- **Limitation:** the hub lives in one process. Running several app instances would need a shared
  bus (Postgres `LISTEN/NOTIFY` would do) so a message sent to one instance reaches streams on
  another.

## Inbox (`/dm`)

Pending invites first, then every chat you're in (direct and channels), most recent activity
first: the name, the time, who sent the last message and its first 120 characters (or the last
join/leave). Unread chats have a thicker border, bold snippet and an `N new` tag. Below the list
is the `new channel` form. Paginated by `pagination.users`.

## Unread badge

When you're logged in, the nav shows `[dm]`, or `[dm 3]` inverted when you have unread messages.
The count is one small query per page load.

## Privacy and limits

- Only members can open a chat's log or stream; direct chats are 404 to anyone else and channels
  show a join prompt instead of the log.
- Messages are stored in plain text in the database (there's no end-to-end encryption; the
  server operator can read them).
- There is no deleting individual messages, blocking, muting, kicking, channel topics or
  attachments yet. See
  [roadmap.md](../roadmap.md).
