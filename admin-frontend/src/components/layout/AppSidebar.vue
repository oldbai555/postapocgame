<template>
  <aside :class="['app-sidebar', {'app-sidebar--collapsed': collapsed}]">
    <el-menu
      :default-active="activePath"
      :collapse="collapsed"
      :unique-opened="true"
      router
      @select="handleMenuSelect"
      class="app-sidebar__menu"
    >
      <template v-for="item in displayMenus" :key="item.id">
        <!-- 有子菜单 -->
        <el-sub-menu 
          v-if="item.children?.length" 
          :index="getSubMenuIndex(item)"
        >
          <template #title>
            <el-icon v-if="getMenuIcon(item.icon)">
              <component :is="getMenuIcon(item.icon)" />
            </el-icon>
            <span>{{ item.name }}</span>
          </template>
          <el-menu-item
            v-for="child in item.children?.filter(isMenu) || []"
            :key="child.id"
            :index="child.path"
          >
            <el-icon v-if="getMenuIcon(child.icon)">
              <component :is="getMenuIcon(child.icon)" />
            </el-icon>
            <span>{{ child.name }}</span>
          </el-menu-item>
        </el-sub-menu>
        <!-- 无子菜单 -->
        <el-menu-item v-else-if="isMenu(item)" :index="item.path">
          <el-icon v-if="getMenuIcon(item.icon)">
            <component :is="getMenuIcon(item.icon)" />
          </el-icon>
          <span>{{ item.name }}</span>
        </el-menu-item>
      </template>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import {computed} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import {useUserStore} from '@/stores/user';
import type {MenuItem} from '@/api/generated/admin';

interface Props {
  collapsed?: boolean;
  menus: MenuItem[];
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
  menus: () => []
});

const route = useRoute();
const router = useRouter();

const activePath = computed(() => route.path);

// 处理菜单选择（作为 router 属性的备用方案）
const handleMenuSelect = async (path: string) => {
  if (!path) {
    return;
  }
  
  // 如果路径相同，不处理
  if (path === route.path) {
    return;
  }
  
  // 确保路由已注册（等待菜单加载完成）
  const userStore = useUserStore();
  if (userStore.menus && userStore.menus.length > 0) {
    // 等待一下，确保路由已经注册
    await new Promise(resolve => setTimeout(resolve, 50));
  }
  
  // 检查路由是否存在
  const resolved = router.resolve(path);
  if (resolved.name === 'NotFound') {
    console.error(`[Menu] 路由不存在: ${path}`);
    return;
  }
  
  // 执行路由跳转
  try {
    await router.push(path).catch((err: any) => {
      // 如果是导航重复错误，这是正常的，不需要处理
      if (err.name === 'NavigationDuplicated') {
        return;
      }
      // 其他错误重新抛出
      throw err;
    });
  } catch (err: any) {
    // 如果路由跳转失败，输出错误信息（排除导航重复错误）
    if (err.name !== 'NavigationDuplicated') {
      console.error(`[Menu] 路由跳转失败: ${path}`, err);
      // 尝试使用 replace 方式跳转
      router.replace(path).catch(() => {
        // 静默处理替换失败
      });
    }
  }
};

// 过滤菜单（这里简化处理，实际应该从父组件传入已过滤的菜单）
const displayMenus = computed(() => props.menus);

const isMenu = (item: MenuItem | undefined | null) => {
  if (!item) return false;
  return item.type === 2 || item.type === 1;
};

// 获取菜单图标
const getMenuIcon = (iconName?: string) => {
  if (!iconName) return null;
  // Element Plus Icons 映射
  const iconMap: Record<string, any> = ElementPlusIconsVue;
  return iconMap[iconName] || null;
};

// 获取子菜单的 index：目录类型使用第一个子菜单的路径，其他使用自身路径
const getSubMenuIndex = (item: MenuItem): string => {
  if (!item) return '';
  // 如果是目录类型（type=1）且有子菜单，返回第一个子菜单的路径
  if (item.type === 1 && item.children && item.children.length > 0) {
    const firstChild = item.children.find((child) => child && child.type === 2 && child.path);
    if (firstChild && firstChild.path) {
      return firstChild.path;
    }
  }
  // 其他情况使用自身路径或 ID
  return item.path || String(item.id);
};
</script>

<style scoped lang="scss">
@use '@/styles/variables.scss' as *;

.app-sidebar {
  width: $sidebar-width;
  background: var(--color-bg-primary);
  border-right: 1px solid var(--color-border);
  transition: width $transition-base;
  overflow: hidden;

  &--collapsed {
    width: $sidebar-collapsed-width;
  }

  &__menu {
    border-right: none;
    height: 100%;
    overflow-y: auto;

    &:not(.el-menu--collapse) {
      width: $sidebar-width;
    }
  }
}
</style>

