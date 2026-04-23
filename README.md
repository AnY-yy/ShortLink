# ShortLink 短链服务项目

## 项目介绍

这是一个基于 Go 语言开发的短链服务项目，旨在提供高效、稳定的 URL 缩短服务。项目采用 Redis 作为缓存层，MySQL 作为持久化存储，结合布隆过滤器等技术，实现高性能的短链生成和访问功能。

该项目适合刚学完 Redis 的开发者作为实践项目，涵盖了短链服务的核心功能，同时包含了一些创新功能，帮助开发者深入理解分布式系统中的常用技术。

## 技术栈

| 技术名称 | 版本要求 | 用途说明 |
|---------|---------|---------|
| Go | 1.20+ | 项目开发语言 |
| Redis | 6.0+ | 缓存层、布隆过滤器存储 |
| MySQL | 8.0+ | 持久化存储 |
| GORM | 最新版 | ORM 框架，简化数据库操作 |
| Gin | 最新版 | Web 框架，提供 HTTP 服务 |
| 布隆过滤器 | - | 防止缓存穿透，提升查询效率 |
| Docker | - | 容器化部署 |
| Nginx | - | 反向代理和负载均衡 |

## 环境要求

- Go 1.20+
- Redis 6.0+
- MySQL 8.0+
- Goland 2025.3（推荐 IDE）

## 基础功能

### 1. 长链生成短链
- 接收用户提交的长链接
- 生成唯一的短链标识
- 支持自定义短链后缀（可选）
- 返回短链访问地址

### 2. 短链访问重定向
- 接收短链访问请求
- 查询对应的长链接
- 实现 301/302 重定向
- 记录访问日志

### 3. Redis 缓存机制
- 短链信息缓存
- 热点数据预加载
- 缓存过期策略

### 4. 布隆过滤器
- 防止缓存穿透
- 快速判断短链是否存在
- 降低数据库查询压力

### 5. 雪花 ID 生成
- 分布式唯一 ID 生成
- 短链标识生成
- 支持高并发场景

### 6. 数据持久化
- 短链信息存储到 MySQL
- 访问日志记录
- 数据备份机制

### 7. 基础统计功能
- 短链访问次数统计
- 创建时间记录
- 最后访问时间记录

## 创新功能

### 1. 自定义短链后缀
- 允许用户自定义短链的后缀名称
- 支持后缀可用性检查
- 提供后缀规则校验

### 2. 短链有效期设置
- 支持设置短链的有效期
- 过期自动失效
- 到期提醒功能

### 3. 访问限流保护
- 防止恶意访问
- 基于用户 IP 的限流
- 基于短链的限流

### 4. 短链二维码生成
- 为短链生成二维码
- 支持自定义二维码样式
- 便于移动端访问

### 5. 批量生成短链
- 支持一次生成多个短链
- 批量导入长链接
- 批量导出结果

## 项目结构

```
shortURL/
├── main.go                 # 程序入口
├── config/                 # 配置文件
│   └── config.yaml
├── controller/             # 控制器层
│   ├── shortlink.go
│   └── redirect.go
├── service/                # 业务逻辑层
│   ├── shortlink_service.go
│   └── redirect_service.go
├── repository/             # 数据访问层
│   ├── shortlink_repo.go
│   └── redis_repo.go
├── model/                  # 数据模型
│   ├── shortlink.go
│   └── access_log.go
├── middleware/             # 中间件
│   ├── ratelimit.go
│   └── logger.go
├── utils/                  # 工具类
│   ├── snowflake.go
│   ├── bloomfilter.go
│   └── qrcode.go
├── router/                 # 路由配置
│   └── router.go
└── go.mod                  # 依赖管理
```

## 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/AnY-yy/shortURL.git
cd shortURL
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 配置环境
编辑 `config/config.yaml` 文件，配置数据库和 Redis 连接信息。

### 4. 启动服务
```bash
go run main.go
```

### 5. 测试接口
```bash
# 生成短链
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com/very/long/url"}'

# 访问短链
curl http://localhost:8080/abc123
```

## API 接口文档

### 生成短链
- **接口**: `POST /api/shorten`
- **参数**: `{"url": "长链接", "custom_suffix": "自定义后缀（可选）"}`
- **返回**: `{"short_url": "短链接", "code": 200}`

### 访问短链
- **接口**: `GET /:short_code`
- **参数**: `short_code`（短链编码）
- **返回**: 重定向到长链接

### 查询短链信息
- **接口**: `GET /api/info/:short_code`
- **参数**: `short_code`（短链编码）
- **返回**: `{"url": "长链接", "create_time": "创建时间", "access_count": 访问次数}`

## 学习目标

通过本项目，你将学习到：
- Go 语言 Web 开发基础
- Redis 缓存应用
- MySQL 数据库设计
- 布隆过滤器原理和应用
- 分布式 ID 生成算法
- RESTful API 设计
- 中间件开发
- 性能优化技巧

## 后续优化方向

- 支持短链统计分析
- 添加用户认证系统
- 实现短链分组管理
- 支持短链分享功能
- 添加监控和告警
- 实现 API 限流和鉴权

## 许可证

MIT License