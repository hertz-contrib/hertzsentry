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
	"time"
)

const valuesKey = "sentry-hub"

// Option defines the config for hertz sentry.
type Option struct {
	// RePanic configures whether Sentry should repanic after recovery.
	// Set to true, if Recover middleware is used.
	// Optional. Default: false
	RePanic bool

	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// If Recover middleware is used, it's safe to either skip this option or set it to false.
	// Optional. Default: false
	WaitForDelivery bool

	// SendHead configures whether you want to add current request head when capturing sentry events.
	// Optional. Default: false
	SendHead bool

	// SendHead configures whether you want to add current request body when capturing sentry events.
	// Optional. Default: false
	SendBody bool

	// Timeout for the event delivery requests.
	// Optional. Default: 2 Seconds
	Timeout time.Duration
}

// option default option for sentry client
var option = Option{
	RePanic:         false,
	WaitForDelivery: false,
	SendHead:        false,
	SendBody:        false,
	Timeout:         time.Second * 2,
}

// NewSentry the config of sentry adn return the handler of sentry middleware
func NewSentry(options ...Option) app.HandlerFunc {
	if len(options) > 1 {
		panic("the max length of options should not be larger than 1")
	} else if len(options) == 1 {
		option = options[0]
		if option.Timeout == 0 {
			option.Timeout = time.Second * 2
		}
	}
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
		if eventID != nil && option.WaitForDelivery {
			hub.Flush(option.Timeout)
		}
		if option.RePanic {
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
	if request, err := adaptor.GetCompatRequest(&ctx.Request); err == nil && option.SendHead {
		hub.Scope().SetRequest(request)
	}

	// set request body for hub
	if bytes, err := ctx.Body(); err == nil && bytes != nil && option.SendBody {
		hub.Scope().SetRequestBody(bytes)
	}
	ctx.Set(valuesKey, hub)
	return hub
}
