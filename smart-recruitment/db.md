# 数据库设计说明（智能招聘系统）

## 概述

- 引擎：MySQL 8+，字符集 `utf8mb4`。
- ORM：logic-grpc-service 使用 GORM 访问数据；**表结构由 [scripts/schema.sql](scripts/schema.sql) 手工创建**，不使用 AutoMigrate（与 `schema.sql` 字段须保持一致）。

## ER 关系（文字）

- `users` 1 — 0..1 `candidate_profiles`（仅 `role=candidate` 使用档案）
- `users` 1 — 0..1 `resumes`（候选人简历元数据，文件在 OSS）
- `users`(HR) 1 — N `jobs`
- `jobs` N — M `users`(候选人) 通过 `applications`（同一岗位同一候选人唯一）

## 表结构

### users

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 用户 ID |
| email | VARCHAR(191) UNIQUE | 登录邮箱 |
| password_hash | VARCHAR(255) | bcrypt 哈希 |
| role | VARCHAR(16) | `hr` / `candidate` |
| created_at | DATETIME(3) | 创建时间 |

### jobs

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 岗位 ID |
| hr_user_id | BIGINT IDX | 发布者 HR 的 user.id |
| title | VARCHAR(255) | 标题 |
| description | TEXT | 描述 |
| status | VARCHAR(32) | `active` 在招 / `archived` 下架 |
| created_at / updated_at | DATETIME(3) | 时间戳 |

### candidate_profiles

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | BIGINT PK | 候选人 user.id |
| name, phone, education, school | 字符串 | 结构化档案必填项（作业要求） |
| experience, skills | TEXT | 经历与技能标签 |
| updated_at | DATETIME(3) | 更新时间 |

### resumes

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | BIGINT PK | 候选人 |
| object_key | VARCHAR(512) | 私有 OSS 对象键 |
| file_name | VARCHAR(255) | 原始文件名 |
| content_type | VARCHAR(128) | MIME |
| size_bytes | BIGINT | 大小 |
| updated_at | DATETIME(3) | 更新时间 |

### applications

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 投递记录 |
| job_id | BIGINT | 岗位 |
| candidate_user_id | BIGINT | 候选人 |
| created_at | DATETIME(3) | 投递时间 |
| UNIQUE(job_id, candidate_user_id) | | 防重复投递 |

### chat_messages

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AI | 消息 ID |
| hr_user_id | BIGINT IDX | HR 账号 |
| role | VARCHAR(32) | `user`（提问）/ `assistant`（回答） |
| content | LONGTEXT | 文本 |
| created_at | DATETIME(3) IDX | 时间 |

## 约束与业务规则（逻辑层）

- 岗位编辑/下架：仅 `jobs.hr_user_id` 与当前 HR 一致。
- 投递：候选人档案必填项齐全且存在 `resumes` 记录；岗位须 `active`。
- 简历下载：仅当该候选人向当前 HR 的任一岗位投递过时，HR 可获取预签名下载 URL。
- AI 对话：仅 HR；消息按 `hr_user_id` 隔离。
