import {createApp} from 'vue';
import {createPinia} from 'pinia';
import App from './App.vue';
import router from './router';
import i18n from './i18n';
import permissionDirective from './directives/permission';
import {useAppStore} from './stores/app';
import ElementPlus from 'element-plus';

import 'element-plus/dist/index.css';
import './styles/theme.scss';
import './styles/layout.scss';

// 全局错误处理：忽略浏览器扩展相关的错误
window.addEventListener('error', (event) => {
  // 忽略浏览器扩展相关的错误
  if (
    event.message?.includes('message channel closed') ||
    event.message?.includes('asynchronous response') ||
    event.message?.includes('Extension context invalidated')
  ) {
    event.preventDefault();
    return false;
  }
});

// 处理未捕获的 Promise 错误
window.addEventListener('unhandledrejection', (event) => {
  // 忽略浏览器扩展相关的错误
  const errorMessage = event.reason?.message || event.reason?.toString() || '';
  if (
    errorMessage.includes('message channel closed') ||
    errorMessage.includes('asynchronous response') ||
    errorMessage.includes('Extension context invalidated')
  ) {
    event.preventDefault();
    return false;
  }
});

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
app.use(router);
app.use(i18n);
app.use(ElementPlus);
app.directive('permission', permissionDirective);

const appStore = useAppStore(pinia);
appStore.init();

if (appStore.lang) {
  i18n.global.locale.value = appStore.lang;
}

app.mount('#app');

