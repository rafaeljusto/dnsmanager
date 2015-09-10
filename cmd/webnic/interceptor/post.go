package interceptor

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gorilla/schema"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/registrobr/trama"
)

type poster interface {
	RequestValue() reflect.Value
	SetRequestValue(reflect.Value)
}

type Post struct {
	trama.NopInterceptor
	handler poster
}

func NewPOST(h poster) *Post {
	return &Post{handler: h}
}

func (p *Post) Before(response trama.Response, r *http.Request) error {
	if r.Method != "POST" {
		return nil
	}

	p.parse()

	request := p.handler.RequestValue()
	if !request.IsValid() {
		return nil
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := r.ParseForm(); err != nil {
		return err
	}

	if request.CanAddr() {
		request = request.Addr()
	}

	err := decoder.Decode(request.Interface(), r.Form)
	if err == nil {
		return nil
	}

	// TODO: Check druns project to identify errors
	return err
}

func (p *Post) parse() {
	st := reflect.ValueOf(p.handler).Elem()

	for j := 0; j < st.NumField(); j++ {
		field := st.Type().Field(j)

		value := field.Tag.Get("request")
		if value == "all" || strings.Contains(value, "post") {
			p.handler.SetRequestValue(st.Field(j))
			break
		}
	}
}

type PostCompliant struct {
	request reflect.Value
}

func (p *PostCompliant) RequestValue() reflect.Value {
	return p.request
}

func (p *PostCompliant) SetRequestValue(r reflect.Value) {
	p.request = r
}
