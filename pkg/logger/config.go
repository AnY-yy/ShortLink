package logger

import (
	"errors"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志配置 编码器 决定输出的格式
// json输出格式 输出到file中的格式
func jsonEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()       // 获取生产环境的编码器配置
	config.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式
	config.EncodeLevel = zapcore.CapitalLevelEncoder // 日志级别大写
	config.EncodeCaller = zapcore.ShortCallerEncoder // 记录代码调用位置 文件名:行号 函数名

	// 返回json格式的编码器
	return zapcore.NewJSONEncoder(config)
}

// 日志配置 在文件中输出
func fileWriterSyncer() zapcore.WriteSyncer {
	lumber := &lumberjack.Logger{
		Filename:   "./log-info/log.log",
		MaxSize:    100,  // 单个文件最大大小 MB
		MaxAge:     7,    // 最大保留天数
		MaxBackups: 5,    // 最大备份文件数
		Compress:   true, // 是否压缩
	}

	// 返回一个写入到文件的写入器
	return zapcore.AddSync(lumber)
}

// 日志配置 在控制台输出
func consoleEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder // 日志级别大写 彩色输出
	config.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(config)
}

// 日志配置 在控制台输出 写入器
func consoleWriterSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

// 异步日志
// @param ws 日志写入器
// @return 异步日志写入器
func warpSyncCore(ws zapcore.WriteSyncer) *zapcore.BufferedWriteSyncer {
	syncWS := &zapcore.BufferedWriteSyncer{
		WS:            ws,
		Size:          1024 * 1024,
		FlushInterval: time.Second * 5, // 五秒刷新一次
		Clock:         zapcore.DefaultClock,
	}

	// 返回加入了异步日志写入器的写入器
	return syncWS
}

// 日志采样
func warpSampler(core zapcore.Core) zapcore.Core {
	// 如果日志突增100条 则只采样1条 1%
	return zapcore.NewSampler(core, time.Second, 100, 1)
}

// 组装核心
func newCore() zapcore.Core {
	// 日志级别
	level := zapcore.InfoLevel

	// 文件核心
	fileCore := zapcore.NewCore(jsonEncoder(), warpSyncCore(fileWriterSyncer()), level)

	// 控制台核心
	consoleCore := zapcore.NewCore(consoleEncoder(), warpSyncCore(consoleWriterSyncer()), level)

	// 核心核心 获得分别在文件和控制台输出的日志核心
	core := zapcore.NewTee(fileCore, consoleCore)

	return core
}

// generateLogger 生成日志实例
func generateLogger() *zap.Logger {
	// 日志核心
	core := newCore()
	return zap.New(
		core,
		zap.WrapCore(warpSampler), // 采样
		zap.AddCaller(),           // 记录代码调用位置
	)
}

// PingLogger
// 测试日志实例是否初始化成功
func PingLogger() error {
	logger := generateLogger()
	if logger == nil {
		return errors.New("日志实例初始化失败")
	}
	return nil
}
