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
  return undefined;
}

function buildRoutesFromMenus(menus: MenuItem[]): RouteRecordRaw[] {
  const res: RouteRecordRaw[] = [];
  const walk = (items: MenuItem[]) => {
    items.forEach((m) => {
      // 处理菜单（type=2）：有实际页面组件
      if (m.type === 2 && m.path) {
        const comp = resolveComponent(m.component, m.path);
        if (comp) {
          res.push({
            path: m.path,
            name: m.path,
            meta: {permission: m.permissionCode, keepAlive: true},
            component: comp
          });
        }
      }
      // 处理目录（type=1）：如果有子菜单，重定向到第一个子菜单
      if (m.type === 1 && m.path && m.children && m.children.length > 0) {
        // 找到第一个有效的子菜单
        const firstChild = m.children.find((child) => child.type === 2 && child.path);
        if (firstChild && firstChild.path) {
          res.push({
            path: m.path,
            name: m.path,
            redirect: firstChild.path,
            meta: {permission: m.permissionCode}
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
      await userStore.fetchProfile().catch(() => {});
      await userStore.fetchMenus().catch(() => {});
    }
    
    // 添加动态路由
    if (userStore.menus && userStore.menus.length > 0) {
      const dynRoutes = buildRoutesFromMenus(userStore.menus);
      dynRoutes.forEach((r) => {
        if (!dynamicAdded.has(r.path as string)) {
          router.addRoute('Root', r);
          dynamicAdded.add(r.path as string);
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

  const needPerm = to.meta?.permission as string | undefined;
  if (needPerm && !hasPermission(needPerm)) {
    next('/403');
    return;
  }

  next();
});

export default router;

