/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

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
		// Before send callback.
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if hint.Context != nil {
				if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
					// You have access to the original Request
					fmt.Println(req)
				}
			}
			fmt.Println(event)
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

	// set interval to 0 means using fs-watching mechanism.
	h := server.Default(server.WithAutoReloadRender(true, 0))

	// use sentry middleware and config with your requirements.
	h.Use(hertzsentry.NewSentry(hertzsentry.Config{
		RePanic:         true,
		WaitForDelivery: false,
		SendHead:        true,
		SendBody:        false,
		Timeout:         0,
	}))

	h.GET("/bytedance", func(c context.Context, ctx *app.RequestContext) {
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
}
