# HTTP API 说明（web-gin-service）

Base URL：`http://localhost:8080`（可用环境变量 `WEB_HTTP_ADDR` 修改）。

鉴权：除注册、登录、公开岗位列表外，均在 Header 携带 `Authorization: Bearer <JWT>`。

JWT Payload（自定义 claims）：`uid`（用户 ID）、`role`（`hr` | `candidate`）。

gRPC 元数据：Web 将 JWT 解析后，对每个受保护请求向 Logic 注入 `x-user-id`、`x-user-role`。

---

## 认证

### POST `/api/register`

注册。

请求 JSON：

```json
{
  "email": "a@b.com",
  "password": "secret",
  "role": "hr"
}
```

`role`：`hr` 或 `candidate`。

响应：`{ "user_id": 1 }`

### POST `/api/login`

登录。

请求 JSON：`{ "email": "", "password": "" }`

响应：`{ "token": "...", "user_id": 1, "role": "hr" }`

---

## 公开

### GET `/api/jobs`

游客岗位列表（仅 `status=active`）。

Query：`page`（默认 1）、`page_size`（默认 20）

响应：与 gRPC `ListPublicJobsResponse` 相同（`jobs` + `total`）。

---

## HR（需 JWT，`role=hr`）

### GET `/api/hr/jobs`

我发布的岗位（含 archived）。

Query：`page`, `page_size`

### POST `/api/hr/jobs`

创建岗位。JSON：`{ "title": "", "description": "" }`

### PATCH `/api/hr/jobs/:id`

更新岗位。JSON：`{ "title": "", "description": "" }`（可部分字段）

### POST `/api/hr/jobs/:id/archive`

下架岗位。

### GET `/api/hr/jobs/:id/applications`

某岗位投递分页台账。Query：`page`, `page_size`

### GET `/api/hr/resume_download`

获取候选人简历预签名下载 URL。Query：`candidate_user_id=123`

### POST `/api/hr/chat`

AI 问答（Eino + MySQL 事实）。JSON：`{ "content": "投递总数是多少？" }`

响应：`{ "answer": "..." }`

### GET `/api/hr/chat/history`

聊天历史。Query：`page`, `page_size`（按 `id` 升序分页）

---

## 候选人（需 JWT，`role=candidate`）

### GET `/me/profile`

获取档案与简历元信息摘要。

### PUT `/me/profile`

更新档案。JSON 字段：`name`, `phone`, `education`, `school`, `experience`, `skills`

### POST `/me/resume/upload_url`

申请直传 OSS 的预签名 PUT。JSON：

```json
{
  "file_name": "cv.pdf",
  "content_type": "application/pdf"
}
```

响应：`upload_url`, `object_key`, `headers`（需原样带上 PUT）。

前端：`PUT upload_url`，Body=文件二进制，Content-Type 与申请时一致。

### POST `/me/resume/confirm`

确认已上传。JSON：

```json
{
  "object_key": "...",
  "file_name": "cv.pdf",
  "content_type": "application/pdf",
  "size_bytes": 12345
}
```

Logic 会从 OSS 读取文件头校验 PDF/DOC/DOCX。

### POST `/me/resume/parse`

解析当前候选人已确认上传的简历，返回可填入档案表单的结构化字段。

说明：

- 当前自动解析支持 DOCX 文本简历；PDF/DOC 仍可上传，但暂不自动解析。
- 后端从私有 OSS 读取当前登录用户自己的简历，不要求前端传对象地址。
- 响应只用于前端填表预览，不会自动保存档案；用户确认后仍需调用 `PUT /me/profile`。

响应：

```json
{
  "profile": {
    "name": "张三",
    "phone": "13800138000",
    "education": "本科",
    "school": "武汉科技大学",
    "experience": "项目经历...",
    "skills": "Go,React,MySQL"
  }
}
```

### POST `/api/jobs/:id/apply`

投递岗位（后端校验档案与简历）。

---

## 错误格式

HTTP 4xx/5xx，body：`{ "error": "message" }`

常见：401 未登录、403 gRPC `PermissionDenied`、400 参数/业务前置条件失败、409 重复投递。

---

## Proto 契约

完整 RPC 定义见 [logic-grpc-service/api/proto/recruitment/v1/recruitment.proto](logic-grpc-service/api/proto/recruitment/v1/recruitment.proto)。生成代码目录：`logic-grpc-service/api/gen/go/...`。
