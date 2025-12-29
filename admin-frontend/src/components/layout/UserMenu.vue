<template>
  <el-dropdown @command="handleCommand" trigger="click">
    <div class="user-menu__trigger">
      <el-avatar :size="32" class="user-menu__avatar" :src="props.user?.avatar">
        {{ userAvatarText }}
      </el-avatar>
      <span class="user-menu__name">{{ userName }}</span>
      <el-icon class="user-menu__icon"><ArrowDown /></el-icon>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item command="profile">
          <el-icon><User /></el-icon>
          <span style="margin-left: 8px">{{ t('common.profile') }}</span>
        </el-dropdown-item>
        <el-dropdown-item command="password">
          <el-icon><Lock /></el-icon>
          <span style="margin-left: 8px">{{ t('common.changePassword') }}</span>
        </el-dropdown-item>
        <el-dropdown-item divided command="logout">
          <el-icon><SwitchButton /></el-icon>
          <span style="margin-left: 8px">{{ t('common.logout') }}</span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import {computed} from 'vue';
import {useRouter} from 'vue-router';
import {ArrowDown, User, Lock, SwitchButton} from '@element-plus/icons-vue';
import {useI18n} from 'vue-i18n';
import type {ProfileResp} from '@/api/generated/admin';

interface Props {
  user: ProfileResp | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  logout: [];
}>();

const {t} = useI18n();
const router = useRouter();

const userName = computed(() => {
  return props.user?.username || props.user?.nickname || 'User';
});

const userAvatarText = computed(() => {
  const name = userName.value;
  return name.charAt(0).toUpperCase();
});

const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      // 跳转到个人信息页面
      router.push('/system/profile');
      break;
    case 'password':
      // 跳转到个人信息页面（该页面包含修改密码功能）
      router.push('/system/profile');
      break;
    case 'logout':
      emit('logout');
      break;
  }
};
</script>

<style scoped lang="scss">
.user-menu {
  &__trigger {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
    transition: background-color 0.3s;

    &:hover {
      background-color: var(--el-fill-color-light);
    }
  }

  &__avatar {
    flex-shrink: 0;
  }

  &__name {
    font-size: 14px;
    color: var(--color-text-primary);
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__icon {
    font-size: 12px;
    color: var(--color-text-secondary);
  }
}
</style>

