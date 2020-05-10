package login

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
	RoleAnonymous     Role = "anonymous" // anonymous/unauthenticated access to the site
	RoleVVGOMember    Role = "vvgo-member"
	RoleVVGOUploader  Role = "vvgo-uploader"
	RoleVVGODeveloper Role = "vvgo-developer"
)

var anonymous = Identity{
	Kind:  "",
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

// Role returns the first role or RoleAnonymous if the identity has no roles.
func (x Identity) Role() Role {
	if len(x.Roles) == 0 {
		return RoleAnonymous
	} else {
		return x.Roles[0]
	}
}

func (x Identity) HasRole(role Role) bool {
	for _, gotRole := range x.Roles {
		if gotRole == role {
			return true
		}
	}
	return false
}

func (x Identity) IsAnonymous() bool {
	return x.Role() == RoleAnonymous
}
