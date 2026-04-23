-- 创建短url表 urls（表名保持不变）
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY COMMENT '雪花全局唯一ID',
    longurl TEXT NOT NULL COMMENT '原始长url',
    shorturl VARCHAR(20) NOT NULL UNIQUE COMMENT '生成的短url',
    selfshorturl VARCHAR(20) COMMENT '自定义短url标识',
    iscustom BOOLEAN DEFAULT FALSE COMMENT '是否自定义短url',
    expiretime TIMESTAMP NOT NULL COMMENT '过期时间',
    createdtime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

-- 创建索引
CREATE INDEX idx_shorturl ON urls(shorturl);
CREATE INDEX idx_expiretime ON urls(expiretime);