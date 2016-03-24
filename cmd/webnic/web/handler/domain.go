package handler

import (
	"net/http"
	"path"

	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/web/interceptor"
	"github.com/registrobr/trama"
)

func init() {
	Mux.Register("/", func() trama.Handler {
		return new(domain)
	})
}

type domain struct {
	trama.NopHandler
}

func (d *domain) Get(response trama.Response, r *http.Request) error {
	response.ExecuteTemplate("domain.tmpl.html", nil)
	return nil
}

func (d *domain) Templates() trama.TemplateGroupSet {
	groupSet := trama.NewTemplateGroupSet(nil)

	for _, language := range config.WebNIC.Templates.Languages {
		templatePath := path.Join(config.WebNIC.Home, config.WebNIC.Templates.Path, language, "domain.tmpl.html")
		groupSet.Insert(trama.TemplateGroup{
			Name:  language,
			Files: []string{templatePath},
		})
	}

	return groupSet
}

func (d *domain) Interceptors() trama.InterceptorChain {
	return trama.NewInterceptorChain(
		interceptor.NewAcceptLanguage(),
	)
}
