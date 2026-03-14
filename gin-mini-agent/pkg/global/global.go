package global

import (
	"log/slog"

	ut "github.com/go-playground/universal-translator"
)

// 全局变量
//
// 这些变量在应用程序启动时初始化，在整个应用程序生命周期中使用。
var (
	// Conf 系统配置
	// 存储从配置文件加载的所有配置信息
	// 在应用启动时通过 viper 加载
	//
	// 使用示例:
	//
	//	addr := global.Conf.RAG.Redis.Addr
	//	model := global.Conf.AiModel.ChatModel.Model
	Conf Configuration

	// Log slog 日志记录器
	// 结构化日志记录器，支持日志级别和格式化输出
	// 在应用启动时初始化
	//
	// 使用示例:
	//
	//	global.Log.Info("服务启动", "port", 8080)
	//	global.Log.Error("请求失败", "error", err)
	Log *slog.Logger

	// Translator validation.v10 翻译器
	// 用于将验证错误信息翻译为本地化语言
	// 支持中文、英文等多种语言
	//
	// 使用示例:
	//
	//	err := validate.Struct(user)
	//	if err != nil {
	//	    errs := err.(validator.ValidationErrors)
	//	    for _, e := range errs {
	//	        fmt.Println(e.Translate(Translator))
	//	    }
	//	}
	Translator ut.Translator
)
