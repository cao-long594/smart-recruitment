# 智能招聘系统（大作业）

双服务：`web-gin-service`（Gin HTTP 网关 + JWT）↔ `logic-grpc-service`（gRPC 业务 + MySQL + 私有 OSS 预签名 + Eino AI）。双前端：`hr-frontend`（5173）、`user-frontend`（5174）。

## 环境要求

- Go 1.22+
- Node.js 20+（前端）
- MySQL 8+
- 对象存储：兼容 S3 API 的私有 Bucket（作业**不考察 Docker**；可直接使用阿里云等云 OSS。下文中的 `docker run minio` **仅为本地可选示例**，不是必须。）
- （可选）阿里云 **通义千问**（DashScope）：用于 Eino ChatModel；未配置时 AI 接口仍返回数据库汇总文本。建议在 **`logic-grpc-service/config/config.yaml`** 的 **`qwen`** 段填写 `api_key`、`base_url`（如 `https://dashscope.aliyuncs.com/compatible-mode/v1`）、`model`（如 `qwen-plus`），或使用环境变量 **`QWEN_API_KEY` / `QWEN_BASE_URL` / `QWEN_MODEL`**。若 `QWEN_BASE_URL` 误写成 `.../chat/completions`，程序会自动截断为 `/v1` 根路径。

## 启动顺序

1. **MySQL**：安装并启动 MySQL 8+ 后，**执行一次** [scripts/schema.sql](scripts/schema.sql)（脚本内已包含 `CREATE DATABASE recruitment` 与 `USE recruitment`，会建好库和全部表）。Logic **不使用** GORM AutoMigrate。  
   - **命令行**（在 `smart-recruitment` 项目根目录下；PowerShell 不支持 `mysql < file`，请用管道或 `cmd /c`）：

     ```powershell
     Get-Content .\scripts\schema.sql -Raw | mysql -u root -p
     ```

     或：`cmd /c "mysql -u root -p < scripts\schema.sql"`

   - **图形界面**（Navicat、MySQL Workbench、DBeaver 等）：新建连接 → 打开 `scripts/schema.sql` → 全选执行。  
   - 若库名或账号不同，在 **`logic-grpc-service/config/config.yaml`** 的 `logic.mysql_dsn` 中修改，或设置环境变量 `MYSQL_DSN`。
2. **对象存储**：创建私有 Bucket，记录 endpoint / region、AK/SK（见下文「配置说明（密钥）」）。若用阿里云 OSS，**无需**安装 Docker。仅在本地想用 MinIO 自测时，可选用：

   ```bash
   docker run -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
   ```

   然后在 MinIO 控制台创建 Bucket（如 `recruitment-resumes`），关闭匿名访问。

3. **统一配置**：复制 [logic-grpc-service/config/config.example.yaml](logic-grpc-service/config/config.example.yaml) 为 **`logic-grpc-service/config/config.yaml`**（已在 [.gitignore](.gitignore) 中忽略），填写 `logic`（MySQL、gRPC 端口）、`web`（HTTP 端口、JWT）、`oss`、**`qwen`** 等。在 **`logic-grpc-service`** 目录启动时读取 `./config/config.yaml`；在 **`web-gin-service`** 目录启动时读取 `../logic-grpc-service/config/config.yaml`。也可用环境变量 **`CONFIG_FILE`** 指定 YAML 绝对路径。

4. **logic-grpc-service**（先启动）：

   ```powershell
   cd logic-grpc-service
   go run .
   ```

5. **web-gin-service**：

   ```powershell
   cd web-gin-service
   go run .
   ```

6. **前端**（两个终端）：

   ```powershell
   cd hr-frontend; npm run dev
   cd user-frontend; npm run dev
   ```

   开发模式下 Vite 已将 `/api` 代理到 `http://127.0.0.1:8080`。

## 配置说明（密钥）

- **禁止**将真实 OSS 密钥、`qwen.api_key`、`web.jwt_secret` 等提交到 Git。使用 **`logic-grpc-service/config/config.yaml`**（已忽略）或流水线密钥环境变量。
- **统一配置文件** [logic-grpc-service/config/config.example.yaml](logic-grpc-service/config/config.example.yaml) → **`logic-grpc-service/config/config.yaml`**，包含：
  - **`logic`**：`grpc_addr`、`mysql_dsn`
  - **`web`**：`http_addr`、`logic_grpc_addr`、`jwt_secret`
  - **`oss`**：`bucket`、`region`、`endpoint`、`access_key_id`、`secret_access_key`
  - **`qwen`**（可选，通义千问）：`api_key`、`model`、`base_url`
