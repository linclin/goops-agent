package global

import (
	"log/slog"

	ut "github.com/go-playground/universal-translator"
)

var (
	// 系统配置
	Conf Configuration
	// slog日志
	Log *slog.Logger
	// validation.v10相关翻译器
	Translator ut.Translator
)
