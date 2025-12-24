<template>
  <div class="page">
    <el-card>
      <div class="toolbar">
        <el-button type="success" @click="openCreate()" v-permission="'menu:create'">
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
          <span class="menu-item">
            <el-icon v-if="getMenuIcon(data.icon)" class="menu-icon">
              <component :is="getMenuIcon(data.icon)" />
            </el-icon>
            <span class="menu-name">{{ data.name }}</span>
            <el-tag v-if="data.type === 1" size="small" type="info">目录</el-tag>
            <el-tag v-else-if="data.type === 2" size="small" type="success">菜单</el-tag>
            <el-tag v-else-if="data.type === 3" size="small" type="warning">按钮</el-tag>
          </span>
          <span class="ops">
            <el-button link type="primary" @click.stop="openCreate(data)" v-permission="'menu:create'">
              {{ t('common.create') }}
            </el-button>
            <el-button link type="primary" @click.stop="openEdit(data)" v-permission="'menu:update'">
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="danger" @click.stop="handleDelete(data)" v-permission="'menu:delete'">
              {{ t('common.delete') }}
            </el-button>
          </span>
        </template>
      </el-tree>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? t('common.edit') : t('common.create')" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item :label="t('common.parent')">
          <el-input v-model="parentName" disabled />
        </el-form-item>
        <el-form-item :label="t('common.name')" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item :label="t('common.type')" prop="type">
          <el-select v-model="form.type" style="width: 100%">
            <el-option :label="t('menu.type.directory')" :value="1" />
            <el-option :label="t('menu.type.menu')" :value="2" />
            <el-option :label="t('menu.type.button')" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('menu.path')" prop="path" v-if="form.type !== 3">
          <el-input v-model="form.path" :placeholder="t('menu.pathPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('menu.component')" prop="component" v-if="form.type === 2">
          <el-input v-model="form.component" :placeholder="t('menu.componentPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('menu.icon')" prop="icon">
          <el-input v-model="form.icon" :placeholder="t('menu.iconPlaceholder')" />
        </el-form-item>
        <!-- 权限编码字段已移除，改用权限-菜单关联表 -->
        <el-form-item :label="t('common.order')">
          <el-input-number v-model="form.orderNum" :min="0" />
        </el-form-item>
        <el-form-item :label="t('menu.visible')">
          <el-switch v-model="form.visible" :active-value="1" :inactive-value="0" />
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
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import {menuTree, menuCreate, menuUpdate, menuDelete} from '@/api/generated/admin';
import type {MenuItem, MenuCreateReq, MenuUpdateReq} from '@/api/generated/admin';
import {useI18n} from 'vue-i18n';

const {t} = useI18n();

const treeData = ref<MenuItem[]>([]);
const loading = ref(false);

const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<InstanceType<typeof ElForm>>();
const form = reactive({
  id: 0,
  parentId: 0,
  name: '',
  path: '',
  component: '',
  icon: '',
  type: 1, // 1 目录 2 菜单 3 按钮
  orderNum: 0,
  visible: 1,
  status: 1
});

const parentName = computed(() => {
  if (form.parentId === 0) return t('menu.root');
  const find = (list: MenuItem[], id: number): MenuItem | undefined => {
    for (const item of list) {
      if (item.id === id) return item;
      if (item.children) {
        const got = find(item.children, id);
        if (got) return got;
      }
    }
    return undefined;
  };
  return find(treeData.value, form.parentId)?.name || t('menu.root');
});

const rules = {
  name: [{required: true, message: t('common.nameRequired'), trigger: 'blur'}],
  type: [{required: true, message: t('common.typeRequired'), trigger: 'change'}]
};

const getMenuIcon = (iconName?: string) => {
  if (!iconName) return null;
  const iconMap: Record<string, any> = ElementPlusIconsVue;
  // 处理 icon 名称，可能是 "ele-DataBoard" 格式，需要转换为 "DataBoard"
  const iconKey = iconName.startsWith('ele-') ? iconName.substring(4) : iconName;
  return iconMap[iconKey] || null;
};

const loadData = async () => {
  loading.value = true;
  try {
    const resp = await menuTree();
    treeData.value = resp.list || [];
  } catch (err: any) {
    ElMessage.error(err.message || t('common.loadFailed'));
  } finally {
    loading.value = false;
  }
};

const openCreate = (parent?: MenuItem) => {
  isEdit.value = false;
  Object.assign(form, {
    id: 0,
    parentId: parent?.id || 0,
    name: '',
    path: '',
    component: '',
    icon: '',
    type: 1,
    orderNum: 0,
    visible: 1,
    status: 1
  });
  dialogVisible.value = true;
};

const openEdit = (data: MenuItem) => {
  isEdit.value = true;
  Object.assign(form, {
    id: data.id,
    parentId: data.parentId,
    name: data.name,
    path: data.path || '',
    component: data.component || '',
    icon: data.icon || '',
    type: data.type,
    orderNum: data.orderNum,
    visible: data.visible,
    status: data.status
  });
  dialogVisible.value = true;
};

const submitLoading = ref(false);

const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (!valid) return;
    submitLoading.value = true;
    try {
      if (isEdit.value) {
        await menuUpdate(form as MenuUpdateReq);
        ElMessage.success(t('common.updateSuccess'));
      } else {
        await menuCreate(form as MenuCreateReq);
        ElMessage.success(t('common.createSuccess'));
      }
      dialogVisible.value = false;
      loadData();
    } catch (err: any) {
      ElMessage.error(err.message || t('common.submitFailed'));
    } finally {
      submitLoading.value = false;
    }
  });
};

const handleDelete = (data: MenuItem) => {
  ElMessageBox.confirm(t('common.confirmDelete'), t('common.confirm'), {type: 'warning'})
    .then(async () => {
      await menuDelete({id: data.id});
      ElMessage.success(t('common.deleteSuccess'));
      loadData();
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
.menu-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.menu-icon {
  font-size: 16px;
}
.menu-name {
  margin-right: 8px;
}
.ops {
  margin-left: 12px;
  display: inline-flex;
  gap: 6px;
}
</style>

