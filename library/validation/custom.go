package validation

import (
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func All() []CustomValidatorFunc {
	return []CustomValidatorFunc{
		uniqueFunc,
	}
}

func uniqueFunc() (string, validator.FuncCtx, validator.RegisterTranslationsFunc) {
	const tag = "unique"
	trans := func(utt ut.Translator) error {
		return utt.Add(tag, "{0}内数据不能重复", true)
	}

	return tag, nil, trans
}
