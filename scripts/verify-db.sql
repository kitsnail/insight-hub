-- 验证表结构
\dt

-- 验证扩展
\dx

-- 验证视图
\dv

-- 验证标签数据
SELECT * FROM tags;

-- 验证活跃追踪视图
SELECT * FROM active_trackings LIMIT 5;

-- 验证最近简报视图
SELECT * FROM recent_briefs LIMIT 5;
