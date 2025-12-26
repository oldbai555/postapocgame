import webapi from "./gocliRequest"
import * as components from "./adminComponents"
export * from "./adminComponents"

/**
 * @description 
 * @param req
 */
export function apiList(req: components.ApiListReq) {
	return webapi.get<components.ApiListResp>(`/api/v1/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function apiCreate(req: components.ApiCreateReq) {
	return webapi.post<null>(`/api/v1/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function apiUpdate(req: components.ApiUpdateReq) {
	return webapi.put<null>(`/api/v1/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function apiDelete(req: components.ApiDeleteReq) {
	return webapi.delete<null>(`/api/v1/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function login(req: components.LoginReq) {
	return webapi.post<components.TokenPair>(`/api/v1/login`, req)
}

/**
 * @description 
 * @param req
 */
export function refresh(req: components.RefreshReq) {
	return webapi.post<components.TokenPair>(`/api/v1/refresh`, req)
}

/**
 * @description 
 * @param req
 */
export function logout(req: components.LogoutReq) {
	return webapi.post<null>(`/api/v1/logout`, req)
}

/**
 * @description 
 */
export function profile() {
	return webapi.get<components.ProfileResp>(`/api/v1/profile`)
}

/**
 * @description 
 * @param req
 */
export function configList(req: components.ConfigListReq) {
	return webapi.get<components.ConfigListResp>(`/api/v1/configs`, req)
}

/**
 * @description 
 * @param req
 */
export function configCreate(req: components.ConfigCreateReq) {
	return webapi.post<null>(`/api/v1/configs`, req)
}

/**
 * @description 
 * @param req
 */
export function configUpdate(req: components.ConfigUpdateReq) {
	return webapi.put<null>(`/api/v1/configs`, req)
}

/**
 * @description 
 * @param req
 */
export function configDelete(req: components.ConfigDeleteReq) {
	return webapi.delete<null>(`/api/v1/configs`, req)
}

/**
 * @description 
 * @param req
 */
export function configGet(req: components.ConfigGetReq) {
	return webapi.get<components.ConfigGetResp>(`/api/v1/configs/get`, req)
}

/**
 * @description 
 * @param req
 */
export function departmentCreate(req: components.DepartmentCreateReq) {
	return webapi.post<null>(`/api/v1/departments`, req)
}

/**
 * @description 
 * @param req
 */
export function departmentUpdate(req: components.DepartmentUpdateReq) {
	return webapi.put<null>(`/api/v1/departments`, req)
}

/**
 * @description 
 * @param req
 */
export function departmentDelete(req: components.DepartmentDeleteReq) {
	return webapi.delete<null>(`/api/v1/departments`, req)
}

/**
 * @description 
 */
export function departmentTree() {
	return webapi.get<components.DepartmentTreeResp>(`/api/v1/departments/tree`)
}

/**
 * @description 
 * @param req
 */
export function dictGet(req: components.DictGetReq) {
	return webapi.get<components.DictGetResp>(`/api/v1/dict`, req)
}

/**
 * @description 
 * @param req
 */
export function dictItemList(req: components.DictItemListReq) {
	return webapi.get<components.DictItemListResp>(`/api/v1/dict-items`, req)
}

/**
 * @description 
 * @param req
 */
export function dictItemCreate(req: components.DictItemCreateReq) {
	return webapi.post<null>(`/api/v1/dict-items`, req)
}

/**
 * @description 
 * @param req
 */
export function dictItemUpdate(req: components.DictItemUpdateReq) {
	return webapi.put<null>(`/api/v1/dict-items`, req)
}

/**
 * @description 
 * @param req
 */
export function dictItemDelete(req: components.DictItemDeleteReq) {
	return webapi.delete<null>(`/api/v1/dict-items`, req)
}

/**
 * @description 
 * @param req
 */
export function dictTypeList(req: components.DictTypeListReq) {
	return webapi.get<components.DictTypeListResp>(`/api/v1/dict-types`, req)
}

/**
 * @description 
 * @param req
 */
export function dictTypeCreate(req: components.DictTypeCreateReq) {
	return webapi.post<null>(`/api/v1/dict-types`, req)
}

/**
 * @description 
 * @param req
 */
export function dictTypeUpdate(req: components.DictTypeUpdateReq) {
	return webapi.put<null>(`/api/v1/dict-types`, req)
}

/**
 * @description 
 * @param req
 */
export function dictTypeDelete(req: components.DictTypeDeleteReq) {
	return webapi.delete<null>(`/api/v1/dict-types`, req)
}

/**
 * @description 
 * @param req
 */
export function fileList(req: components.FileListReq) {
	return webapi.get<components.FileListResp>(`/api/v1/files`, req)
}

/**
 * @description 
 * @param req
 */
export function fileCreate(req: components.FileCreateReq) {
	return webapi.post<null>(`/api/v1/files`, req)
}

/**
 * @description 
 * @param req
 */
export function fileUpdate(req: components.FileUpdateReq) {
	return webapi.put<null>(`/api/v1/files`, req)
}

/**
 * @description 
 * @param req
 */
export function fileDelete(req: components.FileDeleteReq) {
	return webapi.delete<null>(`/api/v1/files`, req)
}

/**
 * @description 
 */
export function fileDownload() {
	return webapi.get<components.FileDownloadResp>(`/api/v1/files/${id}/download`)
}

/**
 * @description 
 */
export function fileUpload() {
	return webapi.post<components.FileUploadResp>(`/api/v1/files/upload`)
}

/**
 * @description 
 * @param req
 */
export function menuCreate(req: components.MenuCreateReq) {
	return webapi.post<null>(`/api/v1/menus`, req)
}

/**
 * @description 
 * @param req
 */
export function menuUpdate(req: components.MenuUpdateReq) {
	return webapi.put<null>(`/api/v1/menus`, req)
}

/**
 * @description 
 * @param req
 */
export function menuDelete(req: components.MenuDeleteReq) {
	return webapi.delete<null>(`/api/v1/menus`, req)
}

/**
 * @description 
 */
export function menuMyTree() {
	return webapi.get<components.MenuTreeResp>(`/api/v1/menus/my-tree`)
}

/**
 * @description 
 */
export function menuTree() {
	return webapi.get<components.MenuTreeResp>(`/api/v1/menus/tree`)
}

/**
 * @description 
 * @param req
 */
export function permissionList(req: components.PermissionListReq) {
	return webapi.get<components.PermissionListResp>(`/api/v1/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionCreate(req: components.PermissionCreateReq) {
	return webapi.post<null>(`/api/v1/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionUpdate(req: components.PermissionUpdateReq) {
	return webapi.put<null>(`/api/v1/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionDelete(req: components.PermissionDeleteReq) {
	return webapi.delete<null>(`/api/v1/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionApiList(req: components.PermissionApiListReq) {
	return webapi.get<components.PermissionApiListResp>(`/api/v1/permissions/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionApiUpdate(req: components.PermissionApiUpdateReq) {
	return webapi.put<null>(`/api/v1/permissions/apis`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionMenuList(req: components.PermissionMenuListReq) {
	return webapi.get<components.PermissionMenuListResp>(`/api/v1/permissions/menus`, req)
}

/**
 * @description 
 * @param req
 */
export function permissionMenuUpdate(req: components.PermissionMenuUpdateReq) {
	return webapi.put<null>(`/api/v1/permissions/menus`, req)
}

/**
 * @description 
 */
export function ping() {
	return webapi.get<components.PingResp>(`/api/v1/ping`)
}

/**
 * @description 
 * @param req
 */
export function roleList(req: components.RoleListReq) {
	return webapi.get<components.RoleListResp>(`/api/v1/roles`, req)
}

/**
 * @description 
 * @param req
 */
export function roleCreate(req: components.RoleCreateReq) {
	return webapi.post<null>(`/api/v1/roles`, req)
}

/**
 * @description 
 * @param req
 */
export function roleUpdate(req: components.RoleUpdateReq) {
	return webapi.put<null>(`/api/v1/roles`, req)
}

/**
 * @description 
 * @param req
 */
export function roleDelete(req: components.RoleDeleteReq) {
	return webapi.delete<null>(`/api/v1/roles`, req)
}

/**
 * @description 
 * @param req
 */
export function rolePermissionList(req: components.RolePermissionListReq) {
	return webapi.get<components.RolePermissionListResp>(`/api/v1/roles/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function rolePermissionUpdate(req: components.RolePermissionUpdateReq) {
	return webapi.put<null>(`/api/v1/roles/permissions`, req)
}

/**
 * @description 
 * @param req
 */
export function userList(req: components.UserListReq) {
	return webapi.get<components.UserListResp>(`/api/v1/users`, req)
}

/**
 * @description 
 * @param req
 */
export function userCreate(req: components.UserCreateReq) {
	return webapi.post<null>(`/api/v1/users`, req)
}

/**
 * @description 
 * @param req
 */
export function userUpdate(req: components.UserUpdateReq) {
	return webapi.put<null>(`/api/v1/users`, req)
}

/**
 * @description 
 * @param req
 */
export function userDelete(req: components.UserDeleteReq) {
	return webapi.delete<null>(`/api/v1/users`, req)
}

/**
 * @description 
 * @param req
 */
export function userRoleList(req: components.UserRoleListReq) {
	return webapi.get<components.UserRoleListResp>(`/api/v1/users/roles`, req)
}

/**
 * @description 
 * @param req
 */
export function userRoleUpdate(req: components.UserRoleUpdateReq) {
	return webapi.put<null>(`/api/v1/users/roles`, req)
}

/**
 * @description 刷新缓存（字典和配置缓存）
 */
export function cacheRefresh() {
	return webapi.post<components.CacheRefreshResp>(`/api/v1/cache/refresh`)
}
