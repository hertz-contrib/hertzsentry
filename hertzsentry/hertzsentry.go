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
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/getsentry/sentry-go"
)

const valuesKey = "sentry-hub"

// NewSentry the config of sentry adn return the handler of sentry middleware
func NewSentry(config ...Config) app.HandlerFunc {
	setSentryConfig(config...)
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
		if eventID != nil && configDefault.WaitForDelivery {
			hub.Flush(configDefault.Timeout)
		}
		if configDefault.RePanic {
			panic(err)
		}
	}
}

// GetHubFromContext get the sentry hub, every RequestContext shares the same hub instance
func GetHubFromContext(ctx *app.RequestContext) *sentry.Hub {
	cfg := configDefault
	fmt.Println(cfg)
	// get the existed hub
	if value, exist := ctx.Get(valuesKey); exist {
		if hub, ok := value.(*sentry.Hub); ok {
			return hub
		}
	}

	// new a cloned hub for this RequestContext
	hub := sentry.CurrentHub().Clone()

	// set request head for hub
	if request, err := adaptor.GetCompatRequest(&ctx.Request); err == nil && configDefault.SendHead {
		hub.Scope().SetRequest(request)
	}

	// set request body for hub
	if bytes, err := ctx.Body(); err == nil && bytes != nil && configDefault.SendBody {
		hub.Scope().SetRequestBody(bytes)
	}
	ctx.Set(valuesKey, hub)
	return hub
}
