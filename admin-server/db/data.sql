-- admin-server 数据库初始化数据脚本
-- 注意：此脚本中的数据为系统初始化数据，不可被删除（包括软删、硬删）
-- 初始化数据的ID范围（每张表从1开始连续）：
--   admin_user: id=1-2 (1=超级管理员, 2=admin 业务管理员)
--   admin_role: id=1-2 (1=super_admin 超级管理员角色, 2=admin 业务管理员角色)
--   admin_permission: id=1-48 (48个权限，含通用权限 common:xxx)
--   admin_department: id=1 (根部门)
--   admin_menu: id=1-43 (13个菜单 + 30个按钮)
--   admin_api: id=1-58 (58个接口，含缓存刷新接口和个人信息相关接口)
--   admin_user_role: id=1-2 (1=super_admin绑定超级管理员, 2=admin绑定业务管理员)
--   admin_role_permission: id=1-4 (超级管理员角色-权限关联，含 common:profile、common:profile_update、common:password_change)
--   admin_permission_menu: id=1-40 (10个菜单关联 + 30个按钮关联)
--   admin_permission_api: id=1-58 (58个权限-接口关联，id=53-58为通用接口)

-- ============================================
-- 1. 初始化基础数据
-- ============================================
-- 初始化用户：1=超级管理员 oldbai，2=业务管理员 admin（密码：admin）
INSERT INTO `admin_user` (`id`, `username`, `password_hash`, `department_id`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  (1, 'oldbai', '$2a$10$TIjB8/yhHDiyNbJn40BUPOACjxeTccaYTD4Ot3p00ZBCKzh7/sL9q', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, 'admin', '$2a$10$F/GioZ0D2TUl7wQX2kErU.fuqu/IJvU8yd.VtuFXbVfcAYPaZaj7S', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `username`=VALUES(`username`), 
  `password_hash`=VALUES(`password_hash`), 
  `status`=VALUES(`status`), 
  `deleted_at`=0;

-- 根部门
INSERT INTO `admin_department` (`id`, `parent_id`, `name`, `order_num`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, 0, '总部', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 初始化角色：1=super_admin 超级管理员角色，2=admin 业务管理员角色
INSERT INTO `admin_role` (`id`, `name`, `code`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  (1, '超级管理员', 'super_admin', '系统内置最高权限角色，拥有全部权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, 'admin', 'admin', '系统内置业务管理员角色，示例账号使用', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 权限列表（完整，ID从1开始连续）
INSERT INTO `admin_permission` (`id`, `name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  -- 超级权限（通配）
  (1, '全部权限', '*', '超级管理员拥有的全量权限通配符', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 角色管理权限
  (2, '角色列表', 'role:list', '查看角色列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (3, '角色新增', 'role:create', '新增角色', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, '角色编辑', 'role:update', '编辑角色', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (5, '角色删除', 'role:delete', '删除角色', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 权限管理权限
  (6, '权限列表', 'permission:list', '查看权限列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (7, '权限新增', 'permission:create', '新增权限', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (8, '权限编辑', 'permission:update', '编辑权限', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (9, '权限删除', 'permission:delete', '删除权限', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 部门管理权限
  (10, '部门树', 'department:tree', '查看部门树', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (11, '部门新增', 'department:create', '新增部门', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (12, '部门编辑', 'department:update', '编辑部门', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (13, '部门删除', 'department:delete', '删除部门', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 菜单管理权限
  (14, '菜单列表', 'menu:list', '查看菜单列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (15, '菜单新增', 'menu:create', '新增菜单', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (16, '菜单编辑', 'menu:update', '编辑菜单', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (17, '菜单删除', 'menu:delete', '删除菜单', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 用户管理权限
  (18, '用户列表', 'user:list', '查看用户列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (19, '用户新增', 'user:create', '新增用户', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (20, '用户编辑', 'user:update', '编辑用户', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (21, '用户删除', 'user:delete', '删除用户', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 接口管理权限
  (22, '接口列表', 'api:list', '查看接口列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (23, '接口新增', 'api:create', '新增接口', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (24, '接口编辑', 'api:update', '编辑接口', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (25, '接口删除', 'api:delete', '删除接口', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 系统配置权限
  (26, '系统配置列表', 'config:list', '查看系统配置列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (27, '系统配置新增', 'config:create', '新增系统配置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (28, '系统配置编辑', 'config:update', '编辑系统配置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (29, '系统配置删除', 'config:delete', '删除系统配置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典类型权限
  (30, '字典类型列表', 'dict_type:list', '查看字典类型列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (31, '字典类型新增', 'dict_type:create', '新增字典类型', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (32, '字典类型编辑', 'dict_type:update', '编辑字典类型', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (33, '字典类型删除', 'dict_type:delete', '删除字典类型', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典项权限
  (34, '字典项列表', 'dict_item:list', '查看字典项列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (35, '字典项新增', 'dict_item:create', '新增字典项', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (36, '字典项编辑', 'dict_item:update', '编辑字典项', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (37, '字典项删除', 'dict_item:delete', '删除字典项', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 文件管理权限
  (38, '文件列表', 'file:list', '查看文件列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (39, '文件新增', 'file:create', '新增文件', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (40, '文件编辑', 'file:update', '编辑文件', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (41, '文件删除', 'file:delete', '删除文件', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 通用权限
  (42, '个人信息', 'common:profile', '查看个人信息', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (43, '退出登录', 'common:logout', '退出登录接口权限', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (44, '字典查询', 'common:dict', '公共字典查询接口权限', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (45, '刷新缓存', 'common:cache_refresh', '刷新配置/字典缓存', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (46, '我的菜单树', 'menu:my_tree', '获取当前用户的菜单树', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 个人信息相关权限
  (47, '个人信息更新', 'common:profile_update', '更新个人信息（头像、个性签名）', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (48, '修改密码', 'common:password_change', '修改登录密码', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`), 
  `code`=VALUES(`code`), 
  `description`=VALUES(`description`), 
  `updated_at`=UNIX_TIMESTAMP(), 
  `deleted_at`=0;

-- 关联：用户-角色
INSERT INTO `admin_user_role` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`)
VALUES 
  (1, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (2, 2, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- 关联：角色-权限
INSERT INTO `admin_role_permission` (`id`, `role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES 
  (1, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (2, 1, 42, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (3, 1, 47, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- 超级管理员 -> common:profile_update
  (4, 1, 48, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())   -- 超级管理员 -> common:password_change
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- 菜单列表（ID从1开始连续）
INSERT INTO `admin_menu` (`id`, `parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  (1, 0, '仪表盘', '/dashboard', 'Dashboard', 'ele-DataBoard', 2, 1, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, 0, '系统管理', '/system', '', 'ele-Setting', 1, 10, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (3, 2, '角色管理', '/system/role', 'system/RoleList', 'ele-UserFilled', 2, 11, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, 2, '权限管理', '/system/permission', 'system/PermissionList', 'ele-Lock', 2, 12, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (5, 2, '部门管理', '/system/department', 'system/DepartmentList', 'ele-OfficeBuilding', 2, 13, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (6, 2, '菜单管理', '/system/menu', 'system/MenuList', 'ele-Menu', 2, 14, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (7, 2, '用户管理', '/system/user', 'system/UserList', 'ele-UserFilled', 2, 15, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (8, 2, '接口管理', '/system/api', 'system/ApiList', 'ele-Connection', 2, 16, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (9, 0, '临时目录', '/temp', '', 'ele-Folder', 1, 999, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 角色管理按钮（parent_id=3）
  (10, 3, '角色管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (11, 3, '角色管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (12, 3, '角色管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 权限管理按钮（parent_id=4）
  (13, 4, '权限管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (14, 4, '权限管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (15, 4, '权限管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 部门管理按钮（parent_id=5）
  (16, 5, '部门管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (17, 5, '部门管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (18, 5, '部门管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 菜单管理按钮（parent_id=6）
  (19, 6, '菜单管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (20, 6, '菜单管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (21, 6, '菜单管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 用户管理按钮（parent_id=7）
  (22, 7, '用户管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (23, 7, '用户管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (24, 7, '用户管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 接口管理按钮（parent_id=8）
  (25, 8, '接口管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (26, 8, '接口管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (27, 8, '接口管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 系统配置菜单和按钮（parent_id=2，系统管理下）
  (28, 2, '系统配置', '/system/config', 'system/ConfigList', 'ele-Setting', 2, 17, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (29, 28, '系统配置 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (30, 28, '系统配置 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (31, 28, '系统配置 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典类型菜单和按钮（parent_id=2，系统管理下）
  (32, 2, '字典类型', '/system/dict-type', 'system/DictTypeList', 'ele-Notebook', 2, 18, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (33, 32, '字典类型 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (34, 32, '字典类型 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (35, 32, '字典类型 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典项菜单和按钮（parent_id=2，系统管理下）
  (36, 2, '字典项', '/system/dict-item', 'system/DictItemList', 'ele-List', 2, 19, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (37, 36, '字典项 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (38, 36, '字典项 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (39, 36, '字典项 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 文件管理菜单和按钮（parent_id=2，系统管理下）
  (40, 2, '文件管理', '/system/file', 'system/FileList', 'ele-Folder', 2, 20, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (41, 40, '文件管理 新增按钮', '', '', '', 3, 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (42, 40, '文件管理 编辑按钮', '', '', '', 3, 2, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (43, 40, '文件管理 删除按钮', '', '', '', 3, 3, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 权限-菜单关联（ID从1开始连续）
INSERT INTO `admin_permission_menu` (`id`, `permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES
  -- 菜单关联（菜单页面权限）
  -- 角色管理菜单(id=3) -> role:list权限(id=2)
  (1, 2, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 权限管理菜单(id=4) -> permission:list权限(id=6)
  (2, 6, 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 部门管理菜单(id=5) -> department:tree权限(id=10)
  (3, 10, 5, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 菜单管理菜单(id=6) -> menu:list权限(id=14)
  (4, 14, 6, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 用户管理菜单(id=7) -> user:list权限(id=18)
  (5, 18, 7, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 接口管理菜单(id=8) -> api:list权限(id=22)
  (6, 22, 8, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  -- 按钮关联（按钮操作权限）
  -- 角色管理按钮
  (7, 3, 10, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- role:create(id=3) -> 角色管理新增按钮(id=10)
  (8, 4, 11, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- role:update(id=4) -> 角色管理编辑按钮(id=11)
  (9, 5, 12, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- role:delete(id=5) -> 角色管理删除按钮(id=12)
  -- 权限管理按钮
  (10, 7, 13, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:create(id=7) -> 权限管理新增按钮(id=13)
  (11, 8, 14, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限管理编辑按钮(id=14)
  (12, 9, 15, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:delete(id=9) -> 权限管理删除按钮(id=15)
  -- 部门管理按钮
  (13, 11, 16, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:create(id=11) -> 部门管理新增按钮(id=16)
  (14, 12, 17, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:update(id=12) -> 部门管理编辑按钮(id=17)
  (15, 13, 18, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:delete(id=13) -> 部门管理删除按钮(id=18)
  -- 菜单管理按钮
  (16, 15, 19, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:create(id=15) -> 菜单管理新增按钮(id=19)
  (17, 16, 20, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:update(id=16) -> 菜单管理编辑按钮(id=20)
  (18, 17, 21, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:delete(id=17) -> 菜单管理删除按钮(id=21)
  -- 用户管理按钮
  (19, 19, 22, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- user:create(id=19) -> 用户管理新增按钮(id=22)
  (20, 20, 23, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- user:update(id=20) -> 用户管理编辑按钮(id=23)
  (21, 21, 24, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- user:delete(id=21) -> 用户管理删除按钮(id=24)
  -- 接口管理按钮
  (22, 23, 25, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:create(id=23) -> 接口管理新增按钮(id=25)
  (23, 24, 26, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:update(id=24) -> 接口管理编辑按钮(id=26)
  (24, 25, 27, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:delete(id=25) -> 接口管理删除按钮(id=27)
  -- 系统配置菜单和按钮
  (25, 26, 28, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:list(id=26) -> 系统配置菜单(id=28)
  (26, 27, 29, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:create(id=27) -> 系统配置新增按钮(id=29)
  (27, 28, 30, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:update(id=28) -> 系统配置编辑按钮(id=30)
  (28, 29, 31, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:delete(id=29) -> 系统配置删除按钮(id=31)
  -- 数据字典类型菜单和按钮
  (29, 30, 32, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:list(id=30) -> 字典类型菜单(id=32)
  (30, 31, 33, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:create(id=31) -> 字典类型新增按钮(id=33)
  (31, 32, 34, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:update(id=32) -> 字典类型编辑按钮(id=34)
  (32, 33, 35, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:delete(id=33) -> 字典类型删除按钮(id=35)
  -- 数据字典项菜单和按钮
  (33, 34, 36, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:list(id=34) -> 字典项菜单(id=36)
  (34, 35, 37, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:create(id=35) -> 字典项新增按钮(id=37)
  (35, 36, 38, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:update(id=36) -> 字典项编辑按钮(id=38)
  (36, 37, 39, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:delete(id=37) -> 字典项删除按钮(id=39)
  -- 文件管理菜单和按钮
  (37, 38, 40, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:list(id=38) -> 文件管理菜单(id=40)
  (38, 39, 41, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:create(id=39) -> 文件管理新增按钮(id=41)
  (39, 40, 42, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:update(id=40) -> 文件管理编辑按钮(id=42)
  (40, 41, 43, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())  -- file:delete(id=41) -> 文件管理删除按钮(id=43)
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- 接口列表（所有业务接口，ID从1开始连续）
INSERT INTO `admin_api` (`id`, `name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  -- 认证相关接口（无需权限，但需要认证）
  (1, '登出', 'POST', '/api/v1/logout', '用户登出', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, '个人信息', 'GET', '/api/v1/profile', '获取当前用户个人信息', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 用户管理接口
  (3, '用户列表', 'GET', '/api/v1/users', '获取用户列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, '用户新增', 'POST', '/api/v1/users', '新增用户', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (5, '用户编辑', 'PUT', '/api/v1/users/:id', '编辑用户', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (6, '用户删除', 'DELETE', '/api/v1/users/:id', '删除用户', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (7, '用户角色列表', 'GET', '/api/v1/users/:userId/roles', '获取用户关联的角色列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (8, '用户角色更新', 'PUT', '/api/v1/users/:userId/roles', '更新用户关联的角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 角色管理接口
  (9, '角色列表', 'GET', '/api/v1/roles', '获取角色列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (10, '角色新增', 'POST', '/api/v1/roles', '新增角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (11, '角色编辑', 'PUT', '/api/v1/roles/:id', '编辑角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (12, '角色删除', 'DELETE', '/api/v1/roles/:id', '删除角色', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (13, '角色权限列表', 'GET', '/api/v1/roles/:roleId/permissions', '获取角色关联的权限列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (14, '角色权限更新', 'PUT', '/api/v1/roles/:roleId/permissions', '更新角色关联的权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 权限管理接口
  (15, '权限列表', 'GET', '/api/v1/permissions', '获取权限列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (16, '权限新增', 'POST', '/api/v1/permissions', '新增权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (17, '权限编辑', 'PUT', '/api/v1/permissions/:id', '编辑权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (18, '权限删除', 'DELETE', '/api/v1/permissions/:id', '删除权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (19, '权限菜单列表', 'GET', '/api/v1/permissions/:permissionId/menus', '获取权限关联的菜单列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (20, '权限菜单更新', 'PUT', '/api/v1/permissions/:permissionId/menus', '更新权限关联的菜单', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (21, '权限接口列表', 'GET', '/api/v1/permissions/:permissionId/apis', '获取权限关联的接口列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (22, '权限接口更新', 'PUT', '/api/v1/permissions/:permissionId/apis', '更新权限关联的接口', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 部门管理接口
  (23, '部门树', 'GET', '/api/v1/departments/tree', '获取部门树', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (24, '部门新增', 'POST', '/api/v1/departments', '新增部门', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (25, '部门编辑', 'PUT', '/api/v1/departments/:id', '编辑部门', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (26, '部门删除', 'DELETE', '/api/v1/departments/:id', '删除部门', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 菜单管理接口
  (27, '菜单树', 'GET', '/api/v1/menus/tree', '获取菜单树', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (28, '我的菜单树', 'GET', '/api/v1/menus/my-tree', '获取当前用户的菜单树', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (29, '菜单新增', 'POST', '/api/v1/menus', '新增菜单', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (30, '菜单编辑', 'PUT', '/api/v1/menus/:id', '编辑菜单', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (31, '菜单删除', 'DELETE', '/api/v1/menus/:id', '删除菜单', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 接口管理接口
  (32, '接口列表', 'GET', '/api/v1/apis', '获取接口列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (33, '接口新增', 'POST', '/api/v1/apis', '新增接口', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (34, '接口编辑', 'PUT', '/api/v1/apis/:id', '编辑接口', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (35, '接口删除', 'DELETE', '/api/v1/apis/:id', '删除接口', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 系统配置接口
  (36, '系统配置列表', 'GET', '/api/v1/configs', '获取系统配置列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (37, '系统配置查询', 'GET', '/api/v1/configs/get', '查询系统配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (38, '系统配置新增', 'POST', '/api/v1/configs', '新增系统配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (39, '系统配置编辑', 'PUT', '/api/v1/configs', '编辑系统配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (40, '系统配置删除', 'DELETE', '/api/v1/configs', '删除系统配置', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典类型接口
  (41, '字典类型列表', 'GET', '/api/v1/dict-types', '获取字典类型列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (42, '字典类型新增', 'POST', '/api/v1/dict-types', '新增字典类型', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (43, '字典类型编辑', 'PUT', '/api/v1/dict-types', '编辑字典类型', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (44, '字典类型删除', 'DELETE', '/api/v1/dict-types', '删除字典类型', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 数据字典项接口
  (45, '字典项列表', 'GET', '/api/v1/dict-items', '获取字典项列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (46, '字典项新增', 'POST', '/api/v1/dict-items', '新增字典项', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (47, '字典项编辑', 'PUT', '/api/v1/dict-items', '编辑字典项', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (48, '字典项删除', 'DELETE', '/api/v1/dict-items', '删除字典项', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 字典查询接口（无需权限）
  (49, '字典查询', 'GET', '/api/v1/dict', '根据编码查询字典项', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 文件管理接口
  (50, '文件列表', 'GET', '/api/v1/files', '获取文件列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (51, '文件新增', 'POST', '/api/v1/files', '新增文件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (52, '文件编辑', 'PUT', '/api/v1/files', '编辑文件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (53, '文件删除', 'DELETE', '/api/v1/files', '删除文件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (54, '文件上传', 'POST', '/api/v1/files/upload', '上传文件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (55, '文件下载', 'GET', '/api/v1/files/:id/download', '下载文件', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 缓存刷新接口
  (56, '刷新缓存', 'POST', '/api/v1/cache/refresh', '刷新配置与字典缓存', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 个人信息相关接口
  (57, '个人信息更新', 'PUT', '/api/v1/profile', '更新当前用户个人信息（头像、个性签名）', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (58, '修改密码', 'POST', '/api/v1/profile/password', '修改当前用户登录密码', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 权限-接口关联（所有权限与接口的关联，ID从1开始连续）
INSERT INTO `admin_permission_api` (`id`, `permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES
  -- 通用接口权限
  (53, 42, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),   -- common:profile(id=42) -> 个人信息(api_id=2)
  (54, 43, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),   -- common:logout(id=43) -> 登出(api_id=1)
  (55, 44, 49, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- common:dict(id=44) -> 字典查询(api_id=49)
  (56, 45, 56, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- common:cache_refresh(id=45) -> 刷新缓存(api_id=56)
  (57, 47, 57, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- common:profile_update(id=47) -> 个人信息更新(api_id=57)
  (58, 48, 58, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- common:password_change(id=48) -> 修改密码(api_id=58)
  -- 用户管理权限
  (1, 18, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:list(id=18) -> 用户列表(api_id=3)
  (2, 19, 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:create(id=19) -> 用户新增(api_id=4)
  (3, 20, 5, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:update(id=20) -> 用户编辑(api_id=5)
  (4, 21, 6, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:delete(id=21) -> 用户删除(api_id=6)
  (5, 20, 7, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:update(id=20) -> 用户角色列表(api_id=7)
  (6, 20, 8, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- user:update(id=20) -> 用户角色更新(api_id=8)
  -- 角色管理权限
  (7, 2, 9, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),   -- role:list(id=2) -> 角色列表(api_id=9)
  (8, 3, 10, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- role:create(id=3) -> 角色新增(api_id=10)
  (9, 4, 11, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),  -- role:update(id=4) -> 角色编辑(api_id=11)
  (10, 5, 12, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- role:delete(id=5) -> 角色删除(api_id=12)
  (11, 4, 13, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- role:update(id=4) -> 角色权限列表(api_id=13)
  (12, 4, 14, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- role:update(id=4) -> 角色权限更新(api_id=14)
  -- 权限管理权限
  (13, 6, 15, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:list(id=6) -> 权限列表(api_id=15)
  (14, 7, 16, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:create(id=7) -> 权限新增(api_id=16)
  (15, 8, 17, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限编辑(api_id=17)
  (16, 9, 18, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:delete(id=9) -> 权限删除(api_id=18)
  (17, 8, 19, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限菜单列表(api_id=19)
  (18, 8, 20, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限菜单更新(api_id=20)
  (19, 8, 21, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限接口列表(api_id=21)
  (20, 8, 22, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- permission:update(id=8) -> 权限接口更新(api_id=22)
  -- 部门管理权限
  (21, 10, 23, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:tree(id=10) -> 部门树(api_id=23)
  (22, 11, 24, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:create(id=11) -> 部门新增(api_id=24)
  (23, 12, 25, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:update(id=12) -> 部门编辑(api_id=25)
  (24, 13, 26, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- department:delete(id=13) -> 部门删除(api_id=26)
  -- 菜单管理权限
  (25, 14, 27, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:list(id=14) -> 菜单树(api_id=27)
  (26, 46, 28, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:my_tree(id=46) -> 我的菜单树(api_id=28)
  (27, 15, 29, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:create(id=15) -> 菜单新增(api_id=29)
  (28, 16, 30, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:update(id=16) -> 菜单编辑(api_id=30)
  (29, 17, 31, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:delete(id=17) -> 菜单删除(api_id=31)
  -- 接口管理权限
  (30, 22, 32, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:list(id=22) -> 接口列表(api_id=32)
  (31, 23, 33, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:create(id=23) -> 接口新增(api_id=33)
  (32, 24, 34, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:update(id=24) -> 接口编辑(api_id=34)
  (33, 25, 35, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:delete(id=25) -> 接口删除(api_id=35)
  -- 系统配置权限
  (34, 26, 36, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:list(id=26) -> 系统配置列表(api_id=36)
  (35, 26, 37, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:list(id=26) -> 系统配置查询(api_id=37)
  (36, 27, 38, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:create(id=27) -> 系统配置新增(api_id=38)
  (37, 28, 39, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:update(id=28) -> 系统配置编辑(api_id=39)
  (38, 29, 40, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- config:delete(id=29) -> 系统配置删除(api_id=40)
  -- 数据字典类型权限
  (39, 30, 41, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:list(id=30) -> 字典类型列表(api_id=41)
  (40, 31, 42, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:create(id=31) -> 字典类型新增(api_id=42)
  (41, 32, 43, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:update(id=32) -> 字典类型编辑(api_id=43)
  (42, 33, 44, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_type:delete(id=33) -> 字典类型删除(api_id=44)
  -- 数据字典项权限
  (43, 34, 45, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:list(id=34) -> 字典项列表(api_id=45)
  (44, 35, 46, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:create(id=35) -> 字典项新增(api_id=46)
  (45, 36, 47, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:update(id=36) -> 字典项编辑(api_id=47)
  (46, 37, 48, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- dict_item:delete(id=37) -> 字典项删除(api_id=48)
  -- 文件管理权限
  (47, 38, 50, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:list(id=38) -> 文件列表(api_id=50)
  (48, 39, 51, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:create(id=39) -> 文件新增(api_id=51)
  (49, 40, 52, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:update(id=40) -> 文件编辑(api_id=52)
  (50, 41, 53, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:delete(id=41) -> 文件删除(api_id=53)
  (51, 39, 54, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- file:create(id=39) -> 文件上传(api_id=54)
  (52, 38, 55, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())  -- file:list(id=38) -> 文件下载(api_id=55)
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- ============================================
-- 2. 其他初始化业务数据（配置、字典等）
-- ============================================

-- 系统配置初始化数据
INSERT INTO `admin_config` (`id`, `group`, `key`, `value`, `type`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  (1, 'system', 'system:app_name', '"后台管理系统"', 'string', '应用名称', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, 'system', 'system:app_logo', '"/static/logo.png"', 'string', '应用Logo路径', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (3, 'system', 'system:app_version', '"1.0.0"', 'string', '应用版本', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, 'system', 'system:timeout', '300', 'number', '会话超时时间（秒）', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (5, 'theme', 'theme:primary_color', '"#409EFF"', 'string', '主题主色', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (6, 'theme', 'theme:sidebar_width', '200', 'number', '侧边栏宽度（px）', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (7, 'upload', 'upload:max_size', '10485760', 'number', '最大上传文件大小（字节，默认10MB）', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (8, 'upload', 'upload:allowed_types', '["jpg","jpeg","png","gif","pdf","doc","docx","xls","xlsx"]', 'json', '允许上传的文件类型', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `value`=VALUES(`value`), `updated_at`=UNIX_TIMESTAMP(), `deleted_at`=0;

-- 数据字典类型初始化数据
INSERT INTO `admin_dict_type` (`id`, `name`, `code`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  (1, '用户状态', 'user_status', '用户账号状态字典', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, '性别', 'gender', '性别字典', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (3, '是否', 'yes_no', '是否字典', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, '文件存储类型', 'file_storage_type', '文件存储类型字典', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 数据字典项初始化数据
INSERT INTO `admin_dict_item` (`id`, `type_id`, `label`, `value`, `sort`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`)
VALUES
  -- 用户状态字典项
  (1, 1, '启用', '1', 1, 1, '用户账号启用状态', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (2, 1, '禁用', '0', 2, 1, '用户账号禁用状态', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 性别字典项
  (3, 2, '男', '1', 1, 1, '男性', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (4, 2, '女', '2', 2, 1, '女性', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (5, 2, '未知', '0', 3, 1, '未知性别', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 是否字典项
  (6, 3, '是', '1', 1, 1, '是', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (7, 3, '否', '0', 2, 1, '否', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  -- 文件存储类型字典项
  (8, 4, '本地存储', 'local', 1, 1, '本地文件系统存储', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (9, 4, 'OSS存储', 'oss', 2, 1, '阿里云OSS存储', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  (10, 4, 'S3存储', 's3', 3, 1, 'AWS S3存储', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- ============================================
-- 3. 日志与监控相关模块初始化数据（归类到系统管理）
-- ============================================

-- 获取系统管理目录菜单ID（path = '/system'）
SET @system_menu_id = (SELECT `id` FROM `admin_menu` WHERE `path` = '/system' AND `deleted_at` = 0 LIMIT 1);

-- ==========================
-- 3.1 操作日志模块
-- ==========================
-- 操作日志主菜单（系统管理下）
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @system_menu_id,
    '操作日志',
    '/system/operation-log',
    'system/OperationLogList',
    'ele-Document',
    2, -- 类型：2 菜单
    30, -- 排序值
    1, -- 是否可见：1 是
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `component`=VALUES(`component`),
    `icon`=VALUES(`icon`),
    `type`=VALUES(`type`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @operation_menu_id = (
  SELECT `id` FROM `admin_menu` 
  WHERE `path` = '/system/operation-log' AND `deleted_at` = 0 
  LIMIT 1
);

-- 操作日志导出按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @operation_menu_id,
    '操作日志 导出按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    1, -- 排序值
    0, -- 是否可见：0 否（按钮不显示在菜单中）
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @operation_export_button_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `parent_id` = @operation_menu_id 
    AND `name` = '操作日志 导出按钮'
    AND `deleted_at` = 0
  LIMIT 1
);

-- 操作日志权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('操作日志列表', 'operation_log:list', '查看操作日志列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('操作日志详情', 'operation_log:detail', '查看操作日志详情', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('操作日志导出', 'operation_log:export', '导出操作日志', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @operation_list_permission_id = (
  SELECT `id` FROM `admin_permission` 
  WHERE `code` = 'operation_log:list' AND `deleted_at` = 0 
  LIMIT 1
);
SET @operation_detail_permission_id = (
  SELECT `id` FROM `admin_permission` 
  WHERE `code` = 'operation_log:detail' AND `deleted_at` = 0 
  LIMIT 1
);
SET @operation_export_permission_id = (
  SELECT `id` FROM `admin_permission` 
  WHERE `code` = 'operation_log:export' AND `deleted_at` = 0 
  LIMIT 1
);

-- 操作日志接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('操作日志列表', 'GET', '/api/v1/operation-logs', '获取操作日志列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('操作日志详情', 'GET', '/api/v1/operation-logs/:id', '获取操作日志详情', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('操作日志导出', 'GET', '/api/v1/operation-logs/export', '导出操作日志', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `status`=VALUES(`status`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @operation_list_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/operation-logs' AND `deleted_at` = 0
  LIMIT 1
);
SET @operation_detail_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/operation-logs/:id' AND `deleted_at` = 0
  LIMIT 1
);
SET @operation_export_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/operation-logs/export' AND `deleted_at` = 0
  LIMIT 1
);

-- 操作日志 权限-菜单 关联
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES 
  (@operation_list_permission_id, @operation_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@operation_export_permission_id, @operation_export_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 操作日志 权限-接口 关联
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES 
  (@operation_list_permission_id, @operation_list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@operation_detail_permission_id, @operation_detail_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@operation_export_permission_id, @operation_export_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 超级管理员角色关联操作日志权限（role_id = 1）
INSERT INTO `admin_role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES
  (1, @operation_list_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @operation_detail_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @operation_export_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- ==========================
-- 3.2 登录日志模块
-- ==========================
-- 登录日志主菜单（系统管理下）
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @system_menu_id,
    '登录日志',
    '/system/login-log',
    'system/LoginLogList',
    'ele-Document',
    2, -- 类型：2 菜单
    31, -- 排序值
    1, -- 是否可见：1 是
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `component`=VALUES(`component`),
    `icon`=VALUES(`icon`),
    `type`=VALUES(`type`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @login_menu_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `path` = '/system/login-log' AND `deleted_at` = 0
  LIMIT 1
);

-- 登录日志详情按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @login_menu_id,
    '登录日志 详情按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    1, -- 排序值
    0, -- 是否可见：0 否
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @login_detail_button_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `parent_id` = @login_menu_id 
    AND `name` = '登录日志 详情按钮'
    AND `deleted_at` = 0
  LIMIT 1
);

-- 登录日志导出按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @login_menu_id,
    '登录日志 导出按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    2, -- 排序值
    0, -- 是否可见：0 否
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @login_export_button_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `parent_id` = @login_menu_id 
    AND `name` = '登录日志 导出按钮'
    AND `deleted_at` = 0
  LIMIT 1
);

-- 登录日志权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('登录日志列表', 'login_log:list', '查看登录日志列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('登录日志详情', 'login_log:detail', '查看登录日志详情', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('登录日志导出', 'login_log:export', '导出登录日志', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @login_list_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'login_log:list' AND `deleted_at` = 0
  LIMIT 1
);
SET @login_detail_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'login_log:detail' AND `deleted_at` = 0
  LIMIT 1
);
SET @login_export_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'login_log:export' AND `deleted_at` = 0
  LIMIT 1
);

-- 登录日志接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('登录日志列表', 'GET', '/api/v1/login-logs', '获取登录日志列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('登录日志详情', 'GET', '/api/v1/login-logs/:id', '获取登录日志详情', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('登录日志导出', 'GET', '/api/v1/login-logs/export', '导出登录日志', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `status`=VALUES(`status`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @login_list_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/login-logs' AND `deleted_at` = 0
  LIMIT 1
);
SET @login_detail_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/login-logs/:id' AND `deleted_at` = 0
  LIMIT 1
);
SET @login_export_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/login-logs/export' AND `deleted_at` = 0
  LIMIT 1
);

-- 登录日志 权限-菜单 关联
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES 
  (@login_list_permission_id, @login_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@login_detail_permission_id, @login_detail_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@login_export_permission_id, @login_export_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 登录日志 权限-接口 关联
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES 
  (@login_list_permission_id, @login_list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@login_detail_permission_id, @login_detail_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@login_export_permission_id, @login_export_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 超级管理员角色关联登录日志权限（role_id = 1）
INSERT INTO `admin_role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES
  (1, @login_list_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @login_detail_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @login_export_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- ==========================
-- 3.3 审计日志模块
-- ==========================
-- 审计日志主菜单（系统管理下）
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @system_menu_id,
    '审计日志',
    '/system/audit-log',
    'system/AuditLogList',
    'ele-Document',
    2, -- 类型：2 菜单
    32, -- 排序值
    1, -- 是否可见：1 是
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `component`=VALUES(`component`),
    `icon`=VALUES(`icon`),
    `type`=VALUES(`type`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @audit_menu_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `path` = '/system/audit-log' AND `deleted_at` = 0
  LIMIT 1
);

-- 审计日志导出按钮
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @audit_menu_id,
    '审计日志 导出按钮',
    '',
    '',
    '',
    3, -- 类型：3 按钮
    1, -- 排序值
    0, -- 是否可见：0 否（按钮不显示在菜单中）
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @audit_export_button_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `parent_id` = @audit_menu_id 
    AND `name` = '审计日志 导出按钮'
    AND `deleted_at` = 0
  LIMIT 1
);

-- 审计日志权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('审计日志列表', 'audit_log:list', '查看审计日志列表', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('审计日志详情', 'audit_log:detail', '查看审计日志详情', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('审计日志导出', 'audit_log:export', '导出审计日志', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @audit_list_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'audit_log:list' AND `deleted_at` = 0
  LIMIT 1
);
SET @audit_detail_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'audit_log:detail' AND `deleted_at` = 0
  LIMIT 1
);
SET @audit_export_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'audit_log:export' AND `deleted_at` = 0
  LIMIT 1
);

-- 审计日志接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('审计日志列表', 'GET', '/api/v1/audit-logs', '获取审计日志列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('审计日志详情', 'GET', '/api/v1/audit-logs/:id', '获取审计日志详情', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('审计日志导出', 'GET', '/api/v1/audit-logs/export', '导出审计日志', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `status`=VALUES(`status`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @audit_list_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/audit-logs' AND `deleted_at` = 0
  LIMIT 1
);
SET @audit_detail_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/audit-logs/:id' AND `deleted_at` = 0
  LIMIT 1
);
SET @audit_export_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/audit-logs/export' AND `deleted_at` = 0
  LIMIT 1
);

-- 审计日志 权限-菜单 关联
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES 
  (@audit_list_permission_id, @audit_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@audit_export_permission_id, @audit_export_button_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 审计日志 权限-接口 关联
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES 
  (@audit_list_permission_id, @audit_list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@audit_detail_permission_id, @audit_detail_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@audit_export_permission_id, @audit_export_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 超级管理员角色关联审计日志权限（role_id = 1）
INSERT INTO `admin_role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES
  (1, @audit_list_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @audit_detail_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (1, @audit_export_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- ==========================
-- 3.4 性能监控日志模块
-- ==========================
-- 性能监控日志主菜单（系统管理下）
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @system_menu_id,
    '性能监控日志',
    '/system/performance-log',
    'system/PerformanceLogList',
    'ele-Document',
    2, -- 类型：2 菜单
    33, -- 排序值
    1, -- 是否可见：1 是
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `component`=VALUES(`component`),
    `icon`=VALUES(`icon`),
    `type`=VALUES(`type`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @performance_menu_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `path` = '/system/performance-log' AND `deleted_at` = 0
  LIMIT 1
);

-- 性能监控日志列表权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '性能监控日志列表',
    'performance_log:list',
    '查看性能监控日志列表',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @performance_list_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'performance_log:list' AND `deleted_at` = 0
  LIMIT 1
);

-- 性能监控日志列表接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '性能监控日志列表',
    'GET',
    '/api/v1/performance-logs',
    '获取性能监控日志列表',
    1,
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `status`=VALUES(`status`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @performance_list_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/performance-logs' AND `deleted_at` = 0
  LIMIT 1
);

-- 性能监控日志 权限-菜单 关联
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@performance_list_permission_id, @performance_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 性能监控日志 权限-接口 关联
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES (@performance_list_permission_id, @performance_list_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 超级管理员角色关联性能监控日志权限（role_id = 1）
INSERT INTO `admin_role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES (1, @performance_list_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- ==========================
-- 3.5 系统监控模块
-- ==========================
-- 系统监控主菜单（系统管理下）
INSERT INTO `admin_menu` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `order_num`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    @system_menu_id,
    '系统监控',
    '/system/monitor',
    'system/MonitorList',
    'ele-Monitor',
    2, -- 类型：2 菜单
    34, -- 排序值
    1, -- 是否可见：1 是
    1, -- 状态：1 启用
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
    `name`=VALUES(`name`),
    `component`=VALUES(`component`),
    `icon`=VALUES(`icon`),
    `type`=VALUES(`type`),
    `order_num`=VALUES(`order_num`),
    `visible`=VALUES(`visible`),
    `status`=VALUES(`status`),
    `updated_at`=UNIX_TIMESTAMP(),
    `deleted_at`=0;
SET @monitor_menu_id = (
  SELECT `id` FROM `admin_menu`
  WHERE `path` = '/system/monitor' AND `deleted_at` = 0
  LIMIT 1
);

-- 系统监控查看权限
INSERT INTO `admin_permission` (`name`, `code`, `description`, `created_at`, `updated_at`, `deleted_at`)
VALUES (
    '系统监控查看',
    'monitor:view',
    '查看系统监控信息',
    UNIX_TIMESTAMP(),
    UNIX_TIMESTAMP(),
    0
)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @monitor_view_permission_id = (
  SELECT `id` FROM `admin_permission`
  WHERE `code` = 'monitor:view' AND `deleted_at` = 0
  LIMIT 1
);

-- 系统监控接口
INSERT INTO `admin_api` (`name`, `method`, `path`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES 
  ('系统监控状态', 'GET', '/api/v1/monitor/status', '获取系统资源使用情况（CPU、内存、磁盘、网络）', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
  ('系统统计', 'GET', '/api/v1/monitor/stats', '获取系统统计数据（用户数、角色数、权限数等）', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`),
  `description`=VALUES(`description`),
  `status`=VALUES(`status`),
  `updated_at`=UNIX_TIMESTAMP(),
  `deleted_at`=0;
SET @monitor_status_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/monitor/status' AND `deleted_at` = 0
  LIMIT 1
);
SET @monitor_stats_api_id = (
  SELECT `id` FROM `admin_api`
  WHERE `method` = 'GET' AND `path` = '/api/v1/monitor/stats' AND `deleted_at` = 0
  LIMIT 1
);

-- 系统监控 权限-菜单 关联
INSERT INTO `admin_permission_menu` (`permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES (@monitor_view_permission_id, @monitor_menu_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 系统监控 权限-接口 关联
INSERT INTO `admin_permission_api` (`permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES 
  (@monitor_view_permission_id, @monitor_status_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
  (@monitor_view_permission_id, @monitor_stats_api_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- 超级管理员角色关联系统监控权限（role_id = 1）
INSERT INTO `admin_role_permission` (`role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES (1, @monitor_view_permission_id, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at` = UNIX_TIMESTAMP();

-- ============================================
-- 4. 保护初始化数据不被删除（触发器）
-- ============================================
-- 注意：触发器只能阻止软删除（UPDATE deleted_at），硬删除（DELETE）需要在业务代码中检查

-- 保护超级管理员用户不被软删除
DROP TRIGGER IF EXISTS trg_block_delete_init_user;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_user BEFORE UPDATE ON admin_user
FOR EACH ROW
BEGIN
    IF OLD.id = 1 AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护超级管理员角色不被软删除
DROP TRIGGER IF EXISTS trg_block_delete_init_role;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_role BEFORE UPDATE ON admin_role
FOR EACH ROW
BEGIN
    IF OLD.id = 1 AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护初始化权限不被软删除（id=1-46）
DROP TRIGGER IF EXISTS trg_block_delete_init_permission;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_permission BEFORE UPDATE ON admin_permission
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 46 
       AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护根部门不被软删除
DROP TRIGGER IF EXISTS trg_block_delete_init_department;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_department BEFORE UPDATE ON admin_department
FOR EACH ROW
BEGIN
    IF OLD.id = 1 AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护初始化菜单不被软删除（id=1-43，包含13个菜单和30个按钮）
DROP TRIGGER IF EXISTS trg_block_delete_init_menu;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_menu BEFORE UPDATE ON admin_menu
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 43 
       AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护初始化接口不被软删除（id=1-56）
DROP TRIGGER IF EXISTS trg_block_delete_init_api;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_api BEFORE UPDATE ON admin_api
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 56 
       AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护初始化关联数据不被物理删除
DROP TRIGGER IF EXISTS trg_block_delete_init_user_role;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_user_role BEFORE DELETE ON admin_user_role
FOR EACH ROW
BEGIN
    IF OLD.id = 1 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_block_delete_init_role_permission;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_role_permission BEFORE DELETE ON admin_role_permission
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 2 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

