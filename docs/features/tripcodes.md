# Names and tripcodes

In [textboard mode](textboard-mode.md) posts are anonymous. A poster who wants to be
recognisable across threads sets a **name and tripcode** once, on the `/name` page (`[name]`
in the nav), and it applies on every board.

## The rule

**A name is only shown together with a tripcode.** Without a tripcode key you post under the
board's default name, whatever name you typed. That is the point: a name alone could be typed
by anyone, so it proves nothing; a tripcode proves the poster knows a secret.

| On `/name` | You post as |
| --- | --- |
| nothing | the board's default name, `Anonymous` unless the board overrides it |
| name, no key | the same; the form refuses to save and says why |
| name + key | `name ◆WokonZwxw2` |
| key only | `Anonymous ◆WokonZwxw2` (the board's default name, with your code) |

The classic `name#key` form in the name field works too and is split at the first `#`.

## The tripcode

The code is the classic 2channel / futaba one, so a key you've used on another textboard gives
the same ten characters here:

```
salt  = characters 2-3 of (key + "H."), with anything outside ./0-9A-Za-z mapped as 2ch does
trip  = last 10 characters of crypt(3)(key, salt), traditional DES
```

`app/tripcode` is a pure-Go port of traditional DES `crypt(3)`, tested against the system's
`crypt` for a set of keys. Consequences of the classic algorithm:

- Only the **first 8 bytes** of the key count; `12345678` and `12345678extra` are the same code.
- Keys are case-sensitive. Non-ASCII keys are hashed as UTF-8 bytes, where 2ch uses Shift_JIS,
  so a Japanese key gives a different code than on 2ch.
- It's a weak hash by modern standards: a short key can be brute-forced to a wanted code. That's
  inherent to the format. Don't reuse a real password as a key.
- There are no "secure" (`##`, server-salted) tripcodes yet.

## Where it's stored

The name and key live in the `mc_name` cookie (`HttpOnly`, `SameSite=Lax`, `Secure` when
configured, one year), base64 of `name\nkey`, exactly like futaba's `namec` cookie. The server
recomputes the code at posting time and stores `name` and `trip` on the post; the key is never
stored. **clear** on `/name` deletes the cookie.

Someone who can read your cookies can post as you; so can anyone you share the key with.

## Names

Up to 30 characters after collapsing whitespace. Control characters and the marks `◆ ◇ ★ ☆` are
removed so a name can't fake a tripcode.

## Board default names

The default name is per board, in `config.yaml`:

```yaml
boards:
  defaults:
    anon_name: Anonymous
  overrides:
    meta:
      anon_name: Nameless Maid
```

The `/name` page lists the boards whose default differs from the site default. Threads in the
archive table, search results and the mod log use the same names.

## Display

A post header shows one of: `Anonymous` (or the board's name), `name ◆trip`, `Anonymous ◆trip`,
or for a staff member who ticked **sign as**, `~username ## mod`. The composer says *posting as
…* with a link to `/name`. In cafe mode none of this appears: every post is by an account.
