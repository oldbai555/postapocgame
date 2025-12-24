-- 文件管理模块初始化 SQL
-- 功能组: file
-- 功能名称: 文件管理

-- ============================================
-- 1. 获取临时目录 ID
-- ============================================
-- 临时目录用于存放新功能模块的菜单，方便后续整理
-- 注意：临时目录已在 data.sql 中初始化（id=9）
SET @temp_dir_id = (SELECT `id` FROM `admin_menu` WHERE `id` = 9 AND `deleted_at` = 0 LIMIT 1);

-- ============================================
-- 2. 插入菜单数据
-- ============================================
-- 文件管理主菜单
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @temp_dir_id,
    '文件管理',
    '/temp/file',
    'temp/FileList',
    'ele-Document',
    2, -- 类型：2 菜单
    0, -- 排序值（可根据需要调整）
    1, -- 是否可见：1 是
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

-- 获取主菜单 ID
SET @main_menu_id = LAST_INSERT_ID();

-- 文件管理新增按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '文件管理 新增按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    1, -- 排序值
    0, -- 是否可见：0 否（按钮不显示在菜单中）
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_button_id = LAST_INSERT_ID();

-- 文件管理编辑按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '文件管理 编辑按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    2, -- 排序值
    0, -- 是否可见：0 否
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_button_id = LAST_INSERT_ID();

-- 文件管理删除按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '文件管理 删除按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    3, -- 排序值
    0, -- 是否可见：0 否
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_button_id = LAST_INSERT_ID();

-- ============================================
-- 3. 插入权限数据
-- ============================================
-- 文件管理列表权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理列表',
    'file:list',
    '查看文件管理列表',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_permission_id = LAST_INSERT_ID();

-- 文件管理新增权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理新增',
    'file:create',
    '新增文件管理',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_permission_id = LAST_INSERT_ID();

-- 文件管理编辑权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理编辑',
    'file:update',
    '编辑文件管理',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_permission_id = LAST_INSERT_ID();

-- 文件管理删除权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理删除',
    'file:delete',
    '删除文件管理',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_permission_id = LAST_INSERT_ID();

-- ============================================
-- 4. 插入接口数据
-- ============================================
-- 文件管理列表接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理列表',
    'GET',
    '/api/v1/files',
    '获取文件管理列表',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_api_id = LAST_INSERT_ID();

-- 文件管理新增接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理新增',
    'POST',
    '/api/v1/files',
    '新增文件管理',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_api_id = LAST_INSERT_ID();

-- 文件管理编辑接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理编辑',
    'PUT',
    '/api/v1/files/:id',
    '编辑文件管理',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_api_id = LAST_INSERT_ID();

-- 文件管理删除接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '文件管理删除',
    'DELETE',
    '/api/v1/files/:id',
    '删除文件管理',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_api_id = LAST_INSERT_ID();

-- ============================================
-- 5. 插入权限-菜单关联数据
-- ============================================
-- 文件管理列表权限 -> 文件管理主菜单
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @main_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理新增权限 -> 文件管理新增按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理编辑权限 -> 文件管理编辑按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理删除权限 -> 文件管理删除按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 6. 插入权限-接口关联数据
-- ============================================
-- 文件管理列表权限 -> GET /api/v1/files接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理新增权限 -> POST /api/v1/files接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理编辑权限 -> PUT /api/v1/files/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 文件管理删除权限 -> DELETE /api/v1/files/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

