package interceptor

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gorilla/schema"
	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
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

func (p *Post) Before(response trama.Response, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	p.parse()

	request := p.handler.RequestValue()
	if !request.IsValid() {
		return
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := r.ParseForm(); err != nil {
		// TODO: response.ExecuteTemplate(...)
		return
	}

	if request.CanAddr() {
		request = request.Addr()
	}

	err := decoder.Decode(request.Interface(), r.Form)
	if err == nil {
		return
	}

	// TODO: Check druns project to identify errors
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