- **作业要求（私有 OSS）**：在阿里云创建 **私有 Bucket**，**阻止公共访问**；将 Bucket 信息与 AK/SK 填入 **`logic-grpc-service/config/config.yaml`** 的 `oss` 段，勿写入业务代码。
- **优先级**：**环境变量覆盖 YAML**。常用环境变量：`CONFIG_FILE`、`MYSQL_DSN`、`LOGIC_GRPC_ADDR`、`OSS_*`、`QWEN_*`、`WEB_HTTP_ADDR`、`JWT_SECRET`。
- 阿里云 OSS 使用 **S3 兼容接口** 时须设置 **`oss.endpoint`**（HTTPS）及一致的 **`oss.region`**（地域 ID，武汉本地域一般为 `cn-wuhan-lr`，若预签名失败请以控制台「Bucket 概览」为准）。**武汉等多地域禁止 path-style**（`oss-xxx.aliyuncs.com/bucket/key`），会返回 **403 `SecondLevelDomainForbidden: Please use virtual hosted style`**，浏览器里常被误报成 CORS；须使用 **virtual-hosted**（`bucket.oss-xxx.aliyuncs.com/key`）。本仓库对 `*.aliyuncs.com` 的 Endpoint 已自动走 virtual-hosted；**本地 MinIO** 仍为 path-style。
- **浏览器直传 OSS 必须配置 CORS**：候选人端会用 `fetch(预签名URL, PUT)` 并携带 `Content-Type` 等头，浏览器会先发 **OPTIONS 预检**。阿里云 OSS 控制台「允许 Methods」**只有** `GET` / `PUT` / `POST` / `DELETE` / `HEAD`，**没有单独勾 OPTIONS**；预检由 OSS 按规则自动应答，**务必勾选实际会用到的 `PUT`（及 `GET`、`HEAD`）**，官方「前端直传」示例通常还包含 **`POST`**。规则保存后约 **15 分钟内**生效。在阿里云 OSS 控制台 → 目标 Bucket → **数据安全** / **跨域设置（CORS）** 示例：
  - **来源**：`http://localhost:5174`；若地址栏是 **`http://127.0.0.1:5174`**，须**另加一行**来源（二者不算同一来源）。HR 端可再加 `http://localhost:5173` 等。
  - **方法**：至少 **`PUT`、`GET`、`HEAD`**；仍报错可再勾 **`POST`**（与文档「前端直传」一致）。
  - **允许 Headers**：`*`（直传预签名常带多种头）或至少包含 `Content-Type` 及预签名中的头（如 `x-amz-*`）。
  - **暴露 Headers**：可填 `ETag`、`x-oss-request-id`（便于排查）。
  - **返回 Vary: Origin**：来源多或调试跨域时可勾选（见控制台说明）。
  - **预签名与浏览器**：若预签名 URL 查询参数里 **`X-Amz-SignedHeaders` 只有 `host`**，浏览器 **不要再额外加 `Content-Type` 等未签名请求头**（否则预检更复杂且易 CORS/签名失败）。本仓库候选人端上传已按此处理；若仍 **403**，再考虑在服务端把 `Content-Type` 纳入签名。

## 接口与数据

- [api.md](api.md)：HTTP 路径与鉴权。
- [db.md](db.md)：表结构与关系。
- [AGENDS.md](AGENDS.md)：已实现功能与各功能所用技术栈对照。
- Protobuf：[logic-grpc-service/api/proto/recruitment/v1/recruitment.proto](logic-grpc-service/api/proto/recruitment/v1/recruitment.proto)

### 重新生成 gRPC 代码

```powershell
cd logic-grpc-service/api
mkdir -Force gen\go | Out-Null
protoc -I proto --go_out=gen/go --go_opt=module=recruitment/api/gen/go `
  --go-grpc_out=gen/go --go-grpc_opt=module=recruitment/api/gen/go `
  proto/recruitment/v1/recruitment.proto
```

## 项目亮点（便于答辩/文档）

- Web 层 **无业务 SQL**，仅 JWT、参数与 gRPC 转发；业务与权限在 Logic 统一实现。
- 简历 **直传 OSS**，服务端 **`Confirm` 阶段拉取对象头字节** 校验 PDF/DOC/DOCX 魔数；下载走 **短期预签名 GET**。
- HR AI：**先查 MySQL 汇总事实**，再经 **CloudWeGo Eino** 调用通义千问生成自然语言；对话 **成对入库**，刷新可加载历史。
- 双前端独立工程，候选人端游客可浏览岗位，投递前后端强校验档案与简历。

## 作业提交相关

演示视频、表单提交、拓展思考题书面作答由课程要求自行完成，本仓库不包含自动提交脚本。
