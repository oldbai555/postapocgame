<template>
  <div class="app-layout">
    <!-- 顶部导航栏 -->
    <AppHeader
      :collapsed="appStore.sidebarCollapsed"
      :show-collapse-button="false"
      :user="userStore.profile"
      @toggle-collapse="handleToggleCollapse"
      @logout="handleLogout"
    >
      <!-- 面包屑导航 -->
      <Breadcrumb v-if="breadcrumbItems.length > 0" :items="breadcrumbItems" />
    </AppHeader>

    <!-- 主体区域 -->
    <main class="app-layout__main">
      <!-- 侧边栏 -->
      <AppSidebar
        :collapsed="appStore.sidebarCollapsed"
        :menus="displayMenus"
      />

      <!-- 内容区域 -->
      <section class="app-layout__content">
        <!-- 页面标题栏 -->
        <PageHeader
          v-if="pageTitle || breadcrumbItems.length > 0"
          :title="pageTitle"
          :breadcrumb="breadcrumbItems"
        />

        <!-- 页面内容 -->
        <div class="app-layout__page-content">
          <RouterView />
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import {computed, watch, onMounted, onUnmounted} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {ElMessage} from 'element-plus';
import {useI18n} from 'vue-i18n';
import {useUserStore} from '@/stores/user';
import {usePermission} from '@/hooks/usePermission';
import {useAppStore} from '@/stores/app';
import {useWebSocketStore} from '@/stores/websocket';
import AppHeader from '@/components/layout/AppHeader.vue';
import AppSidebar from '@/components/layout/AppSidebar.vue';
import PageHeader from '@/components/layout/PageHeader.vue';
import Breadcrumb from '@/components/layout/Breadcrumb.vue';
import {generateBreadcrumb} from '@/utils/breadcrumb';
import type {MenuItem} from '@/api/generated/admin';

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const {hasPermission} = usePermission();
const appStore = useAppStore();
const wsStore = useWebSocketStore();
const {t} = useI18n();

// 初始化应用
onMounted(() => {
  appStore.init();
  if (userStore.token && (!userStore.menus || userStore.menus.length === 0)) {
    userStore.fetchMenus().catch(() => {});
  }

  // 登录后自动连接 WebSocket
  if (userStore.token) {
    wsStore.connect();
  }
});

// 监听登录状态变化
watch(
  () => userStore.token,
  (newToken, oldToken) => {
    if (newToken && !oldToken) {
      // 用户登录，连接 WebSocket
      wsStore.connect();
    } else if (!newToken && oldToken) {
      // 用户退出，断开 WebSocket
      wsStore.disconnect();
    }
  }
);

// 监听路由变化，页面切换时保持连接
watch(
  () => route.path,
  () => {
    // 如果连接断开，尝试重连
    if (userStore.token && !wsStore.connected && !wsStore.connecting) {
      wsStore.connect();
    }
  }
);

// 组件卸载时断开连接
onUnmounted(() => {
  // 注意：这里不断开连接，因为可能在其他页面还需要使用
  // 只在用户退出登录时断开
});

// 过滤菜单
const filterMenu = (items: MenuItem[]): MenuItem[] => {
  return items
    .filter((m) => m.status === 1 && m.visible !== 0)
    .filter((m) => {
      // 如果菜单没有权限码（undefined、null 或空字符串），则允许访问
      const code = m.permissionCode;
      if (!code || code.trim() === '') {
        return true; // 没有权限码的菜单，只要登录就可以访问
      }
      // 有权限码的菜单，需要检查权限
      return hasPermission(code);
    })
    .map((m) => ({
      ...m,
      children: m.children ? filterMenu(m.children) : []
    }));
};

const displayMenus = computed(() => filterMenu(userStore.menus || []));

// 生成面包屑
const breadcrumbItems = computed(() => {
  if (!userStore.menus || userStore.menus.length === 0) {
    return [];
  }
  return generateBreadcrumb(route, userStore.menus);
});

// 页面标题（从路由 meta 或菜单数据获取）
const pageTitle = computed(() => {
  // 优先从路由 meta 获取
  if (route.meta?.title) {
    return route.meta.title as string;
  }
  // 从面包屑最后一项获取
  if (breadcrumbItems.value.length > 0) {
    return breadcrumbItems.value[breadcrumbItems.value.length - 1].title;
  }
  return '';
});

// 切换侧边栏折叠
const handleToggleCollapse = () => {
  appStore.toggleSidebar();
};

// 退出登录
const handleLogout = async () => {
  await userStore.logout();
  ElMessage.success(t('common.logout'));
  router.push('/login');
};
</script>

<style scoped lang="scss">
@use '@/styles/variables.scss' as *;

.app-layout {
  &__page-content {
    flex: 1;
    padding: $spacing-lg;
    overflow-y: auto;
  }
}
</style>
