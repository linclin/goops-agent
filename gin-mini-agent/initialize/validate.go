package initialize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	"gin-mini-agent/pkg/global"
)

// Validate 初始化校验器
//
// 该函数初始化请求参数校验器，配置错误信息的本地化翻译。
// 使用 validator.v10 作为校验引擎，支持多种语言翻译。
//
// 参数:
//   - locale: 语言代码，如 "zh"（中文）、"en"（英文）
//
// 功能特点:
//   - 自动提取 JSON tag 作为字段名
//   - 支持中文和英文错误信息翻译
//   - 与 Gin 框架集成
//
// 支持的语言:
//   - zh: 中文
//   - en: 英文
//   - 默认: 中文
//
// 使用示例:
//
//	// 初始化中文校验器
//	initialize.Validate("zh")
//
//	// 在请求处理中使用
//	type UserRequest struct {
//	    Name  string `json:"name" binding:"required"`
//	    Email string `json:"email" binding:"required,email"`
//	}
//
//	// 校验失败时获取翻译后的错误信息
//	err := c.ShouldBindJSON(&req)
//	if err != nil {
//	    errs := err.(validator.ValidationErrors)
//	    for _, e := range errs {
//	        fmt.Println(e.Translate(global.Translator))
//	    }
//	}
func Validate(locale string) {
	// 获取 Gin 框架的校验器引擎
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义字段名提取函数
		// 使用 JSON tag 作为字段名，使错误信息更友好
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// 提取 JSON tag 的第一个部分（去掉逗号后的选项）
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			// 忽略 "-" 标签
			if name == "-" {
				return ""
			}
			return name
		})

		// 创建语言翻译器
		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 创建通用翻译器
		// 第一个参数是备用语言
		// 后面的参数是支持的语言
		uni := ut.New(enT, zhT, enT)

		// 获取指定语言的翻译器
		var ok bool
		global.Translator, ok = uni.GetTranslator(locale)
		if !ok {
			global.Log.Error(fmt.Sprintf("初始化validator.v10校验器 uni.GetTranslator(%s) 失败", locale))
		}

		// 注册默认翻译
		var err error
		switch locale {
		case "en":
			// 注册英文翻译
			err = enTranslations.RegisterDefaultTranslations(v, global.Translator)
		case "zh":
			// 注册中文翻译
			err = zhTranslations.RegisterDefaultTranslations(v, global.Translator)
		default:
			// 默认使用中文翻译
			err = zhTranslations.RegisterDefaultTranslations(v, global.Translator)
		}
		if err != nil {
			global.Log.Error(fmt.Sprint("初始化validator.v10校验器失败", err))
		}
	}

	global.Log.Info("初始化validator.v10校验器完成")
}
