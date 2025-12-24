<template>
  <div class="error-page">
    <h2>403</h2>
    <p>{{ t('page.noAccess') }}</p>
    <el-button type="primary" @click="$router.push('/')">{{ t('page.backHome') }}</el-button>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n';
const {t} = useI18n();
</script>

<style scoped>
.error-page {
  min-height: 60vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}
</style>

