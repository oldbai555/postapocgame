{{if .hasMiddleware}}
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.PerformanceMiddleware, serverCtx.RateLimitMiddleware, serverCtx.AuthMiddleware, serverCtx.PermissionMiddleware, serverCtx.OperationLogMiddleware},
			{{.routes}},
		),
		{{.jwt}}{{.signature}} {{.prefix}} {{.timeout}} {{.maxBytes}} {{.sse}}
	)
{{else}}
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.PerformanceMiddleware, serverCtx.RateLimitMiddleware},
			{{.routes}},
		),
		{{.jwt}}{{.signature}} {{.prefix}} {{.timeout}} {{.maxBytes}} {{.sse}}
	)
{{end}}
