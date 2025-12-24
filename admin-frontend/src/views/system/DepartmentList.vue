<template>
  <div class="page">
    <el-card>
      <div class="toolbar">
        <el-button type="success" @click="openCreate()" v-permission="'department:create'">
          {{ t('common.create') }}
        </el-button>
      </div>
      <el-tree
        :data="treeData"
        node-key="id"
        :props="{label: 'name', children: 'children'}"
        default-expand-all
      >
        <template #default="{data}">
          <span>{{ data.name }}</span>
          <span class="ops">
            <el-button link type="primary" @click.stop="openCreate(data)" v-permission="'department:create'">
              {{ t('common.create') }}
            </el-button>
            <el-button link type="primary" @click.stop="openEdit(data)" v-permission="'department:update'">
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="danger" @click.stop="handleDelete(data)" v-permission="'department:delete'">
              {{ t('common.delete') }}
            </el-button>
          </span>
        </template>
      </el-tree>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? t('common.edit') : t('common.create')" width="420px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item :label="t('common.department')">
          <el-input v-model="parentName" disabled />
        </el-form-item>
        <el-form-item :label="t('common.name')" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item :label="t('common.order')">
          <el-input-number v-model="form.orderNum" :min="0" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ t('common.save') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {ref, reactive, onMounted, computed} from 'vue';
import {ElMessage, ElMessageBox, ElForm} from 'element-plus';
import {departmentTree, departmentCreate, departmentUpdate, departmentDelete} from '@/api/generated/admin';
import type {DepartmentItem, DepartmentCreateReq, DepartmentUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';

const {t} = useI18n();

const treeData = ref<DepartmentItem[]>([]);
const loading = ref(false);

const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<InstanceType<typeof ElForm>>();
const form = reactive({
  id: 0,
  parentId: 0,
  name: '',
  orderNum: 0,
  status: 1
});
const parentName = computed(() => {
  if (form.parentId === 0) return '根节点';
  const find = (list: DepartmentItem[], id: number): DepartmentItem | undefined => {
    for (const item of list) {
      if (item.id === id) return item;
      if (item.children) {
        const got = find(item.children, id);
        if (got) return got;
      }
    }
    return undefined;
  };
  return find(treeData.value, form.parentId)?.name || '根节点';
});

const rules = {
  name: [{required: true, message: t('common.name'), trigger: 'blur'}]
};

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await departmentTree();
    treeData.value = resp.list || [];
  } catch (err: any) {
    ElMessage.error(err.message || t('common.search'));
  } finally {
    loading.value = false;
  }
};

const openCreate = (parent?: DepartmentItem) => {
  isEdit.value = false;
  Object.assign(form, {id: 0, parentId: parent?.id || 0, name: '', orderNum: 0, status: 1});
  dialogVisible.value = true;
};

const openEdit = (data: DepartmentItem) => {
  isEdit.value = true;
  Object.assign(form, {
    id: data.id,
    parentId: data.parentId,
    name: data.name,
    orderNum: data.orderNum,
    status: data.status
  });
  dialogVisible.value = true;
};

const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (!valid) return;
    submitLoading.value = true;
    try {
      if (isEdit.value) {
        await departmentUpdate(form as DepartmentUpdateReq);
        ElMessage.success('更新成功');
      } else {
        await departmentCreate(form as DepartmentCreateReq);
        ElMessage.success('新增成功');
      }
      // 先刷新数据，等待完成后再关闭对话框
      await loadData();
      dialogVisible.value = false;
    } catch (err: any) {
      ElMessage.error(err.message || '提交失败');
    } finally {
      submitLoading.value = false;
    }
  });
};

const submitLoading = ref(false);

const handleDelete = (data: DepartmentItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await departmentDelete({id: data.id});
      ElMessage.success(t('common.delete'));
      await loadData();
    })
    .catch(() => {});
};

onMounted(loadData);
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.toolbar {
  margin-bottom: 8px;
}
.ops {
  margin-left: 12px;
  display: inline-flex;
  gap: 6px;
}
</style>

