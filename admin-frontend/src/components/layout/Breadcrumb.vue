<template>
  <el-breadcrumb separator="/" class="breadcrumb">
    <el-breadcrumb-item
      v-for="(item, index) in items"
      :key="index"
      :to="item.path"
    >
      {{ item.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
export interface BreadcrumbItem {
  title: string;
  path?: string;
}

interface Props {
  items: BreadcrumbItem[];
}

defineProps<Props>();
</script>

<style scoped lang="scss">
.breadcrumb {
  font-size: 14px;
}
</style>

