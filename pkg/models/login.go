package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"strings"
	"time"
)

// Kind The kind of login.
// This can be used to access additional metadata fields we might add for a particular login.
type Kind string

func (x Kind) String() string { return string(x) }

const (
	KindPassword Kind = "password"
	KindBearer   Kind = "bearer"
	KindBasic    Kind = "basic"
	KindDiscord  Kind = "discord"
	KindApiToken Kind = "api_token"
)

// Role A user role.
// Users can have multiple roles.
// These provide different levels of access to the api.
type Role string

func (x Role) String() string { return string(x) }

const (
	RoleAnonymous             Role = "anonymous"   // anonymous/unauthenticated access to the site
	RoleVVGOVerifiedMember    Role = "vvgo-member" // password login or has the vvgo-member discord role
	RoleVVGOProductionTeam    Role = "vvgo-teams"  // has the teams discord role
	RoleVVGOExecutiveDirector Role = "vvgo-leader" // has the leader discord role

	RoleWriteSpreadsheet Role = "write_spreadsheet"
	RoleReadSpreadsheet  Role = "read_spreadsheet"
	RoleDownload         Role = "download"
)

var anonymous = Identity{
	Kind:  "anonymous",
	Roles: []Role{RoleAnonymous},
}

func Anonymous() Identity { return anonymous }

// Identity A user identity.
type Identity struct {
	Key       string
	Kind      Kind
	Roles     []Role
	ExpiresAt time.Time `json:"ExpiresAt,omitempty"`
	CreatedAt time.Time `json:"CreatedAt,omitempty"`
	DiscordID string    `json:"DiscordID,omitempty"`
}

func ListSessions(ctx context.Context, identity Identity) ([]Identity, error) {
	var keys []string
	redis.Do(ctx, redis.Cmd(&keys, "KEYS", "sessions:*"))

	sessionData := make([]string, 0, len(keys))
	redis.Do(ctx, redis.Cmd(&sessionData, "MGET", keys...))

	sessions := make([]Identity, len(keys))
	for i := range sessionData {
		json.Unmarshal([]byte(sessionData[i]), &sessions[i])
	}
	for i := range sessions {
		sessions[i].Key = strings.TrimPrefix(keys[i], "sessions:")
	}

	var want []Identity
	for _, session := range sessions {
		switch {
		case identity.HasRole(RoleVVGOExecutiveDirector):
			want = append(want, session)
		case session.DiscordID != "" && session.DiscordID == identity.DiscordID:
			want = append(want, session)
		}
	}
	return want, nil
}

func (x Identity) Info() string {
	roles := make([]string, len(x.Roles))
	for i, role := range x.Roles {
		roles[i] = string(role)
	}
	return fmt.Sprintf("kind: %s, roles: %s", x.Kind, strings.Join(roles, " "))
}

func (x Identity) HasRole(role Role) bool {
	if role == RoleAnonymous {
		return true
	}
	for _, gotRole := range x.Roles {
		if gotRole == role {
			return true
		}
	}
	return false
}

func (x Identity) IsAnonymous() bool {
	return len(x.Roles) == 0 || (len(x.Roles) == 1 && x.Roles[0] == RoleAnonymous)
}

func (x Identity) AssumeRoles(roles ...Role) Identity {
	var newRoles []Role
	for _, role := range roles {
		if x.HasRole(role) {
			newRoles = append(newRoles, role)
		}
	}
	x.Roles = newRoles
	return x
}
