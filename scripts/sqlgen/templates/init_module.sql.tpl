-- {{.Name}}模块初始化 SQL
-- 功能组: {{.Group}}
-- 功能名称: {{.Name}}

-- ============================================
-- 1. 获取临时目录 ID
-- ============================================
-- 临时目录用于存放新功能模块的菜单，方便后续整理
-- 注意：临时目录已在 data.sql 中初始化（id=9）
SET @temp_dir_id = (SELECT `id` FROM `admin_menu` WHERE `id` = 9 AND `deleted_at` = 0 LIMIT 1);

-- ============================================
-- 2. 插入菜单数据
-- ============================================
-- {{.Name}}主菜单
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @temp_dir_id,
    '{{.Name}}',
    '{{.Path}}',
    '{{.Component}}',
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

-- {{.Name}}新增按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '{{.Name}} 新增按钮',
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

-- {{.Name}}编辑按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '{{.Name}} 编辑按钮',
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

-- {{.Name}}删除按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '{{.Name}} 删除按钮',
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
-- {{.Name}}列表权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}列表',
    '{{.Group}}:list',
    '查看{{.Name}}列表',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_permission_id = LAST_INSERT_ID();

-- {{.Name}}新增权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}新增',
    '{{.Group}}:create',
    '新增{{.Name}}',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_permission_id = LAST_INSERT_ID();

-- {{.Name}}编辑权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}编辑',
    '{{.Group}}:update',
    '编辑{{.Name}}',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_permission_id = LAST_INSERT_ID();

-- {{.Name}}删除权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}删除',
    '{{.Group}}:delete',
    '删除{{.Name}}',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_permission_id = LAST_INSERT_ID();

-- ============================================
-- 4. 插入接口数据
-- ============================================
-- {{.Name}}列表接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}列表',
    'GET',
    '{{.APIBasePath}}',
    '获取{{.Name}}列表',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_api_id = LAST_INSERT_ID();

-- {{.Name}}新增接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}新增',
    'POST',
    '{{.APIBasePath}}',
    '新增{{.Name}}',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_api_id = LAST_INSERT_ID();

-- {{.Name}}编辑接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}编辑',
    'PUT',
    '{{.APIBasePath}}/:id',
    '编辑{{.Name}}',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_api_id = LAST_INSERT_ID();

-- {{.Name}}删除接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '{{.Name}}删除',
    'DELETE',
    '{{.APIBasePath}}/:id',
    '删除{{.Name}}',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_api_id = LAST_INSERT_ID();

-- ============================================
-- 5. 插入权限-菜单关联数据
-- ============================================
-- {{.Name}}列表权限 -> {{.Name}}主菜单
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @main_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}新增权限 -> {{.Name}}新增按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}编辑权限 -> {{.Name}}编辑按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}删除权限 -> {{.Name}}删除按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 6. 插入权限-接口关联数据
-- ============================================
-- {{.Name}}列表权限 -> GET {{.APIBasePath}}接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}新增权限 -> POST {{.APIBasePath}}接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}编辑权限 -> PUT {{.APIBasePath}}/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- {{.Name}}删除权限 -> DELETE {{.APIBasePath}}/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

