{{if .hasMiddleware}}
	authMiddleware := middleware.NewAuthMiddleware(serverCtx)
	permissionMiddleware := middleware.NewPermissionMiddleware(serverCtx)
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{authMiddleware.Handle, permissionMiddleware.Handle},
			{{.routes}},
		),
		{{.jwt}}{{.signature}} {{.prefix}} {{.timeout}} {{.maxBytes}} {{.sse}}
	)
{{else}}
	server.AddRoutes(
		{{.routes}} {{.jwt}}{{.signature}} {{.prefix}} {{.timeout}} {{.maxBytes}} {{.sse}}
	)
{{end}}
