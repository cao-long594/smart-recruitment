-- 向 recruitment 库插入 10 条示例岗位（绑定「第一个 HR 用户」）。
-- 使用前请先在 HR 端注册过 role=hr 的账号。
-- 若遇 ERROR 1366：请保证本文件为 UTF-8 编码，并用 mysql --default-character-set=utf8mb4 执行。

USE recruitment;

-- 本会话使用 utf8mb4，避免中文写入报错
SET NAMES utf8mb4;

-- 若你已知 HR 的 user id（例如 1），可改成: SET @hr_id = 1;
SET @hr_id = (SELECT id FROM users WHERE role = 'hr' ORDER BY id LIMIT 1);

INSERT INTO jobs (hr_user_id, title, description, status) VALUES
(@hr_id, 'Go 后端工程师', '负责 Gin + gRPC 微服务开发与维护。', 'active'),
(@hr_id, '前端开发（React）', '负责 HR 与候选人双端页面与联调。', 'active'),
(@hr_id, '全栈实习', '参与招聘系统功能迭代，熟悉前后端分离。', 'active'),
(@hr_id, 'MySQL DBA 助理', '协助库表设计、慢查询与数据备份。', 'active'),
(@hr_id, '测试工程师', '接口测试、端到端流程与回归。', 'active'),
(@hr_id, '产品经理（实习）', '需求梳理、原型与验收标准。', 'active'),
(@hr_id, 'DevOps 实习', 'CI、部署脚本与监控告警（作业不考察 Docker 亦可了解）。', 'active'),
(@hr_id, '算法工程（业务统计）', '与业务报表、数据统计相关（不做简历匹配向量检索）。', 'active'),
(@hr_id, '客服/运营支持', '协助企业与候选人沟通与台账整理。', 'active'),
(@hr_id, '安全合规助理', '权限审计、密钥与 OSS 私有访问规范落地。', 'active');
