package domain

import "context"

// RequestInfo carries transport-level forensic context (client IP, user-agent)
// captured at the HTTP boundary and read by the audit recorder. Separate from
// Caller, which is the auth/identity snapshot.
type RequestInfo struct {
	IP        string
	UserAgent string
}

type requestInfoKey struct{}

func WithRequestInfo(ctx context.Context, ri RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoKey{}, ri)
}

func RequestInfoFromCtx(ctx context.Context) (RequestInfo, bool) {
	ri, ok := ctx.Value(requestInfoKey{}).(RequestInfo)
	return ri, ok
}
