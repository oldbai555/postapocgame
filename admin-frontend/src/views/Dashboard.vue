<template>
  <div class="dashboard">
    <el-card class="dashboard__card">
      <h3>{{ t('common.welcome') }}</h3>
      <p>{{ t('common.selectMenu') }}</p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n';
const {t} = useI18n();
</script>

<style scoped>
.dashboard {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.dashboard__card {
  flex: 1;
}
</style>

