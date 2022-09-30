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

package hertzsentry

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/getsentry/sentry-go"
)

const valuesKey = "sentry-hub"

// the sentry config that used in the lifetime of sentry.
var sentryConfig Options

// NewSentry the config of sentry and return the handler of sentry middleware
func NewSentry(options ...Option) app.HandlerFunc {
	sentryConfig = NewOptions(options...)
	// return HandlerFunc for capturing events or messages when recovered
	return func(c context.Context, ctx *app.RequestContext) {
		defer recoverWithSentry(ctx)
		ctx.Next(c)
	}
}

func recoverWithSentry(ctx *app.RequestContext) {
	if err := recover(); err != nil {
		hub := GetHubFromContext(ctx)
		eventID := hub.RecoverWithContext(
			context.WithValue(context.Background(), sentry.RequestContextKey, ctx),
			err,
		)
		if eventID != nil && sentryConfig.WaitForDelivery {
			hub.Flush(sentryConfig.Timeout)
		}
		if sentryConfig.RePanic {
			panic(err)
		}
	}
}

// GetHubFromContext get the sentry hub, every RequestContext shares the same hub instance
func GetHubFromContext(ctx *app.RequestContext) *sentry.Hub {
	// get the existed hub
	if value, exist := ctx.Get(valuesKey); exist {
		if hub, ok := value.(*sentry.Hub); ok {
			return hub
		}
	}

	// new a cloned hub for this RequestContext
	hub := sentry.CurrentHub().Clone()

	// set request head for hub
	if request, err := adaptor.GetCompatRequest(&ctx.Request); err == nil && sentryConfig.SendRequest {
		hub.Scope().SetRequest(request)
	}

	// set request body for hub
	if bytes, err := ctx.Body(); err == nil && bytes != nil && sentryConfig.SendBody {
		hub.Scope().SetRequestBody(bytes)
	}
	ctx.Set(valuesKey, hub)
	return hub
}
