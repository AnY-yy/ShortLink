# ShortURL

一个基于 Go 构建的短链接服务项目，围绕短链生成、缓存加速、数据库持久化、布隆过滤器防穿透、雪花 ID 唯一标识生成等核心问题，完成了较完整的后端实现。

项目整体采用清晰的分层结构，业务层通过接口依赖仓储、缓存、短码生成器、雪花 ID 生成器和布隆过滤器实现，既保留了工程可维护性，也体现了你在 Go 项目组织、接口抽象和基础组件实现上的思考。

## 项目特色

- 基于 `Gin + GORM + MySQL + Redis` 实现短链服务核心链路
- 采用 `api / service / repo / cache / bootstrap / pkg` 分层设计
- 业务层面向接口编程，降低具体实现耦合
- 自定义实现雪花 ID、Base62、布隆过滤器、日志组件
- 具备“创建短链 -> 写库 -> 写缓存 -> 写布隆过滤器 -> 短链跳转解析”的完整流程
- 内置简单前端页面，可直接从浏览器创建短链

## 功能概览

### 已实现

- 长链接转短链接
- 长链接去重，避免重复创建
- 自定义短链标识
- 短链过期时间设置
- MySQL 持久化存储
- Redis 缓存加速查询
- 布隆过滤器防止缓存穿透
- 雪花 ID 生成全局唯一业务 ID
- Base62 编码生成短码
- Gin 请求日志中间件
- 首页模板渲染
- 短链重定向路由

### 当前业务链路

创建短链时：

1. API 层接收请求并校验参数
2. service 层检查长链接是否已存在
3. 若已存在，直接返回已有短链
4. 若不存在，生成雪花 ID
5. 用 Base62 将雪花 ID 编码为短码
6. 若用户指定自定义短链，则先校验唯一性
7. 处理过期时间
8. 写入 MySQL
9. 写入 Redis
10. 写入布隆过滤器
11. 返回短链结果

访问短链时：

1. 先查布隆过滤器，过滤绝对不存在的短链
2. 再查 Redis 缓存
3. 缓存未命中时回源 MySQL
4. 数据库命中后回填 Redis
5. 返回长链接并执行重定向

## 技术栈

### 后端

- Go
- Gin
- GORM
- MySQL
- Redis

### 配置与基础设施

- Viper
- Zap
- Lumberjack

### 校验与算法

- go-playground/validator
- Murmur3

### 前端

- Go HTML Template
- Tailwind CSS CDN
- Font Awesome

## 工程设计

### 1. 面向接口编程

在 `internal/service/service.go` 中，业务层定义并依赖以下接口：

- `Repository`
- `Cache`
- `SBloomFilter`
- `ShortCodeGenerator`
- `SnowFlakeGenerator`

这样做的好处：

- service 层更专注业务编排
- repo、cache、布隆过滤器等实现可以独立替换
- 更方便后续引入第三方实现或 mock 测试

### 2. 分层结构清晰

项目目录职责拆分明确：

- `internal/api`：HTTP 入口层
- `internal/service`：业务逻辑层
- `internal/repo`：数据库访问层
- `internal/cache`：缓存访问层
- `internal/bootstrap`：应用启动与组件装配
- `internal/router`：路由注册
- `pkg`：可复用基础组件
- `database`：数据库与 Redis 初始化、建表 SQL

### 3. 基础组件自实现

项目没有把所有能力都交给外部库，而是自己实现了几个关键基础组件：

- `pkg/snowflake`：雪花 ID 生成器
- `pkg/base62`：Base62 短码生成器
- `pkg/bloom`：自定义布隆过滤器 `SBloomFilter`
- `pkg/logger`：Zap 日志组件封装
- `pkg/jwt`：JWT 工具

这些组件不仅服务于业务，也体现了项目的学习深度和工程思路。

## 布隆过滤器实现亮点

项目中的布隆过滤器是手写实现的 `SBloomFilter`，而不是直接依赖成熟第三方库。

当前实现特点：

- 根据预期容量 `n` 和误判率 `p` 计算最优位图大小 `m`
- 根据公式计算最优哈希函数数量 `k`
- 基于 `murmur3.Sum128` 实现双哈希
- 使用 `big.Int` 作为底层位图
- 使用 `sync.RWMutex` 保证并发读写安全
- 通过空数据保护减少无效写入与查询

这部分很适合展示你对缓存穿透、概率型数据结构、并发安全和工程实现取舍的理解。

## 雪花 ID 与短码生成

短链接不是通过简单随机字符串拼接生成，而是使用：

- 自定义雪花 ID 生成唯一数值 ID
- Base62 将数值编码为更短、更可读的字符串

这种方案具备几个优点：

- 短码生成规则清晰
- 冲突概率低
- 适合高并发场景
- 便于后续做分布式扩展

## 日志体系

日志基于 Zap 进行了工程化封装，包含：

- JSON 编码
- 控制台与文件双写
- 基于 Lumberjack 的日志切割
- BufferedWriteSyncer 异步缓冲
- 日志采样
- Gin 请求日志中间件

说明项目除了功能实现，也在考虑可观测性和日志写入成本。

## 项目结构

