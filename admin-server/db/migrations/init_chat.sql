-- 在线聊天模块初始化 SQL
-- 功能组: chat
-- 功能名称: 在线聊天

-- ============================================
-- 1. 获取临时目录 ID
-- ============================================
-- 临时目录用于存放新功能模块的菜单，方便后续整理
-- 注意：临时目录已在 data.sql 中初始化（id=9）
SET @temp_dir_id = (SELECT `id` FROM `admin_menu` WHERE `id` = 9 AND `deleted_at` = 0 LIMIT 1);

-- ============================================
-- 2. 插入菜单数据
-- ============================================
-- 在线聊天主菜单
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @temp_dir_id,
    '在线聊天',
    '/temp/chat',
    'temp/ChatList',
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

-- 在线聊天新增按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '在线聊天 新增按钮',
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

-- 在线聊天编辑按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '在线聊天 编辑按钮',
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

-- 在线聊天删除按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @main_menu_id,
    '在线聊天 删除按钮',
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
-- 在线聊天列表权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天列表',
    'chat:list',
    '查看在线聊天列表',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_permission_id = LAST_INSERT_ID();

-- 在线聊天新增权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天新增',
    'chat:create',
    '新增在线聊天',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_permission_id = LAST_INSERT_ID();

-- 在线聊天编辑权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天编辑',
    'chat:update',
    '编辑在线聊天',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_permission_id = LAST_INSERT_ID();

-- 在线聊天删除权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天删除',
    'chat:delete',
    '删除在线聊天',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_permission_id = LAST_INSERT_ID();

-- ============================================
-- 4. 插入接口数据
-- ============================================
-- 在线聊天列表接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天列表',
    'GET',
    '/api/v1/chats',
    '获取在线聊天列表',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @list_api_id = LAST_INSERT_ID();

-- 在线聊天新增接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天新增',
    'POST',
    '/api/v1/chats',
    '新增在线聊天',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @create_api_id = LAST_INSERT_ID();

-- 在线聊天编辑接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天编辑',
    'PUT',
    '/api/v1/chats/:id',
    '编辑在线聊天',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @update_api_id = LAST_INSERT_ID();

-- 在线聊天删除接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '在线聊天删除',
    'DELETE',
    '/api/v1/chats/:id',
    '删除在线聊天',
    1, -- 状态：1 启用（可根据需要设置为 0 禁用）
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
);

SET @delete_api_id = LAST_INSERT_ID();

-- ============================================
-- 5. 插入权限-菜单关联数据
-- ============================================
-- 在线聊天列表权限 -> 在线聊天主菜单
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @main_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天新增权限 -> 在线聊天新增按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天编辑权限 -> 在线聊天编辑按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天删除权限 -> 在线聊天删除按钮
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 6. 插入权限-接口关联数据
-- ============================================
-- 在线聊天列表权限 -> GET /api/v1/chats接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@list_permission_id, @list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天新增权限 -> POST /api/v1/chats接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@create_permission_id, @create_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天编辑权限 -> PUT /api/v1/chats/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@update_permission_id, @update_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 在线聊天删除权限 -> DELETE /api/v1/chats/:id接口
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@delete_permission_id, @delete_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 添加聊天配置字典（使用自增ID）

-- 1. 添加字典类型：聊天配置
INSERT INTO `admin_dict_type` (`name`, `code`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('聊天配置', 'chat_config', '在线聊天相关配置字典', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`), 
  `description`=VALUES(`description`), 
  `updated_at`=UNIX_TIMESTAMP(), 
  `deleted_at`=0;

-- 2. 添加字典项：聊天窗口消息数量限制
-- 注意：type_id 需要通过子查询获取，因为使用了自增ID
INSERT INTO `admin_dict_item` (`type_id`, `label`, `value`, `sort`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`)
SELECT 
  id, '聊天窗口消息数量', '30', 1, 1, '每个聊天窗口显示的最新消息数量', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0
FROM `admin_dict_type`
WHERE `code` = 'chat_config' AND `deleted_at` = 0
ON DUPLICATE KEY UPDATE 
  `label`=VALUES(`label`), 
  `value`=VALUES(`value`), 
  `sort`=VALUES(`sort`), 
  `status`=VALUES(`status`), 
  `remark`=VALUES(`remark`), 
  `updated_at`=UNIX_TIMESTAMP(), 
  `deleted_at`=0;

