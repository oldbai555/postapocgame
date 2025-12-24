<template>
  <div class="page-header">
    <div class="page-header__content">
      <!-- 面包屑导航 -->
      <Breadcrumb v-if="breadcrumb && breadcrumb.length > 0" :items="breadcrumb" class="page-header__breadcrumb" />
      
      <!-- 标题和操作按钮 -->
      <div class="page-header__title-bar">
        <h2 v-if="title" class="page-header__title">{{ title }}</h2>
        <div v-if="$slots.actions" class="page-header__actions">
          <slot name="actions" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Breadcrumb, {type BreadcrumbItem} from './Breadcrumb.vue';

interface Props {
  title?: string;
  breadcrumb?: BreadcrumbItem[];
}

withDefaults(defineProps<Props>(), {
  title: '',
  breadcrumb: () => []
});
</script>

<style scoped lang="scss">
@use '@/styles/variables.scss' as *;

.page-header {
  padding: $spacing-md $spacing-lg;
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);

  &__content {
    display: flex;
    flex-direction: column;
    gap: $spacing-sm;
  }

  &__breadcrumb {
    margin-bottom: $spacing-xs;
  }

  &__title-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: var(--color-text-primary);
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
  }
}
</style>

