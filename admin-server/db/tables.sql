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

