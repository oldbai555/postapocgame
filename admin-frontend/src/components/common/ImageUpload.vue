<template>
  <div class="image-upload">
    <el-upload
      :action="uploadUrl"
      :headers="uploadHeaders"
      :on-success="handleUploadSuccess"
      :on-error="handleUploadError"
      :before-upload="beforeUpload"
      :show-file-list="false"
      :disabled="disabled"
      class="image-uploader"
    >
      <el-image
        v-if="modelValue"
        :src="imageUrl"
        fit="cover"
        class="upload-image"
        :preview-src-list="[imageUrl]"
        :initial-index="0"
        preview-teleported
      >
        <template #error>
          <div class="image-slot">
            <el-icon><Picture /></el-icon>
          </div>
        </template>
      </el-image>
      <el-icon v-else class="uploader-icon"><Plus /></el-icon>
      <template #tip>
        <div class="el-upload__tip">
          {{ tip || '只能上传jpg/png文件，且不超过2MB' }}
        </div>
      </template>
    </el-upload>
    <el-button
      v-if="modelValue && !disabled"
      type="danger"
      size="small"
      class="delete-btn"
      @click="handleDelete"
    >
      删除
    </el-button>
  </div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue';
import {ElMessage} from 'element-plus';
import {Picture, Plus} from '@element-plus/icons-vue';
import {useUserStore} from '@/stores/user';
import {fileUpload} from '@/api/generated/admin';
import type {FileUploadResp} from '@/api/generated/admin';

interface Props {
  modelValue?: string; // 图片URL
  disabled?: boolean; // 是否禁用
  tip?: string; // 提示文字
  maxSize?: number; // 最大文件大小（MB），默认2MB
  accept?: string; // 接受的文件类型，默认图片
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  disabled: false,
  tip: '',
  maxSize: 2,
  accept: 'image/*'
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
  'change': [value: string];
}>();

const userStore = useUserStore();
const baseUrl = computed(() => import.meta.env.VITE_API_BASE_URL || '');

// 上传URL
const uploadUrl = computed(() => `${baseUrl.value}/api/v1/files/upload`);

// 上传请求头
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`
}));

// 图片URL（如果是相对路径，需要拼接baseUrl）
const imageUrl = computed(() => {
  if (!props.modelValue) return '';
  // 如果已经是完整URL，直接返回
  if (props.modelValue.startsWith('http://') || props.modelValue.startsWith('https://')) {
    return props.modelValue;
  }
  // 如果是相对路径，拼接baseUrl
  return `${baseUrl.value}${props.modelValue}`;
});

// 上传前验证
const beforeUpload = (file: File) => {
  // 验证文件类型
  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    ElMessage.error('只能上传图片文件！');
    return false;
  }

  // 验证文件大小
  const isValidSize = file.size / 1024 / 1024 < props.maxSize;
  if (!isValidSize) {
    ElMessage.error(`图片大小不能超过 ${props.maxSize}MB！`);
    return false;
  }

  return true;
};

// 上传成功
const handleUploadSuccess = (response: FileUploadResp) => {
  if (response && response.url) {
    emit('update:modelValue', response.url);
    emit('change', response.url);
    ElMessage.success('上传成功');
  } else {
    ElMessage.error('上传失败：服务器返回数据格式错误');
  }
};

// 上传失败
const handleUploadError = (error: any) => {
  ElMessage.error('上传失败：' + (error.message || '未知错误'));
};

// 删除图片
const handleDelete = () => {
  emit('update:modelValue', '');
  emit('change', '');
};
</script>

<style scoped lang="scss">
.image-upload {
  display: inline-block;

  .image-uploader {
    :deep(.el-upload) {
      border: 1px dashed var(--el-border-color);
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: var(--el-transition-duration-fast);

      &:hover {
        border-color: var(--el-color-primary);
      }
    }

    .upload-image {
      width: 100px;
      height: 100px;
      display: block;
    }

    .uploader-icon {
      font-size: 28px;
      color: #8c939d;
      width: 100px;
      height: 100px;
      line-height: 100px;
      text-align: center;
    }

    .image-slot {
      display: flex;
      justify-content: center;
      align-items: center;
      width: 100%;
      height: 100%;
      background: var(--el-fill-color-light);
      color: var(--el-text-color-placeholder);
      font-size: 30px;
    }
  }

  .delete-btn {
    margin-top: 8px;
    width: 100px;
  }

  :deep(.el-upload__tip) {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
    margin-top: 4px;
  }
}
</style>

