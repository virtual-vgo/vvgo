package login

import "strings"

// The kind of login
// This can be used to access additional metadata fields we might add for a particular login.
type Kind string

func (x Kind) String() string { return string(x) }

const (
	KindPassword Kind = "password"
	KindDiscord  Kind = "discord"
)

// A user role.
// Users can have multiple roles.
// These provide different levels of access to the api.
type Role string

func (x Role) String() string { return string(x) }

const (
	RoleAnonymous     Role = "anonymous" // anonymous/unauthenticated access to the site
	RoleVVGOMember    Role = "vvgo-member"
	RoleVVGOUploader  Role = "vvgo-uploader"
	RoleVVGODeveloper Role = "vvgo-developer"
)

// A user identity.
// This _absolutely_ should not contain any personally identifiable information.
// Numeric user id's are fine, but no emails, user names, addresses, etc.
type Identity struct {
	Kind  string `json:"kind" redis:"kind"`
	Roles string `json:"roles" redis:"roles"`
}

// Role returns the first role or RoleAnonymous if the identity has no roles.
func (x Identity) Role() Role {
	if len(x.Roles) == 0 {
		return RoleAnonymous
	} else {
		return Role(strings.Split(x.Roles, ":")[0])
	}
}

func (x Identity) HasRole(role Role) bool {

	for _, gotRole := range strings.Split(x.Roles, ":") {
		if gotRole == role.String() {
			return true
		}
	}
	return false
}
