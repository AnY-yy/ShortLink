package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志配置 编码器 决定日志的格式
func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()       // 获取生产环境的编码器配置
	config.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式
	config.EncodeLevel = zapcore.CapitalLevelEncoder // 日志级别大写格式
	config.EncodeCaller = zapcore.ShortCallerEncoder // 记录代码调用位置 文件名:行号 函数名

	// 返回JSON格式的日志编码器
	return zapcore.NewJSONEncoder(config)
}

// 日志配置 决定日志的输出位置 console + file
func getWriterSyncer() zapcore.WriteSyncer {
	lumberJack := &lumberjack.Logger{
		Filename:   "./log/app.log", // 日志文件路径
		MaxSize:    100,             // 单个日志文件最大大小（MB）默认100MB
		MaxAge:     7,               // 日志文件最大保留天数，默认7天
		MaxBackups: 5,               // 最大备份文件数，默认5个
		Compress:   true,            // 是否压缩备份文件
	}

	// 返回一个同时写入控制台和文件的日志同步器
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJack))
}

// 异步日志 防止阻塞主程序执行
func warpAsyncCore(ws zapcore.WriteSyncer) *zapcore.BufferedWriteSyncer {
	// 直接将日志同步器包装成异步日志核心并返回
	asyncWS := &zapcore.BufferedWriteSyncer{
		WS:            ws,
		Size:          256 * 1024,           // 缓冲区大小 256KB
		FlushInterval: time.Second * 5,      // 每10秒刷新一次缓冲区
		Clock:         zapcore.DefaultClock, // 使用默认时钟
	}

	// 返回包装后的异步日志核心
	return asyncWS
}

// 日志采样 日志在1秒内不超过100条全部打印 若超过100条则只打印1% 防止日志暴增、磁盘IO爆炸
func warpSampler(core zapcore.Core) zapcore.Core {
	// 包装核心core 日志采样器
	return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
}

// 组装日志核心 日志采样器 异步日志核心 编码器 日志同步器
// 生产日志实例
func generateLogger() *zap.Logger {
	// 编码器
	encoder := getEncoder()
	// 日志同步器 包装成异步日志核心
	asyncWS := warpAsyncCore(getWriterSyncer())
	// 日志级别 INFO以上打印
	level := zapcore.InfoLevel

	// 组装日志核心
	core := zapcore.NewCore(encoder, asyncWS, level)

	// 生成日志实例 加入日志采样 显示文件与行数 Error级别自动打印堆栈 并返回日志实例
	return zap.New(
		core,
		zap.WrapCore(warpSampler),             // 采样
		zap.AddCaller(),                       // 显示文件与行数
		zap.AddStacktrace(zapcore.ErrorLevel), // Error级别自动打印堆栈
	)
	// 这个日志实例包含所有组件设置的所有功能
}

// NewLoger 初始化日志实例
func NewLoger() *zap.Logger {
	// 创建Logger
	logger := generateLogger()
	return logger
}
