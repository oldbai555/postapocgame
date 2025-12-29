-- 聊天消息表
CREATE TABLE IF NOT EXISTS `chat_message` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送用户 ID',
  `to_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接收用户 ID（0表示群聊）',
  `room_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '聊天室 ID（群聊或私聊）',
  `content` TEXT NOT NULL COMMENT '消息内容',
  `message_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型：1文本，2图片，3文件',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1已发送，2已读，3已撤回',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_chat_message_room_id` (`room_id`),
  KEY `idx_chat_message_from_user_id` (`from_user_id`),
  KEY `idx_chat_message_to_user_id` (`to_user_id`),
  KEY `idx_chat_message_created_at` (`created_at`),
  KEY `idx_chat_message_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';

-- 在线用户表（用于记录 WebSocket 连接状态）
CREATE TABLE IF NOT EXISTS `chat_online_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `connection_id` VARCHAR(64) NOT NULL COMMENT 'WebSocket 连接 ID',
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP 地址',
  `user_agent` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户代理',
  `last_active_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最后活跃时间(秒级时间戳)',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_chat_online_user_connection_id` (`connection_id`),
  KEY `idx_chat_online_user_user_id` (`user_id`),
  KEY `idx_chat_online_user_last_active_at` (`last_active_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='在线用户表';

