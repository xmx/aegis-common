package validation

import (
	"strings"

	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type ValidError struct {
	unitran *ut.UniversalTranslator
	valid   validator.ValidationErrors
}

func (ve *ValidError) Error() string {
	fallback := ve.unitran.GetFallback()
	return ve.translate(fallback)
}

func (ve *ValidError) Translate(langs []string) string {
	var tran ut.Translator
	var found bool
	for _, lang := range langs {
		lang = strings.Replace(lang, "-", "_", -1)
		if tran, found = ve.unitran.GetTranslator(lang); found {
			break
		}
	}
	if tran == nil {
		tran = ve.unitran.GetFallback()
	}

	return ve.translate(tran)
}

func (ve *ValidError) translate(tran ut.Translator) string {
	trans := ve.valid.Translate(tran)
	causes := make([]string, 0, len(ve.valid))
	for _, err := range ve.valid {
		ns := err.Namespace()
		cause := trans[ns]
		causes = append(causes, cause)
	}

	return strings.Join(causes, ",")
}
