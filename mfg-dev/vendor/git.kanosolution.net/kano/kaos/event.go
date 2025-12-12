package kaos

import (
	"net/http"
	"reflect"
	"time"

	"github.com/ariefdarmawan/byter"
	"github.com/sebarcode/codekit"
)

type PublishOpts struct {
	Headers codekit.M
	Config  codekit.M
	Timeout time.Duration
}

type EventHub interface {
	EventType() string
	SetPrefix(p string) EventHub
	Prefix() string
	SetSecret(secret string) EventHub
	Secret() string
	SetService(svc *Service)
	Service() *Service
	SetSignature(sign string) EventHub
	Signature() string
	SetByter(btr byter.Byter) EventHub
	Byter() byter.Byter
	Timeout() time.Duration
	SetTimeout(d time.Duration) EventHub
	Publish(name string, data interface{}, reply interface{}, opts *PublishOpts) error
	Unsubscribe(name string, model *ServiceModel)
	Subscribe(name string, model *ServiceModel, fn interface{}) error
	SubscribeEx(name string, model *ServiceModel, fn interface{}) error
	SubscribeExWithType(name string, model *ServiceModel, fn interface{}, reqType reflect.Type) error
	Close()
	Error() error
	SetDefaultOpts(opts *PublishOpts) EventHub
}

func (opts *PublishOpts) CopyDataFromCtx(ctx *Context) {
	kvs := ctx.Data().data
	for k, v := range kvs {
		opts.Headers.Set(k, v)
	}
}

func (opts *PublishOpts) CopyDataFromHTTPRequest(r *http.Request) {
	for k, v := range r.Header {
		switch len(v) {
		case 0:
			continue

		case 1:
			opts.Headers.Set(k, v[0])

		default:
			opts.Headers.Set(k, v)
		}
	}
}

func MergePublishOpts(opts *PublishOpts, others ...*PublishOpts) *PublishOpts {
	if opts == nil {
		opts = new(PublishOpts)
	}

	for _, oth := range others {
		if oth.Headers != nil {
			if opts.Headers == nil {
				opts.Headers = codekit.M{}
			}
			for k, v := range oth.Headers {
				opts.Headers.Set(k, v)
			}
		}

		if oth.Timeout > time.Duration(0) {
			opts.Timeout = oth.Timeout
		}

		if oth.Config != nil {
			if opts.Config == nil {
				opts.Config = codekit.M{}
			}
			for k, v := range oth.Config {
				opts.Config.Set(k, v)
			}
		}

		if oth.Timeout > time.Duration(0) {
			opts.Timeout = oth.Timeout
		}
	}

	return opts
}
