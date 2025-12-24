import axios from 'axios';
import {useUserStore} from '@/stores/user';

const instance = axios.create({
  baseURL: '/api',
  timeout: 15000
});

instance.interceptors.request.use((config) => {
  const userStore = useUserStore();
  if (userStore.token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${userStore.token}`;
  }
  return config;
});

// 根据后端 Envelope 结构统一处理响应：{ code, msg, data }
instance.interceptors.response.use(
  (resp) => {
    const res = resp.data;
    // 标准包裹结构
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code === 0) {
        return (res as any).data;
      }
      const msg = (res as any).msg || '请求失败';
      return Promise.reject(new Error(msg));
    }
    // 非标准结构，直接返回原始 data（兼容特殊接口）
    return res;
  },
  (error) => {
    const data = error?.response?.data;
    const msg =
      (data && (data.msg || data.message)) ||
      error.message ||
      '请求失败';
    return Promise.reject(new Error(msg));
  }
);

export default instance;

