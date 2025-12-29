<template>
  <div class="profile-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>个人信息</span>
        </div>
      </template>

      <el-form :model="profileForm" label-width="120px" style="max-width: 600px">
        <el-form-item label="用户名">
          <el-input v-model="profileForm.username" disabled />
        </el-form-item>

        <el-form-item label="头像">
          <ImageUpload v-model="profileForm.avatar" />
        </el-form-item>

        <el-form-item label="个性签名">
          <el-input
            v-model="profileForm.signature"
            type="textarea"
            :rows="3"
            placeholder="请输入个性签名"
            maxlength="255"
            show-word-limit
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSaveProfile">
            保存
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="mt-4">
      <template #header>
        <div class="card-header">
          <span>修改密码</span>
        </div>
      </template>

      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="120px" style="max-width: 600px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input
            v-model="passwordForm.oldPassword"
            type="password"
            placeholder="请输入原密码"
            show-password
          />
        </el-form-item>

        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="changingPassword" @click="handleChangePassword">
            修改密码
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import {ref, reactive, onMounted} from 'vue';
import {ElMessage, ElMessageBox, type FormInstance, type FormRules} from 'element-plus';
import {useUserStore} from '@/stores/user';
import {profile, profileUpdate, passwordChange} from '@/api/generated/admin';
import type {ProfileResp, ProfileUpdateReq, PasswordChangeReq} from '@/api/generated/admin';
import ImageUpload from '@/components/common/ImageUpload.vue';

const userStore = useUserStore();

// 个人信息表单
const profileForm = reactive<{
  username: string;
  avatar: string;
  signature: string;
}>({
  username: '',
  avatar: '',
  signature: ''
});

// 密码修改表单
const passwordForm = reactive<{
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}>({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const passwordFormRef = ref<FormInstance>();
const saving = ref(false);
const changingPassword = ref(false);

// 密码验证规则
const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'));
  } else {
    callback();
  }
};

const passwordRules: FormRules = {
  oldPassword: [
    {required: true, message: '请输入原密码', trigger: 'blur'}
  ],
  newPassword: [
    {required: true, message: '请输入新密码', trigger: 'blur'},
    {min: 6, message: '密码长度不能少于6位', trigger: 'blur'}
  ],
  confirmPassword: [
    {required: true, message: '请再次输入新密码', trigger: 'blur'},
    {validator: validateConfirmPassword, trigger: 'blur'}
  ]
};

// 加载个人信息
const loadProfile = async () => {
  try {
    const data = await profile();
    profileForm.username = data.username || '';
    profileForm.avatar = data.avatar || '';
    profileForm.signature = data.signature || '';
  } catch (err: any) {
    ElMessage.error(err.message || '加载个人信息失败');
  }
};

// 保存个人信息
const handleSaveProfile = async () => {
  saving.value = true;
  try {
    const req: ProfileUpdateReq = {
      avatar: profileForm.avatar,
      signature: profileForm.signature
    };
    await profileUpdate(req);
    ElMessage.success('保存成功');
    // 刷新用户信息
    await userStore.fetchProfile(true);
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败');
  } finally {
    saving.value = false;
  }
};

// 修改密码
const handleChangePassword = async () => {
  if (!passwordFormRef.value) return;

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return;

    changingPassword.value = true;
    try {
      const req: PasswordChangeReq = {
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword
      };
      await passwordChange(req);
      ElMessage.success('密码修改成功，请重新登录');
      // 清空表单
      passwordForm.oldPassword = '';
      passwordForm.newPassword = '';
      passwordForm.confirmPassword = '';
      passwordFormRef.value.resetFields();
      // 延迟退出登录，让用户看到成功提示
      setTimeout(() => {
        userStore.logout();
      }, 1500);
    } catch (err: any) {
      ElMessage.error(err.message || '密码修改失败');
    } finally {
      changingPassword.value = false;
    }
  });
};

onMounted(() => {
  loadProfile();
});
</script>

<style scoped lang="scss">
.profile-page {
  padding: 20px;

  .mt-4 {
    margin-top: 20px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

