-- admin-server 数据库初始化数据脚本
-- 注意：此脚本中的数据为系统初始化数据，不可被删除（包括软删、硬删）
-- 初始化数据的ID范围（每张表从1开始连续）：
--   admin_user: id=1 (超级管理员)
--   admin_role: id=1 (超级管理员角色)
--   admin_permission: id=1-25 (25个权限)
--   admin_department: id=1 (根部门)
--   admin_menu: id=1-9 (9个菜单)
--   admin_api: id=1-35 (35个接口)
--   admin_user_role: id=1 (超级管理员用户-角色关联)
--   admin_role_permission: id=1 (超级管理员角色-权限关联)
--   admin_permission_menu: id=1-6 (6个权限-菜单关联)
--   admin_permission_api: id=1-33 (33个权限-接口关联)

-- ============================================
-- 1. 初始化基础数据
-- ============================================
-- 超级管理员：用户名 oldbai，密码 oldbai（bcrypt 哈希）
INSERT INTO `admin_user` (`id`, `username`, `password_hash`, `department_id`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, 'oldbai', '$2a$10$TIjB8/yhHDiyNbJn40BUPOACjxeTccaYTD4Ot3p00ZBCKzh7/sL9q', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `username`=VALUES(`username`), `password_hash`=VALUES(`password_hash`), `status`=1, `deleted_at`=0;

-- 根部门
INSERT INTO `admin_department` (`id`, `parent_id`, `name`, `order_num`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, 0, '总部', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 超级管理员角色
INSERT INTO `admin_role` (`id`, `name`, `code`, `description`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, '超级管理员', 'super_admin', '系统内置最高权限角色，拥有全部权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
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
  (25, '接口删除', 'api:delete', '删除接口', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE 
  `name`=VALUES(`name`), 
  `code`=VALUES(`code`), 
  `description`=VALUES(`description`), 
  `updated_at`=UNIX_TIMESTAMP(), 
  `deleted_at`=0;

-- 关联：用户-角色
INSERT INTO `admin_user_role` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`)
VALUES (1, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- 关联：角色-权限
INSERT INTO `admin_role_permission` (`id`, `role_id`, `permission_id`, `created_at`, `updated_at`)
VALUES (1, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- 基础菜单（ID从1开始连续）
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
  (9, 0, '临时目录', '/temp', '', 'ele-Folder', 1, 999, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 权限-菜单关联（ID从1开始连续）
INSERT INTO `admin_permission_menu` (`id`, `permission_id`, `menu_id`, `created_at`, `updated_at`)
VALUES
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
  (6, 22, 8, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
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
  (35, '接口删除', 'DELETE', '/api/v1/apis/:id', '删除接口', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `deleted_at`=0;

-- 权限-接口关联（所有权限与接口的关联，ID从1开始连续）
INSERT INTO `admin_permission_api` (`id`, `permission_id`, `api_id`, `created_at`, `updated_at`)
VALUES
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
  (26, 14, 28, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:list(id=14) -> 我的菜单树(api_id=28)
  (27, 15, 29, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:create(id=15) -> 菜单新增(api_id=29)
  (28, 16, 30, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:update(id=16) -> 菜单编辑(api_id=30)
  (29, 17, 31, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- menu:delete(id=17) -> 菜单删除(api_id=31)
  -- 接口管理权限
  (30, 22, 32, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:list(id=22) -> 接口列表(api_id=32)
  (31, 23, 33, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:create(id=23) -> 接口新增(api_id=33)
  (32, 24, 34, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()), -- api:update(id=24) -> 接口编辑(api_id=34)
  (33, 25, 35, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())  -- api:delete(id=25) -> 接口删除(api_id=35)
ON DUPLICATE KEY UPDATE `updated_at`=UNIX_TIMESTAMP();

-- ============================================
-- 2. 保护初始化数据不被删除（触发器）
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

-- 保护初始化权限不被软删除（id=1-25）
DROP TRIGGER IF EXISTS trg_block_delete_init_permission;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_permission BEFORE UPDATE ON admin_permission
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 25 
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

-- 保护初始化菜单不被软删除（id=1-9）
DROP TRIGGER IF EXISTS trg_block_delete_init_menu;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_menu BEFORE UPDATE ON admin_menu
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 9 
       AND NEW.deleted_at <> OLD.deleted_at AND NEW.deleted_at <> 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

-- 保护初始化接口不被软删除（id=1-35）
DROP TRIGGER IF EXISTS trg_block_delete_init_api;
DELIMITER $$
CREATE TRIGGER trg_block_delete_init_api BEFORE UPDATE ON admin_api
FOR EACH ROW
BEGIN
    IF OLD.id BETWEEN 1 AND 35 
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
    IF OLD.id = 1 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '初始化数据不可删除';
    END IF;
END$$
DELIMITER ;

