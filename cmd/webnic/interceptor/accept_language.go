package interceptor

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rafaeljusto/dnsmanager/Godeps/_workspace/src/github.com/gustavo-hms/trama"
	"github.com/rafaeljusto/dnsmanager/cmd/webnic/config"
)

type AcceptLanguage struct {
	trama.NopInterceptor
}

func NewAcceptLanguage() *AcceptLanguage {
	return new(AcceptLanguage)
}

func (i AcceptLanguage) Before(response trama.Response, r *http.Request) error {
	selectedLanguage := acceptLanguage(r)
	response.SetTemplateGroup(selectedLanguage)
	return nil
}

func acceptLanguage(r *http.Request) string {
	acceptLanguage := r.Header.Get("Accept-Language")
	acceptLanguageParts := strings.Split(acceptLanguage, ",")

	var selectedLanguage string
	var selectedQuality float64

	for _, part := range acceptLanguageParts {
		languageAndOptions := strings.Split(part, ";")

		language := languageAndOptions[0]
		var quality float64 = 1 // By default is quatility 100%

		for i := 1; i < len(languageAndOptions); i++ {
			option := languageAndOptions[i]
			optionParts := strings.Split(option, "=")

			if strings.ToUpper(optionParts[0]) == "Q" && len(optionParts) == 2 {
				var err error
				quality, err = strconv.ParseFloat(optionParts[1], 64)
				if err != nil {
					quality = 1
				}
			}
		}

		supported := false
		for _, supportedLanguage := range config.WebNIC.Templates.Languages {
			languageParts := strings.Split(language, "-")

			if strings.ToLower(language) == strings.ToLower(supportedLanguage) ||
				strings.ToLower(languageParts[0]) == strings.ToLower(supportedLanguage) {

				language = supportedLanguage
				supported = true
				break
			}
		}

		if supported && selectedQuality < quality {
			selectedLanguage = language
			selectedQuality = quality
		}
	}

	if selectedLanguage == "" && len(config.WebNIC.Templates.Languages) > 0 {
		selectedLanguage = config.WebNIC.Templates.Languages[0]
	}

	return selectedLanguage
}