```text
shortURL/
├── config/
│   ├── config.go
│   └── config.yml
├── database/
│   ├── db/
│   │   └── db.go
│   ├── migrate/
│   │   └── table.sql
│   └── rdb/
│       └── rdb.go
├── internal/
│   ├── api/
│   │   └── url.go
│   ├── bootstrap/
│   │   └── setup.go
│   ├── cache/
│   │   └── cache.go
│   ├── middleware/
│   │   └── logMiddle/
│   │       └── logger.go
│   ├── model/
│   │   ├── app.go
│   │   └── url.go
│   ├── repo/
│   │   └── repository.go
│   ├── router/
│   │   └── router.go
│   └── service/
│       └── service.go
├── log/
│   └── app.log
├── pkg/
│   ├── base62/
│   │   └── base62.go
│   ├── bloom/
│   │   └── bloom.go
│   ├── jwt/
│   │   └── jwt.go
│   ├── logger/
│   │   └── zap.go
│   └── snowflake/
│       └── snowflake.go
├── templates/
│   └── index.tmpl
├── test/
│   └── app.go
├── main.go
├── go.mod
└── README.md
```

## 核心模块说明

### `main.go`

- 应用入口
- 调用 `bootstrap.Setup()` 初始化组件
- 启动 Gin 服务

### `internal/bootstrap`

- 集中初始化配置、MySQL、Redis、Logger、SnowFlake、BloomFilter
- 将全局组件挂到 `Application` 上统一管理

### `internal/api`

- 负责接收 HTTP 请求
- 参数绑定与基础校验
- 调用 service 层完成业务处理

### `internal/service`

- 项目核心业务层
- 编排 repo、cache、snowflake、base62、bloom 等依赖
- 实现创建短链与解析短链两条主链路

### `internal/repo`

- 负责 MySQL 数据访问
- 处理短链和长链存在性查询
- 提供写库与回查能力

### `internal/cache`

- 负责 Redis 缓存写入与读取
- 在跳转查询时优先命中缓存

### `pkg`

- 沉淀独立可复用能力
- 既服务本项目，也方便后续拆分复用或替换实现

## 数据模型

当前核心表为 `urls`：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `BIGINT` | 雪花算法生成的全局唯一 ID |
| `longurl` | `TEXT` | 原始长链接 |
| `shorturl` | `VARCHAR(20)` | 系统实际使用的短链标识 |
| `selfshorturl` | `VARCHAR(20)` | 用户自定义短链标识 |
| `iscustom` | `BOOLEAN` | 是否为自定义短链 |
| `expiretime` | `TIMESTAMP` | 过期时间 |
| `createdtime` | `TIMESTAMP` | 创建时间 |

索引：

- `idx_shorturl`
- `idx_expiretime`

## 当前接口

### 1. 首页

```http
GET /
```

作用：

- 直接渲染短链创建页面

### 2. 首页别名

```http
GET /api/v1/index
```

作用：

- 在 API 分组下渲染同一个首页模板

### 3. 创建短链

```http
POST /api/v1/createurl
Content-Type: application/json
```

请求示例：

```json
{
  "longurl": "https://example.com/article/123",
  "selfshorturl": "mydemo",
  "expiretime": 24
}
```

字段说明：

- `longurl`：必填，合法 URL
- `selfshorturl`：可选，自定义短链，长度 4 到 10，仅允许字母数字
- `expiretime`：可选，单位小时，范围 `0 ~ 100`
- `expiretime = 0`：表示长期有效

成功响应示例：

```json
{
  "shorturl": "mydemo"
}
```

### 4. 短链重定向

```http
GET /:code
```

说明：

- 根据短码查询长链接
- 处理链路经过布隆过滤器、Redis、MySQL
- 命中后执行 `301` 重定向

## 参数校验

`CreateURLRequest` 当前通过 `validator` 进行约束：

- `longurl`：`required,url`
- `selfshorturl`：`omitempty,min=4,max=10,alphanum`
- `expiretime`：`omitempty,min=0,max=100`

## 配置说明

配置文件位于：

```text
config/config.yml
```

当前包含：

- MySQL 配置
- Redis 配置

运行前建议按本地环境修改：

- `Database.Host`
- `Database.Port`
- `Database.User`
- `Database.Password`
- `Database.DB`
- `Redis.HostPort`
- `Redis.Password`
- `Redis.DB`

## 快速启动

### 1. 克隆项目

```bash
git clone https://github.com/AnY-yy/shortURL.git
cd shortURL
```

### 2. 准备环境

确保本机已安装并启动：

- Go
- MySQL
- Redis

### 3. 修改配置

编辑：

```text
config/config.yml
```

### 4. 安装依赖

```bash
go mod tidy
```

### 5. 启动服务

```bash
go run main.go
```

默认监听端口：

```text
:8080
```

### 6. 访问页面

浏览器打开：

```text
http://localhost:8080/
```

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

### 创建结果

```json
{
  "shorturl": "demo123"
}
```

### 访问短链

```bash
curl -i http://localhost:8080/demo123
```

## 项目体现的能力点

这个项目能够比较集中地体现以下能力：

- Go 后端项目分层与目录组织能力
- 面向接口编程的设计意识
- MySQL 与 Redis 的组合使用能力
- 缓存穿透治理思路
- 布隆过滤器与雪花算法的落地实现能力
- API 设计与参数校验能力
- 日志体系与基础可观测性意识

## 后续可优化方向

- 增加短链访问次数统计
- 增加后台管理接口
- 补充统一响应结构与错误码
- 完善缓存空值和异常场景处理
- 增加单元测试与集成测试
- 将 JWT 正式接入用户体系
- 对比自实现布隆过滤器与第三方布隆过滤器的性能和误判率
- 增加布隆过滤器持久化或热加载方案

## 适合用于展示的场景

- Go 后端学习项目
- 简历项目
- 面试项目讲解
- GitHub 作品展示
- 基础组件实现与工程化思维展示

## License

MIT
