type (
	// {{.Name}}
	{{.GroupUpper}}Item {
		id        uint64 `json:"id"`
		name      string `json:"name"`
		status    int64  `json:"status"`
		createdAt string `json:"createdAt"`
	}
	{{.GroupUpper}}ListReq {
		// 注意：GET 请求的查询参数需要同时包含 json 和 form 标签
		// json 标签用于请求体（POST/PUT/DELETE），form 标签用于查询参数（GET）
		// 重要：form 标签中必须包含 optional，否则 httpx.Parse 无法正确解析查询参数
		page     int64  `json:"page,optional" form:"page,optional"`
		pageSize int64  `json:"pageSize,optional" form:"pageSize,optional"`
		name     string `json:"name,optional" form:"name,optional"`
	}
	{{.GroupUpper}}ListResp {
		total int64           `json:"total"`
		list  []{{.GroupUpper}}Item `json:"list"`
	}
	{{.GroupUpper}}CreateReq {
		name   string `json:"name"`
		status int64  `json:"status,optional"`
	}
	{{.GroupUpper}}UpdateReq {
		id     uint64 `json:"id"`
		name   string `json:"name,optional"`
		status int64  `json:"status,optional"`
	}
	{{.GroupUpper}}DeleteReq {
		id     uint64 `json:"id"`
	}
)

@server (
	group:      {{.Group}}
	prefix:     /api/v1
	middleware: AuthMiddleware,PermissionMiddleware
)
service admin-api {
	@handler {{.GroupUpper}}List
	get /{{.Group}}s ({{.GroupUpper}}ListReq) returns ({{.GroupUpper}}ListResp)

	@handler {{.GroupUpper}}Create
	post /{{.Group}}s ({{.GroupUpper}}CreateReq)

	@handler {{.GroupUpper}}Update
	put /{{.Group}}s ({{.GroupUpper}}UpdateReq)

	@handler {{.GroupUpper}}Delete
	delete /{{.Group}}s ({{.GroupUpper}}DeleteReq)
}

