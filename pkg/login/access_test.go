package login

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKind_String(t *testing.T) {
	assert.Equal(t, "Cheese", Kind("Cheese").String())
}

func TestRole_String(t *testing.T) {
	assert.Equal(t, "Cheese", Role("Cheese").String())
}

func TestIdentity_HasRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester"}}
		assert.True(t, identity.HasRole("Tester"))
	})
	t.Run("failure", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester"}}
		assert.False(t, identity.HasRole("Baker"))
	})
}

func TestIdentity_Role(t *testing.T) {
	t.Run("no roles", func(t *testing.T) {
		identity := Identity{Roles: []Role{}}
		assert.Equal(t, identity.Role(), RoleAnonymous)
	})
	t.Run("one role", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester"}}
		assert.Equal(t, identity.Role(), Role("Tester"))
	})
	t.Run("two roles", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester", "Baker"}}
		assert.Equal(t, identity.Role(), Role("Tester"))
	})
}

func TestAnonymous(t *testing.T) {
	assert.Equal(t, Anonymous(), anonymous)
}

func TestIdentity_IsAnonymous(t *testing.T) {
	t.Run("no roles", func(t *testing.T) {
		identity := Identity{Roles: []Role{}}
		assert.True(t, identity.IsAnonymous())
	})
	t.Run("one role", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester"}}
		assert.False(t, identity.IsAnonymous())
	})
}
