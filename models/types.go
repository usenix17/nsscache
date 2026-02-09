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

