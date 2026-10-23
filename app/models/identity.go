package models

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"maidcafe-engine/app/tripcode"
)

const MaxName = 30

type Identity struct {
	Name string
	Trip string
}

func (id Identity) Anonymous() bool { return id.Trip == "" }

func CleanName(name string) string {
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '◆' || r == '◇' || r == '★' || r == '☆' {
			return -1
		}
		return r
	}, name)
	name = strings.Join(strings.Fields(name), " ")
	if utf8.RuneCountInString(name) > MaxName {
		name = string([]rune(name)[:MaxName])
	}
	return name
}

func NewIdentity(name, key string) Identity {
	key = strings.TrimSpace(key)
	if i := strings.IndexByte(name, '#'); i >= 0 {
		if key == "" {
			key = strings.TrimSpace(name[i+1:])
		}
		name = name[:i]
	}
	if key == "" {
		return Identity{}
	}
	return Identity{Name: CleanName(name), Trip: tripcode.Classic(key)}
}
