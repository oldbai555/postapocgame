import {defineStore} from 'pinia';

type Theme = 'light' | 'dark';
type Lang = 'zh' | 'en';

const themeKey = 'admin_theme';
const langKey = 'admin_lang';
const sidebarCollapsedKey = 'admin_sidebar_collapsed';

export const useAppStore = defineStore('app', {
  state: () => ({
    theme: (localStorage.getItem(themeKey) as Theme) || 'light',
    lang: (localStorage.getItem(langKey) as Lang) || 'zh',
    sidebarCollapsed: localStorage.getItem(sidebarCollapsedKey) === 'true',
    fullscreen: false
  }),
  actions: {
    setTheme(theme: Theme) {
      this.theme = theme;
      localStorage.setItem(themeKey, theme);
      document.documentElement.setAttribute('data-theme', theme);
    },
    setLang(lang: Lang) {
      this.lang = lang;
      localStorage.setItem(langKey, lang);
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
      localStorage.setItem(sidebarCollapsedKey, String(this.sidebarCollapsed));
    },
    setSidebarCollapsed(collapsed: boolean) {
      this.sidebarCollapsed = collapsed;
      localStorage.setItem(sidebarCollapsedKey, String(collapsed));
    },
    toggleFullscreen() {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().then(() => {
          this.fullscreen = true;
        });
      } else {
        document.exitFullscreen().then(() => {
          this.fullscreen = false;
        });
      }
    },
    init() {
      document.documentElement.setAttribute('data-theme', this.theme);
      // 监听全屏状态变化
      document.addEventListener('fullscreenchange', () => {
        this.fullscreen = !!document.fullscreenElement;
      });
    }
  }
});

