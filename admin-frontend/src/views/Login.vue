<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-card__header">
        <div class="logo">Admin System</div>
        <div class="subtitle">{{ t('common.welcome') }}</div>
      </div>
      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="login-form"
      >
        <el-form-item :label="t('auth.username')" prop="username">
          <el-input
            v-model="form.username"
            size="large"
            placeholder="admin"
            autocomplete="username"
            clearable
          />
        </el-form-item>
        <el-form-item :label="t('auth.password')" prop="password">
          <el-input
            v-model="form.password"
            size="large"
            type="password"
            placeholder="••••••"
            show-password
            autocomplete="current-password"
          />
        </el-form-item>
        <el-form-item class="login-actions">
          <el-button type="primary" size="large" :loading="loading" @click="handleSubmit" class="full-btn">
            {{ t('common.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue';
import {ElForm, ElMessage} from 'element-plus';
import {useRouter} from 'vue-router';
import {useUserStore} from '@/stores/user';
import {useI18n} from 'vue-i18n';

const router = useRouter();
const userStore = useUserStore();
const {t} = useI18n();

const form = reactive({
  username: '',
  password: ''
});

const rules = {
  username: [{required: true, message: t('auth.username'), trigger: 'blur'}],
  password: [{required: true, message: t('auth.password'), trigger: 'blur'}]
};

const formRef = ref<InstanceType<typeof ElForm>>();
const loading = ref(false);

const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (!valid) return;
    loading.value = true;
    try {
      await userStore.login(form);
      ElMessage.success(t('auth.loginSuccess'));
      router.push('/');
    } catch (err: any) {
      ElMessage.error(err.message || t('auth.loginFail'));
    } finally {
      loading.value = false;
    }
  });
};
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(120% 120% at 50% 20%, #eef2ff, #f7f9fc);
  padding: 24px;
}
.login-card {
  width: 420px;
  padding: 32px 32px 24px;
  background: var(--color-card, #fff);
  border-radius: 16px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.04);
}
.login-card__header {
  text-align: center;
  margin-bottom: 24px;
}
.logo {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-primary, #409eff);
}
.subtitle {
  margin-top: 6px;
  color: #606266;
  font-size: 14px;
}
.login-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: #303133;
}
.login-actions {
  margin-top: 8px;
}
.full-btn {
  width: 100%;
}
</style>

