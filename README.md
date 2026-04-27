# ShortURL

一个基于 Go 实现的短链接服务，使用 `Gin + GORM + MySQL + Redis` 构建，包含短链生成、短链跳转、缓存加速、布隆过滤器防穿透、定时清理过期数据等核心能力。

当前项目同时提供：

- Web 页面入口：直接在浏览器中创建短链
- HTTP 接口：用于创建短链和访问短链
- 基础工程组件：日志、雪花 ID、Base62、布隆过滤器

## 功能特性

- 长链接转短链接
- 支持自定义短码
- 支持过期时间设置
- 相同长链接去重，避免重复生成
- Redis 缓存加速短链解析
- 自定义布隆过滤器防止缓存穿透
- 雪花 ID + Base62 生成短码
- 定时清理过期短链数据
- Gin 模板首页，可直接在页面中提交创建请求

## 技术栈

- Go 1.25
- Gin
- GORM
- MySQL
- Redis
- Viper
- Zap
- Lumberjack

## 核心流程

### 创建短链

1. 接收并校验 `longurl / selfshorturl / expiretime`
2. 检查长链接是否已存在
3. 若已存在，直接返回已有短链
4. 若不存在，生成雪花 ID
5. 使用 Base62 生成短码，或使用用户自定义短码
6. 写入 MySQL
7. 写入 Redis 缓存
8. 写入布隆过滤器

### 访问短链

1. 先经过布隆过滤器判断短码是否可能存在
2. 优先查询 Redis
3. 缓存未命中时回源 MySQL
4. 将结果重新写回 Redis
5. 返回 `301` 重定向

## 项目结构

```text
shortURL/
├── config/                 # 配置文件与配置加载
├── database/
│   ├── db/                 # MySQL 初始化
│   ├── migrate/            # 参考 SQL
│   └── rdb/                # Redis 初始化
├── internal/
│   ├── api/                # HTTP 处理层
│   ├── bootstrap/          # 应用启动装配
│   ├── cache/              # Redis 缓存访问
│   ├── middleware/         # 中间件
│   ├── model/              # 数据模型与请求结构
│   ├── repo/               # MySQL 数据访问
│   ├── router/             # 路由注册
│   ├── service/            # 核心业务逻辑
│   └── task/               # 定时任务
├── pkg/
│   ├── base62/             # Base62 编码
│   ├── bloom/              # 自定义布隆过滤器
│   ├── jwt/                # JWT 工具
│   ├── logger/             # 日志封装
│   └── snowflake/          # 雪花 ID
├── templates/              # 前端页面模板
├── test/                   # 简单测试代码
├── main.go
└── README.md
```

## 运行前准备

本项目启动时会初始化并连接以下依赖：

- MySQL
- Redis

请先确保本地服务可用，并按实际环境修改配置文件：

`config/config.yml`

```yml
Database:
  Host: 127.0.0.1
  Port: 3306
  User: root
  Password: your_password
  DB: shorturl

Redis:
  HostPort: 127.0.0.1:6379
  Password:
  DB: 0
```

## 快速启动

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务

```bash
go run main.go
```

服务默认监听：

```text
http://localhost:8080
```

### 3. 打开首页

浏览器访问：

```text
http://localhost:8080/
```

## 接口说明

### `GET /`

首页，渲染短链创建页面。

### `GET /api/v1/index`

返回同一个首页模板，作为页面入口别名。

### `POST /api/v1/createurl`

创建短链。

请求体：

```json
{
  "longurl": "https://example.com/article/123",
  "selfshorturl": "demo123",
  "expiretime": 12
}
```

字段说明：

- `longurl`：必填，合法 URL
- `selfshorturl`：可选，自定义短码，长度 4 到 10，仅允许字母和数字
- `expiretime`：可选，单位为小时，范围 `0 ~ 100`
- `expiretime = 0`：表示长期有效
- 未传 `expiretime`：当前实现默认 1 小时

说明：

- 当前实现返回的是 HTML 页面渲染结果，而不是标准 JSON API 响应
- 页面中的前端脚本已兼容这种返回方式

### `GET /:code`

根据短码跳转到原始长链接，成功时返回 `301 Moved Permanently`。

## 使用示例

### 创建短链

```bash
curl --location 'http://localhost:8080/api/v1/createurl' \
--header 'Content-Type: application/json' \
--data '{
  "longurl": "https://example.com/posts/short-url-demo",
  "selfshorturl": "demo123",
  "expiretime": 12
}'
```

### 访问短链

```bash
curl -i http://localhost:8080/demo123
```

## 定时任务

服务启动后会开启定时任务：

- 任务名：`cleanExpiredURL`
- 执行周期：每 5 分钟
- 作用：删除已过期的短链记录

## 亮点实现

### 自定义布隆过滤器

位于 `pkg/bloom`，基于 `murmur3` 与 `big.Int` 实现，支持：

- 按预期容量和误判率计算位图大小
- 按公式计算哈希函数个数
- 读写锁保证并发安全

### 雪花 ID + Base62

位于 `pkg/snowflake` 与 `pkg/base62`：

- 雪花 ID 提供全局唯一整数 ID
- Base62 将 ID 压缩为更短、可读性更好的短码

### 日志体系

位于 `pkg/logger`：

- 基于 Zap 封装
- 支持文件输出
- 支持 Gin 请求日志中间件

## 当前实现说明

- 首页通过 `templates/index.tmpl` 渲染
- `POST /api/v1/createurl` 由服务端直接回填页面结果
- MySQL 表结构在启动时通过 GORM `AutoMigrate` 自动迁移
- 仓库中保留了 `database/migrate/table.sql` 作为参考
- 项目中包含 `pkg/jwt`，但当前主流程未接入用户体系

## 后续可优化方向

- 补充统一 JSON 响应结构
- 增加短链访问次数统计
- 增加后台管理接口
- 为布隆过滤器增加持久化或预热方案
- 补充单元测试与集成测试
- 处理更多异常场景，例如缓存空值和数据库失效保护

## License

MIT
