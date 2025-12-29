-- admin-server 数据库建表脚本
-- 注意：所有业务表统一包含 created_at、updated_at、deleted_at 字段（BIGINT类型，秒级时间戳）
-- 关联关系表（如用户-角色、角色-权限）不包含 deleted_at 字段，使用物理删除

-- ============================================
-- 1. 后台管理用户表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `password_hash` VARCHAR(255) NOT NULL COMMENT 'bcrypt 加密后的密码',
  `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `signature` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个性签名',
  `department_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '部门ID',
  `status` INT NOT NULL DEFAULT 1 COMMENT '账号状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_user_username` (`username`),
  KEY `idx_admin_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理用户表';

-- ============================================
-- 2. 后台角色表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_role` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名称',
  `code` VARCHAR(64) NOT NULL COMMENT '角色编码（唯一）',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '角色描述',
  `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_role_code` (`code`),
  KEY `idx_admin_role_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台角色表';

-- ============================================
-- 3. 后台权限表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `name` VARCHAR(64) NOT NULL COMMENT '权限名称',
  `code` VARCHAR(128) NOT NULL COMMENT '权限编码（唯一，例如 system:user:list）',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '权限描述',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permission_code` (`code`),
  KEY `idx_admin_permission_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台权限表';

-- ============================================
-- 4. 后台部门表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_department` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '部门ID',
  `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父部门ID',
  `name` VARCHAR(64) NOT NULL COMMENT '部门名称',
  `order_num` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_admin_department_parent` (`parent_id`),
  KEY `idx_admin_department_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台部门表';

-- ============================================
-- 5. 用户-角色关联表（关联表不使用软删除）
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_user_role` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_user_role` (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色关联表';

-- ============================================
-- 6. 角色-权限关联表（关联表不使用软删除）
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_role_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_role_permission` (`role_id`,`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-权限关联表';

-- ============================================
-- 7. 后台菜单/按钮表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_menu` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  `parent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID',
  `name` VARCHAR(64) NOT NULL COMMENT '菜单名称',
  `path` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '前端路由路径',
  `component` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '前端组件路径',
  `icon` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '图标',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型：1 目录 2 菜单 3 按钮',
  `order_num` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `visible` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可见：1 是，0 否',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_admin_menu_parent` (`parent_id`),
  KEY `idx_admin_menu_type` (`type`),
  KEY `idx_admin_menu_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台菜单/按钮表';

-- ============================================
-- 8. 权限-菜单关联表（关联表不使用软删除）
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_permission_menu` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `menu_id` BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permission_menu` (`permission_id`,`menu_id`),
  KEY `idx_admin_permission_menu_permission` (`permission_id`),
  KEY `idx_admin_permission_menu_menu` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限-菜单关联表';

-- ============================================
-- 9. 接口表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_api` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '接口ID',
  `name` VARCHAR(64) NOT NULL COMMENT '接口名称',
  `method` VARCHAR(10) NOT NULL COMMENT 'HTTP方法（GET、POST、PUT、DELETE等）',
  `path` VARCHAR(255) NOT NULL COMMENT '接口路径（如 /api/v1/users）',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '接口描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_api_method_path` (`method`,`path`),
  KEY `idx_admin_api_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='接口表';

-- ============================================
-- 10. 权限-接口关联表（关联表不使用软删除）
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_permission_api` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `api_id` BIGINT UNSIGNED NOT NULL COMMENT '接口ID',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permission_api` (`permission_id`,`api_id`),
  KEY `idx_admin_permission_api_permission` (`permission_id`),
  KEY `idx_admin_permission_api_api` (`api_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限-接口关联表';

-- ============================================
-- 11. 系统配置表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `group` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '配置分组（如 system、app、theme 等）',
  `key` VARCHAR(128) NOT NULL COMMENT '配置键（唯一，格式：group:key）',
  `value` TEXT COMMENT '配置值（JSON 格式存储复杂数据）',
  `type` VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT '配置类型（string、number、boolean、json）',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '配置描述',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_config_key` (`key`),
  KEY `idx_admin_config_group` (`group`),
  KEY `idx_admin_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ============================================
-- 12. 数据字典类型表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_dict_type` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '字典类型ID',
  `name` VARCHAR(64) NOT NULL COMMENT '字典类型名称',
  `code` VARCHAR(64) NOT NULL COMMENT '字典类型编码（唯一）',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '字典类型描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_dict_type_code` (`code`),
  KEY `idx_admin_dict_type_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据字典类型表';

-- ============================================
-- 13. 数据字典项表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_dict_item` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '字典项ID',
  `type_id` BIGINT UNSIGNED NOT NULL COMMENT '字典类型ID',
  `label` VARCHAR(64) NOT NULL COMMENT '字典项标签（显示名称）',
  `value` VARCHAR(128) NOT NULL COMMENT '字典项值',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1 启用，0 禁用',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_admin_dict_item_type` (`type_id`),
  KEY `idx_admin_dict_item_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据字典项表';

-- ============================================
-- 14. 文件表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_file` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `name` VARCHAR(255) NOT NULL COMMENT '文件名称',
  `original_name` VARCHAR(255) NOT NULL COMMENT '原始文件名称',
  `path` VARCHAR(512) NOT NULL COMMENT '文件存储路径（相对路径）',
  `url` VARCHAR(512) NOT NULL COMMENT '文件访问URL',
  `size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
  `mime_type` VARCHAR(128) DEFAULT NULL COMMENT 'MIME类型',
  `ext` VARCHAR(16) DEFAULT NULL COMMENT '文件扩展名',
  `storage_type` VARCHAR(32) NOT NULL DEFAULT 'local' COMMENT '存储类型（local、oss、s3等）',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1 正常，0 禁用',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_admin_file_storage_type` (`storage_type`),
  KEY `idx_admin_file_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

-- ============================================
-- 15. 操作日志表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_operation_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户 ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `operation_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作类型：create/update/delete/query/export等',
  `operation_object` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '操作对象（模块/表名，如user/role/permission）',
  `method` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '请求方法：GET/POST/PUT/DELETE',
  `path` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
  `request_params` TEXT COMMENT '请求参数（JSON格式）',
  `response_code` INT NOT NULL DEFAULT 0 COMMENT '响应状态码',
  `response_msg` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '响应消息',
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP 地址',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '用户代理',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '请求耗时（毫秒）',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_operation_log_user_id` (`user_id`),
  KEY `idx_operation_log_operation_type` (`operation_type`),
  KEY `idx_operation_log_operation_object` (`operation_object`),
  KEY `idx_operation_log_created_at` (`created_at`),
  KEY `idx_operation_log_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- ============================================
-- 16. 登录日志表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_login_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户 ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '登录 IP 地址',
  `location` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '登录地点（通过IP解析）',
  `browser` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '浏览器',
  `os` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作系统',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '用户代理',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '登录状态：0失败 1成功',
  `message` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录消息（失败原因或成功提示）',
  `login_at` BIGINT NOT NULL DEFAULT 0 COMMENT '登录时间(秒级时间戳)',
  `logout_at` BIGINT NOT NULL DEFAULT 0 COMMENT '登出时间(秒级时间戳,0表示未登出)',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_login_log_user_id` (`user_id`),
  KEY `idx_login_log_username` (`username`),
  KEY `idx_login_log_status` (`status`),
  KEY `idx_login_log_login_at` (`login_at`),
  KEY `idx_login_log_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';

-- ============================================
-- 17. 审计日志表
-- ============================================
CREATE TABLE IF NOT EXISTS `audit_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户 ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `audit_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '审计类型：permission_assign/role_change/config_modify/data_delete等',
  `audit_object` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '审计对象（模块/表名，如user_role/role_permission/role/config）',
  `audit_detail` TEXT COMMENT '审计详情（JSON格式，记录变更前后的数据）',
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP 地址',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '用户代理',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_audit_log_user_id` (`user_id`),
  KEY `idx_audit_log_audit_type` (`audit_type`),
  KEY `idx_audit_log_audit_object` (`audit_object`),
  KEY `idx_audit_log_created_at` (`created_at`),
  KEY `idx_audit_log_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- ============================================
-- 18. 接口性能监控日志表
-- ============================================
CREATE TABLE IF NOT EXISTS `admin_performance_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户 ID',
  `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `method` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '请求方法：GET/POST/PUT/DELETE等',
  `path` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
  `status_code` INT NOT NULL DEFAULT 0 COMMENT '响应状态码',
  `duration` INT NOT NULL DEFAULT 0 COMMENT '请求耗时（毫秒）',
  `is_slow` TINYINT NOT NULL DEFAULT 0 COMMENT '是否慢接口：0 否，1 是',
  `slow_threshold` INT NOT NULL DEFAULT 0 COMMENT '慢接口阈值（毫秒）',
  `ip_address` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP 地址',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '用户代理',
  `error_msg` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '错误信息（状态码>=400时记录）',
  `created_at` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_performance_log_path` (`path`),
  KEY `idx_performance_log_created_at` (`created_at`),
  KEY `idx_performance_log_is_slow` (`is_slow`),
  KEY `idx_performance_log_duration` (`duration`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='接口性能监控日志表';

