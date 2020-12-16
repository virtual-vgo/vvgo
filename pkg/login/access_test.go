package login

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/redis"
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

func TestAnonymous(t *testing.T) {
	assert.Equal(t, Anonymous(), anonymous)
}

func TestIdentity_IsAnonymous(t *testing.T) {
	t.Run("no roles", func(t *testing.T) {
		identity := Identity{Roles: []Role{}}
		assert.True(t, identity.IsAnonymous())
	})
	t.Run("anonymous role", func(t *testing.T) {
		identity := Identity{Roles: []Role{RoleAnonymous}}
		assert.True(t, identity.IsAnonymous())
	})
	t.Run("tester role", func(t *testing.T) {
		identity := Identity{Roles: []Role{"Tester"}}
		assert.False(t, identity.IsAnonymous())
	})
}

func TestIdentity_AssumeRoles(t *testing.T) {
	identity := Identity{Kind: "weenie", Roles: []Role{"flute", "piccolo", "clarinet"}}
	got := identity.AssumeRoles("flute", "piccolo", "tuba")
	want := Identity{Kind: "weenie", Roles: []Role{"flute", "piccolo"}}
	assert.Equal(t, want, got)
}
