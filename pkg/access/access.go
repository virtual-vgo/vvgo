package access

// A user identity.
// This _absolutely_ should not contain any personally identifiable information.
// Numeric user id's are fine, but no emails, user names, addresses, etc.
type Identity struct {
	Kind  Kind   `json:"kind"`
	Roles []Role `json:"roles"`
}

// A user role.
// These provide different levels of access to the api.
type Role string

func (x Role) String() string { return string(x) }

const (
	RoleAnonymous     Role = "anonymous"
	RoleVVGOMember    Role = "vvgo-member"
	RoleVVGOUploader  Role = "vvgo-uploader"
	RoleVVGODeveloper Role = "vvgo-developer"
)

type Kind string

func (x Kind) String() string { return string(x) }

const (
	KindPassword Kind = "password"
	KindDiscord  Kind = "discord"
)

// returns the first role or RoleAnonymous if the identity has no roles.
func (x Identity) Role() Role {
	if len(x.Roles) == 0 {
		return RoleAnonymous
	} else {
		return x.Roles[0]
	}
}

// Returns true if this identity has the vvgo members role.
func (x Identity) IsVVGOMember() bool {
	for _, role := range x.Roles {
		if role == RoleVVGOMember {
			return true
		}
	}
	return false
}
