import {RouteLocationNormalized} from 'vue-router';
import type {MenuItem} from '@/api/generated/admin';
import type {BreadcrumbItem} from '@/components/layout/Breadcrumb.vue';

/**
 * 从菜单数据生成面包屑
 */
export function generateBreadcrumb(
  route: RouteLocationNormalized,
  menus: MenuItem[]
): BreadcrumbItem[] {
  const breadcrumbs: BreadcrumbItem[] = [];
  
  // 首页
  breadcrumbs.push({
    title: '首页',
    path: '/dashboard'
  });

  // 查找当前路由对应的菜单项
  const findMenuByPath = (path: string, items: MenuItem[]): MenuItem | null => {
    for (const item of items) {
      if (item.path === path) {
        return item;
      }
      if (item.children?.length) {
        const found = findMenuByPath(path, item.children);
        if (found) return found;
      }
    }
    return null;
  };

  // 查找菜单路径链
  const findMenuPath = (targetPath: string, items: MenuItem[], path: MenuItem[] = []): MenuItem[] | null => {
    for (const item of items) {
      const currentPath = [...path, item];
      if (item.path === targetPath) {
        return currentPath;
      }
      if (item.children?.length) {
        const found = findMenuPath(targetPath, item.children, currentPath);
        if (found) return found;
      }
    }
    return null;
  };

  const menuPath = findMenuPath(route.path, menus);
  if (menuPath) {
    menuPath.forEach((menu) => {
      if (menu.path && menu.type === 2) {
        breadcrumbs.push({
          title: menu.name,
          path: menu.path
        });
      }
    });
  } else {
    // 如果没有找到对应的菜单，使用路由路径生成
    const pathSegments = route.path.split('/').filter(Boolean);
    pathSegments.forEach((segment, index) => {
      const path = '/' + pathSegments.slice(0, index + 1).join('/');
      breadcrumbs.push({
        title: segment.charAt(0).toUpperCase() + segment.slice(1),
        path: index === pathSegments.length - 1 ? undefined : path
      });
    });
  }

  return breadcrumbs;
}

