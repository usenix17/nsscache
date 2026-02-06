package models

import (
	"fmt"
	"strings"
)

// User represents a passwd entry (RFC 2307 posixAccount)
type User struct {
	Name   string `json:"name"`   // uid
	Passwd string `json:"passwd"` // typically "x"
	UID    int    `json:"uid"`    // uidNumber
	GID    int    `json:"gid"`    // gidNumber
	GECOS  string `json:"gecos"`  // gecos/cn
	Dir    string `json:"dir"`    // homeDirectory
	Shell  string `json:"shell"`  // loginShell
}

// ToPasswdLine formats the user as a passwd file line
func (u *User) ToPasswdLine() string {
	return fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s",
		u.Name, u.Passwd, u.UID, u.GID, u.GECOS, u.Dir, u.Shell)
}

// Group represents a group entry (RFC 2307 posixGroup)
type Group struct {
	Name    string   `json:"name"`    // cn
	Passwd  string   `json:"passwd"`  // typically "x"
	GID     int      `json:"gid"`     // gidNumber
	Members []string `json:"members"` // memberUid
}

// ToGroupLine formats the group as a group file line
func (g *Group) ToGroupLine() string {
	return fmt.Sprintf("%s:%s:%d:%s",
		g.Name, g.Passwd, g.GID, strings.Join(g.Members, ","))
}

// Shadow represents a shadow entry (RFC 2307 shadowAccount)
type Shadow struct {
	Name     string `json:"name"`     // uid
	Passwd   string `json:"passwd"`   // userPassword (hashed) or "!!"
	LastChg  int    `json:"lstchg"`   // shadowLastChange
	Min      int    `json:"min"`      // shadowMin
	Max      int    `json:"max"`      // shadowMax
	Warn     int    `json:"warn"`     // shadowWarning
	Inactive int    `json:"inactive"` // shadowInactive
	Expire   int    `json:"expire"`   // shadowExpire
	Flag     int    `json:"flag"`     // shadowFlag
}

// ToShadowLine formats the shadow entry as a shadow file line
func (s *Shadow) ToShadowLine() string {
	// Use empty string for -1 values (unset)
	inactive := intToShadowField(s.Inactive)
	expire := intToShadowField(s.Expire)
	flag := intToShadowField(s.Flag)

	return fmt.Sprintf("%s:%s:%d:%d:%d:%d:%s:%s:%s",
		s.Name, s.Passwd, s.LastChg, s.Min, s.Max, s.Warn,
		inactive, expire, flag)
}

// intToShadowField converts an int to shadow field format
// -1 means unset (empty field)
func intToShadowField(v int) string {
	if v < 0 {
		return ""
	}
	return fmt.Sprintf("%d", v)
}
