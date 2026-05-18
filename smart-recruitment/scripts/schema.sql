-- 手工执行：启动 logic-grpc-service 前运行本脚本（不使用 AutoMigrate）。
-- 会创建库 recruitment（若不存在）并建表。

CREATE DATABASE IF NOT EXISTS recruitment
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE recruitment;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(191) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(16) NOT NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)
);

CREATE TABLE IF NOT EXISTS jobs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  hr_user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  INDEX idx_jobs_hr (hr_user_id)
);

CREATE TABLE IF NOT EXISTS candidate_profiles (
  user_id BIGINT PRIMARY KEY,
  name VARCHAR(128),
  phone VARCHAR(64),
  education VARCHAR(64),
  school VARCHAR(128),
  experience TEXT,
  skills TEXT,
  updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);

CREATE TABLE IF NOT EXISTS resumes (
  user_id BIGINT PRIMARY KEY,
  object_key VARCHAR(512) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  content_type VARCHAR(128) NOT NULL,
  size_bytes BIGINT NOT NULL,
  updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);

CREATE TABLE IF NOT EXISTS applications (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  job_id BIGINT NOT NULL,
  candidate_user_id BIGINT NOT NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY idx_job_candidate (job_id, candidate_user_id),
  INDEX idx_app_job (job_id),
  INDEX idx_app_cand (candidate_user_id)
);

CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  hr_user_id BIGINT NOT NULL,
  role VARCHAR(32) NOT NULL,
  content LONGTEXT NOT NULL,
  created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_chat_hr (hr_user_id),
  INDEX idx_chat_created (created_at)
);
