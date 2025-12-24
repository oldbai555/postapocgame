import {computed} from 'vue';
import {useUserStore} from '@/stores/user';

export function usePermission() {
  const userStore = useUserStore();

  const permissions = computed(() => userStore.permissions || []);

  const hasPermission = (code?: string) => {
    if (!code) return true;
    const list = permissions.value;
    if (!list || list.length === 0) return false;
    return list.includes('*') || list.includes(code);
  };

  return {hasPermission, permissions};
}

