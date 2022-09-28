# hertzsentry -> (This is a community driven project)

### Introduction

Sentry is an open source real-time error monitoring project that supports many sides, including the Web front end, server side, mobile side, and game side.In order to introduce sentry to `Hertz` on the basis of `Sentry-Go`, we implement the middleware hertzsentry, which provides some unified interfaces to help users get `sentry hub` and report error messages.

### Installation

```go
go get github.com/getsentry/sentry-go/gin
```

```go
import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/getsentry/sentry-go"
	"hertzsentry/v1/hertzsentry"
	"log"
	"net/http"
)

func main() {
	// init sentry client
	if err := sentry.Init(sentry.ClientOptions{
		// The DSN to use. If the DSN is not set, the client is effectively disabled.
		Dsn: "your dsn",
	}); err != nil{
		log.Fatal("sentry init failed")
	}

	// set interval to 0 means using fs-watching mechanism.
	h := server.Default(server.WithAutoReloadRender(true, 0))

	// use sentry middleware and config with your requirements.
  h.Use(hertzsentry.NewSentry())

  // use sentry hub in handler.
	h.GET("/hertz", func(c context.Context, ctx *app.RequestContext) {
		if hub := hertzsentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("bytedance", "CloudWeGo for Bytedance")
				scope.SetLevel(sentry.LevelDebug)
				hub.CaptureMessage("Just for debug")
			})
		}
		ctx.SetStatusCode(0)
	})
  
	h.Spin()
}
```

### Configuration

hertzsentry provides a struct named `Config` for meeting your requirements. `Config` maintains the unique instance in sentry.

```go
type Config struct {
	// RePanic configures whether Sentry should repanic after recovery.
	RePanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	WaitForDelivery bool

	// SendHead configures whether you want to add current request head when capturing sentry events.
	SendHead bool

	// SendHead configures whether you want to add current request body when capturing sentry events.
	SendBody bool

	// Timeout for the event delivery requests.
	Timeout time.Duration
}
```

### Usage

Hertzsentry attaches an instance of `*sentry.Hub` to the `*app. RequestContext` so that the hub can be used during the lifetime of the request. Concerning about concurrent scenarios, every request should maintain an unique instance of the `*sentry.Hub` for the security of hub. Therefore, We will clone a new instance of `sentry.CurrentHub` only when the request used it, so as to reduce the memory cost.

```go
// set interval to 0 means using fs-watching mechanism.
h := server.Default(server.WithAutoReloadRender(true, 0))

// use sentry middleware and config with your requirements.
h.Use(hertzsentry.NewSentry(hertzsentry.Config{
  RePanic:         true,
  WaitForDelivery: false,
  SendHead: 		 true,
  SendBody: 		 false,
  Timeout:         0,
}))

h.GET("/bytedance", func(c context.Context, ctx *app.RequestContext) {
  // use GetHubFromContext to get the hub
  if hub := hertzsentry.GetHubFromContext(ctx); hub != nil {
    hub.WithScope(func(scope *sentry.Scope) {
      scope.SetTag("bytedance", "CloudWeGo for Bytedance")
      scope.SetLevel(sentry.LevelDebug)
      hub.CaptureMessage("Just for debug")
    })
  }
  ctx.SetStatusCode(0)
})

enhanceSentryEvent := func(c context.Context, ctx *app.RequestContext) {
  if hub := hertzsentry.GetHubFromContext(ctx); hub != nil {
    hub.Scope().SetTag("EnhanceSentryEvent Tag", "Maybe you need it")
  }
  ctx.Next(c)
}

h.GET("/hertz", enhanceSentryEvent, func(c context.Context, ctx *app.RequestContext) {
  panic("hertz")
})

h.Spin()
```



