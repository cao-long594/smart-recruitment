# 实现说明与功能—技术栈对照

本文档说明本仓库「智能招聘系统」已实现的能力，以及每项能力主要用到的技术/组件。详细接口见 [api.md](api.md)，表结构见 [db.md](db.md)。

---

## 1. 两层服务架构（HTTP 网关 + gRPC 业务）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| HTTP 入口 | 接收浏览器请求、跨域、路由 | Go、[gin-gonic/gin](https://github.com/gin-gonic/gin)、[gin-contrib/cors](https://github.com/gin-contrib/cors) |
| 服务间调用 | Web 仅转发到业务层，不直连本地业务包 | [gRPC](https://grpc.io/)、[protobuf](https://protobuf.dev/)、仓库内 [logic-grpc-service/api/proto](logic-grpc-service/api/proto/recruitment/v1/recruitment.proto) 生成代码 |
| 业务服务 | 用户、岗位、投递、OSS、AI 等全在 Logic | Go、gRPC Server、`recruitment/api` 生成的 `RecruitmentService` 实现 |

---

## 2. 认证与双角色（HR / 候选人）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 注册 / 登录 | 邮箱密码；密码 bcrypt 存储 | Go、[golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| JWT | Web 签发与校验；claims 含用户 ID 与角色 | [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) |
| 调用链传递身份 | Web 解析 JWT 后写入 gRPC metadata：`x-user-id`、`x-user-role` | `google.golang.org/grpc/metadata` |
| Logic 侧鉴权 | 按 RPC 要求 HR / 候选人 / 匿名；资源归属（如岗位仅创建者 HR） | gRPC 拦截约定 + 业务代码校验 |

---

## 3. 岗位与投递业务

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 公开岗位列表 | 游客可见 `active` 岗位 | MySQL、[GORM](https://gorm.io/)、Logic `ListPublicJobs` |
| HR 岗位管理 | 创建、更新、下架；仅能操作本人岗位 | GORM、Logic 按 `hr_user_id` 校验 |
| 候选人投递 | 校验档案完整与已上传简历；防重复投递 | GORM、`applications` 唯一约束与查询 |
| HR 投递台账 | 按岗位分页查看投递与候选人档案摘要 | GORM、JOIN/多次查询组装 `ApplicationRow` |

---

## 4. 候选人结构化档案

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 档案读写 | 姓名、电话、学历、院校、经历、技能等 | MySQL 表 `candidate_profiles`、GORM |
| 完整性校验 | 投递前 Logic 校验必填项 | Go 业务逻辑；前端仅展示与提交表单（React） |

---

## 5. 私有 OSS 与简历（签名 URL）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 上传预签名 | 候选人获取 PUT URL，浏览器直传 Bucket | [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)（`service/s3` PresignPut）、环境变量配置 Endpoint（兼容 MinIO / 云 OSS） |
| 下载预签名 | HR 获取候选人简历 GET URL（需存在投递关系） | S3 PresignGet |
| 格式与安全 | 后缀 + Content-Type 白名单；Confirm 时读取对象头字节校验 PDF/DOC/DOCX 魔数 | Go 标准库、`GetObject` Range 读取 |
| 不落本地盘 | 不落临时文件，仅元数据入库 | 设计上仅 `resumes` 表存 `object_key` 等 |

---

## 6. HR AI 对话（Eino + MySQL 事实）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 业务事实汇总 | 根据问题拼接投递数、在招数、热门岗位等（MySQL 聚合） | GORM / SQL、Go `strings` 规则分支 |
| 大模型回答 | 将事实与用户问题交给 Chat 模型生成自然语言 | [CloudWeGo Eino](https://github.com/cloudwego/eino)、[eino-ext ChatModel 组件](https://github.com/cloudwego/eino-ext)、通义千问 DashScope（`config` 中 `qwen` 或 `QWEN_*`） |
| 未配置模型时 | 仍可返回数据库汇总文本（便于本地无 Key 调试） | 同上分支逻辑 |
| 对话持久化 | 用户消息与助手回复成对写入 `chat_messages`，绑定 `hr_user_id` | GORM |
| 历史加载 | HR 端分页拉取历史消息 | gRPC `ListChatHistory` → HTTP `GET /api/hr/chat/history` |

---

## 7. 前端（两套独立工程）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| HR 端 | 登录/注册 HR、我的岗位、下架、投递台账、简历下载入口、常驻 AI 抽屉 | [React 19](https://react.dev/)、[TypeScript](https://www.typescriptlang.org/)、[Vite](https://vite.dev/)、[React Router](https://reactrouter.com/)、Fetch、`localhost:5173` |
| 候选人端 | 游客浏览岗位、候选人登录/注册、资料编辑、OSS 直传简历、投递 | 同上栈，`localhost:5174` |
| 开发联调 | 同源代理 API | Vite `server.proxy` → `web-gin-service :8080` |

---

## 8. 数据层约定（无 AutoMigrate）

| 能力 | 说明 | 技术栈 |
|------|------|--------|
| 表结构创建 | 由运维/开发者手工执行 SQL | [scripts/schema.sql](scripts/schema.sql) |
| 运行时 | 仅连接数据库，不自动改表 | GORM `Open`，**不使用** `AutoMigrate` |

---

## 9. 契约与文档

| 产出 | 说明 | 技术栈 |
|------|------|--------|
| RPC / HTTP 契约 | 前后端对齐 | Protocol Buffers、`protoc` 生成 Go 代码 |
| 对外说明 | 接口、库表、启动方式 | Markdown：`api.md`、`db.md`、`README.md` |

---

## 10. 联调踩坑记录（OSS 浏览器直传）

以下为真实环境中遇到过的问题与结论，便于答辩或后人排查。

| 现象 | 原因 | 处理 |
|------|------|------|
| 浏览器报 CORS / `Failed to fetch`，控制台写预检无 `Access-Control-Allow-Origin` | ① **武汉等地域 OSS 禁止 path-style**（`https://oss-cn-xxx.aliyuncs.com/bucket/key`），返回 **403 `SecondLevelDomainForbidden`**；错误响应常**不带 CORS 头**，被误当成纯 CORS 问题。② `aws-sdk-go-v2` 在自定义 Endpoint 下若 **`HostnameImmutable: true`**，会一直锁在 endpoint 主机上，**即使用 `UsePathStyle=false` 也仍生成 path-style URL**。 | 对 `*.aliyuncs.com`：**`UsePathStyle=false`** 且 **`HostnameImmutable=false`**（与 MinIO 的 path-style 区分），使预签名为 **virtual-hosted**：`https://{bucket}.oss-cn-xxx.aliyuncs.com/{key}?...`。见 [`logic-grpc-service/internal/osssvc/oss.go`](logic-grpc-service/internal/osssvc/oss.go)。 |
| 预签名里 `X-Amz-SignedHeaders` 只有 `host`，仍加 `Content-Type` | 预检会带 `content-type`，且可能与签名不一致。 | 候选人端 [`user-frontend/src/ProfilePage.tsx`](user-frontend/src/ProfilePage.tsx)：仅用服务端返回的 headers，**不额外设 `Content-Type`**；`body` 用 **`arrayBuffer`**，避免 `File` 自动带类型头。 |
| 阿里云控制台跨域「允许 Methods」没有 OPTIONS | 属正常：控制台只列 GET/PUT/POST/DELETE/HEAD；**预检由 OSS 按规则自动应答**。 | 须勾选业务需要的 **PUT、GET、HEAD**（可按文档再加 **POST**）；**来源**与地址栏一致（`localhost` 与 `127.0.0.1` 需**分别**配置）。 |
| virtual-hosted URL 已通 CORS，但 PUT **403**，XML `AccessDenied` / bucket acl | **RAM 策略或 Bucket 权限**未允许当前 AK 对该 Bucket 执行 **PutObject**（或策略资源 ARN 写错）。 | 在阿里云 RAM 为使用的 AK 对应用户授权 `oss:PutObject`（及所需资源范围）；检查 Bucket 策略是否 Deny。 |
| 用 `curl -v -X OPTIONS` 自查 | 可直接看到是 **403 SecondLevelDomainForbidden** 还是已返回 **`Access-Control-Allow-Origin`**。 | 预签名 URL 过期需重新上传拿新 URL；勿在公开渠道泄露带签名的完整 URL。 |