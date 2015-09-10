package handler

import (
	"reflect"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/registrobr/trama"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy"
	handyinterceptor "github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/trajber/handy/interceptor"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/interceptor"
)

type defaultHandler struct {
	interceptor.PostCompliant
	uriVars handy.URIVars
	handyinterceptor.IntrospectorCompliant
}

func (d *defaultHandler) URIVars() handy.URIVars {
	return d.uriVars
}

type defaultChainHandler interface {
	Field(string, string) interface{}
	SetFields(handyinterceptor.StructFields)
	RequestValue() reflect.Value
	SetRequestValue(reflect.Value)
	URIVars() handy.URIVars
}

func defaultChain(h defaultChainHandler) trama.InterceptorChain {
	return trama.NewInterceptorChain(
		interceptor.NewURIVars(h),
		interceptor.NewPOST(h),
		interceptor.NewAcceptLanguage(),
	)
}
