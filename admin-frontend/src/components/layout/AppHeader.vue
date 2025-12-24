<template>
  <header class="app-header">
    <!-- 左侧：Logo + 折叠按钮 -->
    <div class="app-header__left">
      <div class="app-header__logo" @click="handleLogoClick">
        <span class="app-header__logo-text">Admin System</span>
      </div>
      <el-button
        v-if="showCollapseButton"
        :icon="Fold"
        text
        class="app-header__collapse-btn"
        @click="handleToggleCollapse"
      />
    </div>

    <!-- 中间：面包屑导航（可选） -->
    <div class="app-header__center">
      <slot />
    </div>

    <!-- 右侧：操作按钮 + 用户菜单 -->
    <div class="app-header__right">
      <!-- 全屏切换 -->
      <el-tooltip :content="fullscreen ? t('common.exitFullscreen') : t('common.fullscreen')" placement="bottom">
        <el-button
          :icon="fullscreen ? Aim : FullScreen"
          text
          circle
          class="app-header__action-btn"
          @click="handleFullscreen"
        />
      </el-tooltip>

      <!-- 刷新页面 -->
      <el-tooltip :content="t('common.refresh')" placement="bottom">
        <el-button
          :icon="Refresh"
          text
          circle
          class="app-header__action-btn"
          @click="handleRefresh"
        />
      </el-tooltip>

      <!-- 主题切换 -->
      <el-tooltip :content="theme === 'dark' ? t('common.light') : t('common.dark')" placement="bottom">
        <el-button
          :icon="theme === 'dark' ? Sunny : Moon"
          text
          circle
          class="app-header__action-btn"
          @click="handleToggleTheme"
        />
      </el-tooltip>

      <!-- 语言切换 -->
      <el-dropdown @command="handleLangChange" trigger="click">
        <el-button text circle class="app-header__action-btn">
          <el-icon><Setting /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh" :disabled="lang === 'zh'">中文</el-dropdown-item>
            <el-dropdown-item command="en" :disabled="lang === 'en'">English</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- 用户菜单 -->
      <UserMenu :user="user" @logout="handleLogout" />
    </div>
  </header>
</template>

<script setup lang="ts">
import {computed} from 'vue';
import {useRouter} from 'vue-router';
import {
  Fold,
  FullScreen,
  Aim,
  Refresh,
  Sunny,
  Moon,
  Setting
} from '@element-plus/icons-vue';
import {ElMessage} from 'element-plus';
import {useI18n} from 'vue-i18n';
import {useAppStore} from '@/stores/app';
import {useUserStore} from '@/stores/user';
import UserMenu from './UserMenu.vue';
import type {ProfileResp} from '@/api/generated/admin';

interface Props {
  collapsed?: boolean;
  showCollapseButton?: boolean;
  user?: ProfileResp | null;
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
  showCollapseButton: false,
  user: null
});

const emit = defineEmits<{
  'toggle-collapse': [];
  'logout': [];
}>();

const {t, locale} = useI18n();
const router = useRouter();
const appStore = useAppStore();
const userStore = useUserStore();

const theme = computed(() => appStore.theme);
const lang = computed(() => appStore.lang);
const fullscreen = computed(() => appStore.fullscreen);

const handleLogoClick = () => {
  router.push('/');
};

const handleToggleCollapse = () => {
  emit('toggle-collapse');
};

const handleFullscreen = () => {
  appStore.toggleFullscreen();
};

const handleRefresh = () => {
  window.location.reload();
};

const handleToggleTheme = () => {
  const next = theme.value === 'dark' ? 'light' : 'dark';
  appStore.setTheme(next);
};

const handleLangChange = (val: string) => {
  appStore.setLang(val as 'zh' | 'en');
  locale.value = val;
};

const handleLogout = () => {
  emit('logout');
};
</script>

<style scoped lang="scss">
@use '@/styles/variables.scss' as *;

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 $spacing-lg;
  background: var(--color-bg-primary);
  border-bottom: 1px solid var(--color-border);
  box-shadow: $shadow-sm;

  &__left {
    display: flex;
    align-items: center;
    gap: $spacing-md;
  }

  &__logo {
    display: flex;
    align-items: center;
    cursor: pointer;
    user-select: none;

    &-text {
      font-size: 18px;
      font-weight: 600;
      color: var(--color-primary);
    }
  }

  &__collapse-btn {
    font-size: 18px;
  }

  &__center {
    flex: 1;
    display: flex;
    justify-content: center;
    padding: 0 $spacing-lg;
  }

  &__right {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
  }

  &__action-btn {
    font-size: 18px;
  }
}
</style>

