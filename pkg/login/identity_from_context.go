package login

import (
	"context"
)

const CtxKeyVVGOIdentity = "vvgo_identity"

func IdentityFromContext(ctx context.Context) *Identity {
	ctxIdentity := ctx.Value(CtxKeyVVGOIdentity)
	identity, ok := ctxIdentity.(*Identity)
	if !ok {
		identity = new(Identity)
		*identity = Anonymous()
	}
	return identity
}
