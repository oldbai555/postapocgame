import {defineStore} from 'pinia';
import {login, logout, profile, menuMyTree} from '@/api/generated/admin';
import type {LoginReq, TokenPair, ProfileResp, MenuItem} from '@/api/generated/admin';

interface UserState {
  token: string;
  refreshToken: string;
  profile: ProfileResp | null;
  permissions: string[];
  menus: MenuItem[];
  cacheAt: number;
}

const tokenKey = 'admin_token';
const refreshKey = 'admin_refresh_token';
const permKey = 'admin_permissions';
const menuKey = 'admin_menus';
const cacheAtKey = 'admin_cache_at';
const CACHE_TTL = 5 * 60 * 1000;

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: localStorage.getItem(tokenKey) || '',
    refreshToken: localStorage.getItem(refreshKey) || '',
    profile: null,
    permissions: JSON.parse(localStorage.getItem(permKey) || '[]'),
    menus: JSON.parse(localStorage.getItem(menuKey) || '[]'),
    cacheAt: Number(localStorage.getItem(cacheAtKey) || 0)
  }),
  actions: {
    async login(payload: LoginReq) {
      const data = await login(payload);
      this.token = data.accessToken;
      this.refreshToken = data.refreshToken;
      localStorage.setItem(tokenKey, this.token);
      localStorage.setItem(refreshKey, this.refreshToken);
      await this.fetchProfile(true);
      await this.fetchMenus(true);
      
      // 登录后自动连接 WebSocket（如果有权限）
      const {useWebSocketStore} = await import('./websocket');
      const wsStore = useWebSocketStore();
      wsStore.connect();
    },
    async fetchProfile(force = false) {
      if (!force && this.cacheValid()) {
        return;
      }
      const profileData = await profile();
      this.profile = profileData;
      const perms: string[] = (profileData as any)?.permissions || [];
      this.permissions = perms || [];
      this.persistCache();
    },
    async fetchMenus(force = false) {
      if (!force && this.cacheValid() && this.menus.length > 0) {
        return;
      }
      // 使用 my-tree 接口，根据用户权限过滤菜单
      const resp = await menuMyTree();
      this.menus = resp.list || [];
      this.persistCache();
    },
    async logout() {
      try {
        await logout({
          accessToken: this.token,
          refreshToken: this.refreshToken
        });
      } catch {
        // ignore
      }
      
      // 退出登录时断开 WebSocket
      const {useWebSocketStore} = await import('./websocket');
      const wsStore = useWebSocketStore();
      wsStore.disconnect();
      
      this.token = '';
      this.refreshToken = '';
      this.profile = null;
      this.permissions = [];
      this.menus = [];
      this.cacheAt = 0;
      localStorage.removeItem(tokenKey);
      localStorage.removeItem(refreshKey);
      localStorage.removeItem(permKey);
      localStorage.removeItem(menuKey);
      localStorage.removeItem(cacheAtKey);
    },
    cacheValid() {
      if (!this.cacheAt) return false;
      return Date.now() - this.cacheAt < CACHE_TTL;
    },
    persistCache() {
      this.cacheAt = Date.now();
      localStorage.setItem(permKey, JSON.stringify(this.permissions || []));
      localStorage.setItem(menuKey, JSON.stringify(this.menus || []));
      localStorage.setItem(cacheAtKey, String(this.cacheAt));
    }
  }
});

