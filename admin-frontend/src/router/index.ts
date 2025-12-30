import {createRouter, createWebHashHistory, RouteRecordRaw} from 'vue-router';
import {useUserStore} from '@/stores/user';
import {usePermission} from '@/hooks/usePermission';
import type {MenuItem} from '@/api/generated/admin';

const viewModules = import.meta.glob('../views/**/*.vue');

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/',
    name: 'Root',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {path: '', name: 'RootRedirect', redirect: '/dashboard'},
      {
        path: '/dashboard',
        name: 'Dashboard',
        meta: {keepAlive: true},
        component: () => import('@/views/Dashboard.vue')
      },
      {
        path: '/system/role',
        name: 'RoleList',
        meta: {permission: 'role:list', keepAlive: true},
        component: () => import('@/views/system/RoleList.vue')
      },
      {
        path: '/system/permission',
        name: 'PermissionList',
        meta: {permission: 'permission:list', keepAlive: true},
        component: () => import('@/views/system/PermissionList.vue')
      },
      {
        path: '/system/department',
        name: 'DepartmentList',
        meta: {permission: 'department:tree', keepAlive: true},
        component: () => import('@/views/system/DepartmentList.vue')
      },
      {
        path: '/system/api',
        name: 'ApiList',
        meta: {permission: 'api:list', keepAlive: true},
        component: () => import('@/views/system/ApiList.vue')
      },
      {
        path: '/system/profile',
        name: 'Profile',
        meta: {keepAlive: true},
        component: () => import('@/views/system/Profile.vue')
      },
      {
        path: '/403',
        name: 'NoAccess',
        component: () => import('@/views/error/NoAccess.vue')
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/NotFound.vue')
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

// 添加全局错误处理
router.onError((error) => {
  console.error('[Router] 路由错误:', error);
});

let initialized = false;
const dynamicAdded = new Set<string>();

function resolveComponent(component?: string, path?: string) {
  const candidates: string[] = [];
  if (component) {
    candidates.push(component.replace(/^\//, ''));
  }
  if (path) {
    candidates.push(path.replace(/^\//, ''));
  }
  for (const [key, loader] of Object.entries(viewModules)) {
    const clean = key.replace(/^..\/views\//, '').replace(/\.vue$/, '');
    if (candidates.includes(clean)) {
      return loader;
    }
  }
  // 如果找不到组件，输出错误信息
  if (component || path) {
    console.error(`[Router] 无法解析组件: component="${component}", path="${path}"`);
  }
  return undefined;
}

function buildRoutesFromMenus(menus: MenuItem[]): RouteRecordRaw[] {
  const res: RouteRecordRaw[] = [];
  const usedNames = new Set<string>();
  
  const walk = (items: MenuItem[]) => {
    items.forEach((m) => {
      // 处理菜单（type=2）：有实际页面组件
      if (m.type === 2 && m.path) {
        const comp = resolveComponent(m.component, m.path);
        if (comp) {
          // 生成唯一的路由名称
          let routeName = m.path.replace(/^\//, '').replace(/\//g, '_');
          if (routeName === '') {
            routeName = 'root';
          }
          // 确保名称唯一
          let uniqueName = routeName;
          let counter = 1;
          while (usedNames.has(uniqueName)) {
            uniqueName = `${routeName}_${counter}`;
            counter++;
          }
          usedNames.add(uniqueName);
          
          res.push({
            path: m.path,
            name: uniqueName,
            meta: {
              permission: m.permissionCode || undefined, // 如果没有权限码，设为 undefined
              keepAlive: true
            },
            component: comp
          });
        } else {
          console.error(`[Router] 路由注册失败: path="${m.path}", component="${m.component}"`);
        }
      }
      // 处理目录（type=1）：如果有子菜单，重定向到第一个子菜单
      if (m.type === 1 && m.path && m.children && m.children.length > 0) {
        // 找到第一个有效的子菜单
        const firstChild = m.children.find((child) => child.type === 2 && child.path);
        if (firstChild && firstChild.path) {
          // 生成唯一的路由名称
          let routeName = m.path.replace(/^\//, '').replace(/\//g, '_');
          if (routeName === '') {
            routeName = 'root';
          }
          // 确保名称唯一
          let uniqueName = routeName;
          let counter = 1;
          while (usedNames.has(uniqueName)) {
            uniqueName = `${routeName}_${counter}`;
            counter++;
          }
          usedNames.add(uniqueName);
          
          res.push({
            path: m.path,
            name: uniqueName,
            redirect: firstChild.path,
            meta: {
              permission: m.permissionCode || undefined // 如果没有权限码，设为 undefined
            }
          });
        }
      }
      if (m.children?.length) {
        walk(m.children);
      }
    });
  };
  walk(menus);
  return res;
}

router.beforeEach(async (to, _from, next) => {
  try {
    const userStore = useUserStore();
    const {hasPermission} = usePermission();

    if (to.path !== '/login' && !userStore.token) {
      next('/login');
      return;
    }

    if (userStore.token) {
      // 如果未初始化或菜单数据为空，重新获取
      if (!initialized || !userStore.menus || userStore.menus.length === 0) {
        initialized = true;
        try {
          await userStore.fetchProfile();
        } catch (err) {
          console.error('[Router] 获取用户信息失败:', err);
        }
        try {
          await userStore.fetchMenus();
        } catch (err) {
          console.error('[Router] 获取菜单失败:', err);
        }
      }
      
      // 添加动态路由
      if (userStore.menus && userStore.menus.length > 0) {
        const dynRoutes = buildRoutesFromMenus(userStore.menus);
        dynRoutes.forEach((r) => {
          if (!dynamicAdded.has(r.path as string)) {
            try {
              router.addRoute('Root', r);
              dynamicAdded.add(r.path as string);
            } catch (err) {
              console.error(`[Router] 添加路由失败: ${r.path}`, err);
            }
          }
        });

        // 首次加载时，如果当前路由是 404，但动态路由刚刚注入，需要重新匹配一次
        if (to.name === 'NotFound') {
          const resolved = router.resolve(to.fullPath);
          // 只有在新匹配结果不是 NotFound 时才重定向，避免死循环
          if (resolved.name && resolved.name !== 'NotFound') {
            next({...resolved, replace: true});
            return;
          }
        }
      }
    }

    // 权限检查：只有当 meta.permission 存在且不为空时才检查权限
    const needPerm = to.meta?.permission as string | undefined;
    if (needPerm && needPerm.trim() !== '' && !hasPermission(needPerm)) {
      next('/403');
      return;
    }

    next();
  } catch (err) {
    console.error('[Router] 路由守卫错误:', err);
    // 如果路由守卫出错，仍然允许导航（避免阻塞）
    next();
  }
});

export default router;

