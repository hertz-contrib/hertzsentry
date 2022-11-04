# hertzsentry -> (This is a community driven project)

### Introduction

Sentry is an open source real-time error monitoring project that supports many sides, including the Web front end, server side, mobile side, and game side.In order to introduce sentry to `Hertz` on the basis of `Sentry-Go`, we implement the middleware hertzsentry, which provides some unified interfaces to help users get `sentry hub` and report error messages.

This project refers to <a  href ="https://github.com/gofiber/contrib/tree/main/fibersentry">fibersentry</a>.

### Install

```go
go get github.com/hertz-contrib/hertzsentry
```

### import

```go
import "github.com/hertz-contrib/hertzsentry"
```



### Config

hertzsentry provides a struct named `Option` for meeting your requirements. For exampple, you can use `WithRePanic` function to configure whether you want to block the request before moving forward with the response. 

```go
// WithRePanic configures whether Sentry should repanic after recovery.
// Set to true, if Recover middleware is used.
func WithRePanic(rePanic bool) Option {
	return Option{F: func(o *options) {
		o.rePanic = rePanic
	}}
}
```

### Usage

Hertzsentry attaches an instance of `*sentry.Hub` to the `*app. RequestContext` so that the hub can be used during the lifetime of the request. Concerning about concurrent scenarios, every request should maintain an unique instance of the `*sentry.Hub` for the security of hub. Therefore, We will clone a new instance of `sentry.CurrentHub` only when the request used it, so as to reduce the memory cost.

```go
package main

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/getsentry/sentry-go"
	"github.com/hertz-contrib/hertzsentry"
	"log"
)

var yourDsn = ""

func main()  {
	// set interval to 0 means using fs-watching mechanism.
	h := server.Default(server.WithAutoReloadRender(true, 0))

	// init sentry
	if err := sentry.Init(sentry.ClientOptions{
		// The DSN to use. If the DSN is not set, the client is effectively disabled.
		Dsn: yourDsn,
		// Before send callback.
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return event
		},
		// In debug mode, the debug information is printed to stdout to help you understand what
		// sentry is doing.
		Debug: true,
		// Configures whether SDK should generate and attach stacktraces to pure capture message calls.
		AttachStacktrace: true,
	}); err != nil {
		log.Fatal("sentry init failed")
	}

	// use sentry middleware and config with your requirements.
  // attention! you should use sentry handler after recovery.Recovery() 
	h.Use(hertzsentry.NewSentry(
		hertzsentry.WithSendRequest(true),
		hertzsentry.WithRePanic(true),
		))

	h.GET("/hello", func(c context.Context, ctx *app.RequestContext) {
		// use GetHubFromContext to get the hub
		if hub := hertzsentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("hertz", "CloudWeGo Hertz")
				scope.SetLevel(sentry.LevelDebug)
				hub.CaptureMessage("Just for debug")
			})
		}
		ctx.SetStatusCode(0)
	})

	h.Spin()
}
```

### Test 

Send a request to the interface `localhost:8888/hello`, then you can see the event in your sentry UI.

```sh
curl localhost:8888/hello
```



