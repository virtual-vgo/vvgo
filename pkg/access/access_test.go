package access

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
	type fields struct {
		Kind  Kind
		Roles []Role
	}
	type args struct {
		role Role
	}
	for _, tt := range []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "success",
			fields: fields{Roles: []Role{"Tester"}},
			args:   args{"Tester"},
			want:   true,
		},
		{
			name:   "failure",
			fields: fields{Roles: []Role{"Cheater"}},
			args:   args{"Tester"},
			want:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := Identity{
				Kind:  tt.fields.Kind,
				Roles: tt.fields.Roles,
			}
			if got := x.HasRole(tt.args.role); got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentity_Role(t *testing.T) {
	type fields struct {
		Kind  Kind
		Roles []Role
	}
	tests := []struct {
		name   string
		fields fields
		want   Role
	}{
		{
			name:   "no roles",
			fields: fields{Roles: []Role{}},
			want:   RoleAnonymous,
		},
		{
			name:   "one role",
			fields: fields{Roles: []Role{"Tester"}},
			want:   "Tester",
		},
		{
			name:   "two roles",
			fields: fields{Roles: []Role{"Tester", "Cheater"}},
			want:   "Tester",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := Identity{
				Kind:  tt.fields.Kind,
				Roles: tt.fields.Roles,
			}
			if got := x.Role(); got != tt.want {
				t.Errorf("Role() = %v, want %v", got, tt.want)
			}
		})
	}
}
