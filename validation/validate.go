package validation

import (
	"reflect"
	"strings"

	arlocale "github.com/go-playground/locales/ar"
	enlocale "github.com/go-playground/locales/en"
	eslocale "github.com/go-playground/locales/es"
	falocale "github.com/go-playground/locales/fa"
	frlocale "github.com/go-playground/locales/fr"
	frchlocale "github.com/go-playground/locales/fr_CH"
	rulocale "github.com/go-playground/locales/ru"
	zhlocale "github.com/go-playground/locales/zh"
	zhhanslocale "github.com/go-playground/locales/zh_Hans"
	zhhanttwlocate "github.com/go-playground/locales/zh_Hant_TW"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	artrans "github.com/go-playground/validator/v10/translations/ar"
	entrans "github.com/go-playground/validator/v10/translations/en"
	estrans "github.com/go-playground/validator/v10/translations/es"
	fatrans "github.com/go-playground/validator/v10/translations/fa"
	frtrans "github.com/go-playground/validator/v10/translations/fr"
	rutrans "github.com/go-playground/validator/v10/translations/ru"
	zhtrans "github.com/go-playground/validator/v10/translations/zh"
	zhtwtrans "github.com/go-playground/validator/v10/translations/zh_tw"
)

func New() *Validate {
	valid := validator.New()
	valid.RegisterTagNameFunc(jsonTag)

	arloc := arlocale.New()
	enloc := enlocale.New()
	esloc := eslocale.New()
	faloc := falocale.New()
	frloc := frlocale.New()
	frchloc := frchlocale.New()
	ruloc := rulocale.New()
	zhloc := zhlocale.New()
	zhhansloc := zhhanslocale.New()
	zhhanttwloc := zhhanttwlocate.New()

	unitran := ut.New(zhloc, arloc, enloc, esloc, faloc, frloc, frchloc, ruloc, zhloc, zhhansloc, zhhanttwloc)
	artran, _ := unitran.GetTranslator(arloc.Locale())
	entran, _ := unitran.GetTranslator(enloc.Locale())
	estran, _ := unitran.GetTranslator(esloc.Locale())
	fatran, _ := unitran.GetTranslator(faloc.Locale())
	frtran, _ := unitran.GetTranslator(frloc.Locale())
	frchtran, _ := unitran.GetTranslator(frchloc.Locale())
	rutran, _ := unitran.GetTranslator(ruloc.Locale())
	zhtran, _ := unitran.GetTranslator(zhloc.Locale())
	zhhanstran, _ := unitran.GetTranslator(zhhansloc.Locale())
	zhhanttwtran, _ := unitran.GetTranslator(zhhanttwloc.Locale())
	trans := []ut.Translator{
		artran, entran, estran, fatran, frtran, rutran, zhtran, zhhanstran, zhhanttwtran,
	}

	_ = artrans.RegisterDefaultTranslations(valid, artran)
	_ = entrans.RegisterDefaultTranslations(valid, entran)
	_ = estrans.RegisterDefaultTranslations(valid, estran)
	_ = fatrans.RegisterDefaultTranslations(valid, fatran)
	_ = frtrans.RegisterDefaultTranslations(valid, frtran)
	_ = frtrans.RegisterDefaultTranslations(valid, frchtran)
	_ = rutrans.RegisterDefaultTranslations(valid, rutran)
	_ = zhtrans.RegisterDefaultTranslations(valid, zhtran)
	_ = zhtrans.RegisterDefaultTranslations(valid, zhhanstran)
	_ = zhtwtrans.RegisterDefaultTranslations(valid, zhhanttwtran)

	return &Validate{
		valid:   valid,
		trans:   trans,
		unitran: unitran,
	}
}

type Validate struct {
	valid   *validator.Validate
	trans   []ut.Translator
	unitran *ut.UniversalTranslator
}

func (v *Validate) Validate(val any) error {
	err := v.valid.Struct(val)
	switch ve := err.(type) {
	case validator.ValidationErrors:
		return &ValidError{unitran: v.unitran, valid: ve}
	default:
		return err
	}
}

func (v *Validate) RegisterValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) error {
	return v.valid.RegisterValidationCtx(tag, fn, callValidationEvenIfNull...)
}

func (v *Validate) RegisterValidationTranslation(tag string, trans ut.Translator, registerFn validator.RegisterTranslationsFunc, translationFn validator.TranslationFunc) error {
	return v.valid.RegisterTranslation(tag, trans, registerFn, translationFn)
}

type CustomValidatorFunc func() (tag string, valid validator.FuncCtx, trans validator.RegisterTranslationsFunc, tranFunc validator.TranslationFunc)

func (v *Validate) RegisterCustomValidations(customs []CustomValidatorFunc) error {
	for _, custom := range customs {
		if err := v.RegisterCustomValidation(custom); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validate) RegisterCustomValidation(custom CustomValidatorFunc) error {
	if custom == nil {
		return nil
	}
	tag, validationFunc, regTranFunc, tranFunc := custom()
	if tag == "" {
		return nil
	}

	if validationFunc != nil {
		if err := v.RegisterValidationCtx(tag, validationFunc); err != nil {
			return err
		}
	}
	if regTranFunc != nil {
		if tranFunc == nil {
			tranFunc = v.defaultTranslation
		}
		for _, tran := range v.trans {
			if err := v.RegisterValidationTranslation(tag, tran, regTranFunc, tranFunc); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validate) RegisterStructValidationCtx(fn validator.StructLevelFuncCtx, types ...any) {
	v.valid.RegisterStructValidationCtx(fn, types...)
}

func (v *Validate) defaultTranslation(utt ut.Translator, fe validator.FieldError) string {
	str, _ := utt.T(fe.Tag(), fe.Field())
	return str
}

func jsonTag(f reflect.StructField) string {
	str := f.Tag.Get("json")
	cut, _, _ := strings.Cut(str, ",")
	if cut != "-" && cut != "" {
		return cut
	}

	return f.Name
}
