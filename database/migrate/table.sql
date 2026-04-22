-- 创建短url表 shortURLs
CREATE TABLE IF NOT EXISTS urls (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '雪花全局唯一ID',
    longURL TEXT NOT NULL COMMENT '初识长url',
    shortURL VARCHAR(20) NOT NULL UNIQUE COMMENT '生成的短url',
    is_custom BOOL DEFAULT FALSE COMMENT '是否自定义短url',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '时间戳 创建时间',
    expired_at TIMESTAMP NOT NULL COMMENT '时间戳 过期时间'
);

-- 创建索引
CREATE INDEX id_short_code ON urls(shortURL);  -- 为短码创建唯一索引进行兜底
CREATE INDEX id_expired_at ON urls(expired_at);  -- 为过期时间创建索引
