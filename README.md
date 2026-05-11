# shortLink.v2

一个基于 Go 的短链接系统，采用分层架构与面向接口编程，支持短链创建、短链跳转、自定义短码、过期控制、缓存加速、布隆过滤器防穿透、定时清理与结构化日志。

## 版本说明

- GitHub 仓库 `main` 分支：项目第一版（v1）
- GitHub 仓库 `master` 分支：项目第二版（v2）
- 当前目录代码：v2

## 项目定位

本项目用于解决“长链接缩短 + 高并发跳转”场景，重点是：

- 创建短链时保证唯一性与可扩展性
- 跳转时尽量减少数据库压力
- 出错时可观测、可恢复、可维护

## 编码风格与架构思想

项目采用：

- 面向接口编程
- 依赖注入（统一在 `bootstrap` 装配）
- 分层解耦（API / Service / Repo / Cache / Middleware / Task）

核心收益：

- 业务层不依赖具体实现，方便替换 MySQL/Redis/日志组件
- 更容易单元测试（可 mock 接口）
- 后续扩展新功能时改动范围更可控

## 功能特性

- 短链创建
  - 输入长链接生成短链
  - 支持自定义短码（字母数字）
  - 支持过期时间（小时）
- 短链跳转
  - `GET /:code` 重定向到原始长链接
- 缓存加速
  - 创建后写入 Redis
  - 跳转优先走 Redis，未命中回源 MySQL 并回填缓存
- 防穿透
  - 使用自定义布隆过滤器拦截绝对不存在的短码
- 定时任务
  - 周期清理过期短链
- 可观测性
  - Web 请求日志
  - Panic 恢复中间件
  - 文件滚动日志

## 技术栈

- Go 1.25
- Gin
- GORM + MySQL
- Redis（go-redis/v9）
- Zap + Lumberjack
- YAML 配置
- Murmur3（布隆过滤器哈希）

## 第三方 API

### ipinfo.io

用途：请求日志中补充客户端 IP 的地理信息（国家、地区、城市、运营商等）。

配置示例（`config/config.yml`）：

```yaml
ipinfo:
  token: your_ipinfo_token
```

建议：

- 生产环境使用环境变量注入 token
- 增加超时、重试与本地缓存，避免外部服务抖动影响主链路

## 核心流程

### 1. 创建短链流程

1. API 接收请求并校验参数
2. Service 查询该长链接是否已存在短链
3. 若存在，直接返回已有短码
4. 若不存在：
   - 生成雪花 ID
   - 用户有自定义短码则校验冲突
   - 否则用 Base62 将雪花 ID 转短码
   - 写入 MySQL
   - 写入 Redis
   - 写入布隆过滤器
5. 返回短码

### 2. 跳转流程

1. 校验 `:code` 格式（字母数字）
2. 先查布隆过滤器
   - 不存在：直接返回无效（减少缓存/数据库压力）
3. 查 Redis
   - 命中：直接重定向
   - 未命中：查 MySQL
4. MySQL 命中后回填 Redis，再重定向

## 项目创新点

- 自定义实现布隆过滤器（位图 + Murmur3 双哈希）
- 雪花 ID + Base62 短码生成组合
- 结构化日志 + 文件滚动 + 采样
- 中间件化日志与 panic recovery
- 接口驱动的分层设计，便于维护与扩展

## API 接口说明

### 1) 获取首页

- 方法：`GET`
- 路径：`/api/v2/index`
- 说明：返回模板页面 `templates/index.tmpl`

### 2) 创建短链

- 方法：`POST`
- 路径：`/api/v2/createurl`
- `Content-Type`：`application/json`

请求体示例：

```json
{
  "longurl": "https://www.baidu.com",
  "selfshorturl": "baidu88",
  "expiretime": 24
}
```

字段说明：

- `longurl`：必填，合法 URL
- `selfshorturl`：可选，自定义短码（4~10 位字母数字）
- `expiretime`：可选，单位小时，范围 0~100
  - `0` 表示长期有效（实现中使用超长有效期）
  - 不传默认 1 小时

### 3) 短链跳转

- 方法：`GET`
- 路径：`/:code`
- 示例：`GET /baidu88`
- 行为：302 重定向

## 定时任务

当前任务：过期短链清理

- 启动后先执行一次
- 默认每 5 分钟执行一次
- 清理条件：`expireat < now`
- 单次任务带超时控制

## 项目结构

```text
shortLink.v2/
├─ main.go
├─ config/
├─ database/
│  ├─ db/
│  └─ rdb/
├─ internal/
│  ├─ api-v2/
│  ├─ bootstrap/
│  ├─ cache/
│  ├─ external/
│  ├─ middleware/
│  ├─ model/
│  ├─ repo/
│  ├─ router/
│  ├─ service/
│  └─ task/
├─ pkg/
│  ├─ base62/
│  ├─ bloomFilter/
│  ├─ logger/
│  └─ snowflake/
├─ templates/
└─ test/
```

## 快速启动

### 1) 准备依赖

- MySQL
- Redis
- Go 1.25+

### 2) 配置

编辑 `config/config.yml`：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  username: root
  password: your_password
  dbname: shorturl

redis:
  addr: 127.0.0.1:6379
  password:
  db: 0

ipinfo:
  token: your_ipinfo_token
```

### 3) 启动

```bash
go mod tidy
go run .
```

服务默认监听：`http://127.0.0.1:8080`

## 使用示例

### 创建随机短链

```bash
curl -X POST "http://127.0.0.1:8080/api/v2/createurl" \
  -H "Content-Type: application/json" \
  -d '{"longurl":"https://www.baidu.com","expiretime":12}'
```

### 创建自定义短链

```bash
curl -X POST "http://127.0.0.1:8080/api/v2/createurl" \
  -H "Content-Type: application/json" \
  -d '{"longurl":"https://www.baidu.com","selfshorturl":"baidu88","expiretime":24}'
```

### 访问短链

```bash
curl -I "http://127.0.0.1:8080/baidu88"
```

## 后期维护是否简便

总体维护成本较低，原因：

- 各层职责清晰，改动边界明确
- 依赖抽象接口，替换实现成本低
- 日志、任务、中间件已模块化

建议继续完善：

- 单元测试 + 集成测试 + 压测
- 统一错误码与 JSON API 响应协议
- 配置中心化与密钥脱敏
- 监控告警（Prometheus/Grafana）

## 后续更新方向

### 你提出的方向

1. 按 IP/用户维度生成个性化短码
- 同一个长链接（如 `https://www.baidu.com`）可为不同用户生成不同短码
- 每个用户可维护自己的短码集合

2. 自定义短码分用户隔离
- 不同用户可对同一长链接设置不同自定义短码
- 需要在数据模型中引入 `user_id` 或 `tenant_key`

3. 增加 goemail 反馈能力
- 用户可提交问题反馈
- 系统向管理员发送邮件通知
- 后台记录反馈处理状态

### 额外建议方向

- 访问统计：PV/UV、地区、设备、来源渠道
- 风控体系：限流、黑白名单、可疑行为拦截
- 用户系统：登录、鉴权、API Key、配额
- 批量能力：批量创建、导入导出
- 多域名支持：不同业务线绑定不同短链域名
- 管理后台：查询、失效、续期、审计
- 高可用：读写分离、Redis 高可用、多实例部署
- 安全增强：签名、防重放、敏感参数审计
- OpenAPI 文档与 SDK 自动生成

## License

可按开源计划补充（MIT / Apache-2.0 等）。
