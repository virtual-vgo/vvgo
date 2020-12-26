package login

import (
	"fmt"
	"strings"
)

// The kind of login
// This can be used to access additional metadata fields we might add for a particular login.
type Kind string

func (x Kind) String() string { return string(x) }

const (
	KindPassword Kind = "password"
	KindBearer   Kind = "bearer"
	KindBasic    Kind = "basic"
	KindDiscord  Kind = "discord"
)

// A user role.
// Users can have multiple roles.
// These provide different levels of access to the api.
type Role string

func (x Role) String() string { return string(x) }

const (
	RoleAnonymous  Role = "anonymous"   // anonymous/unauthenticated access to the site
	RoleVVGOMember Role = "vvgo-member" // password login or has the vvgo-member discord role
	RoleVVGOTeams  Role = "vvgo-teams"  // has the teams discord role
	RoleVVGOLeader Role = "vvgo-leader" // has the leader discord role
)

var anonymous = Identity{
	Kind:  "anonymous",
	Roles: []Role{RoleAnonymous},
}

func Anonymous() Identity { return anonymous }

// A user identity.
// This _absolutely_ should not contain any personally identifiable information.
// Numeric user id's are fine, but no emails, user names, addresses, etc.
type Identity struct {
	Kind  Kind   `json:"kind"`
	Roles []Role `json:"roles"`
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
	var new []Role
	for _, role := range roles {
		if x.HasRole(role) {
			new = append(new, role)
		}
	}
	x.Roles = new
	return x
}
