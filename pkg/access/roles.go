package access

type Role string

func (x Role) String() string { return string(x) }

const (
	RoleVVGOMember    Role = "vvgo-member"
)
