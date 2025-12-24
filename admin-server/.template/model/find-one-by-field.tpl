func (m *default{{.upperStartCamelObject}}Model) FindOneBy{{.upperField}}(ctx context.Context, {{.in}}) (*{{.upperStartCamelObject}}, error) {
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}Rows, "deleted_at")
	andClause := ""
	if hasDeletedAt {
		andClause = " and deleted_at = 0"
	}
	{{if .withCache}}{{.cacheKey}}
	var resp {{.upperStartCamelObject}}
	err := m.QueryRowIndexCtx(ctx, &resp, {{.cacheKeyVariable}}, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where {{.originalField}}%s limit 1", {{.lowerStartCamelObject}}Rows, m.table, andClause)
		if err := conn.QueryRowCtx(ctx, &resp, query, {{.lowerStartCamelField}}); err != nil {
			return nil, err
		}
		return resp.{{.upperStartCamelPrimaryKey}}, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{else}}var resp {{.upperStartCamelObject}}
	query := fmt.Sprintf("select %s from %s where {{.originalField}}%s limit 1", {{.lowerStartCamelObject}}Rows, m.table, andClause)
	err := m.conn.QueryRowCtx(ctx, &resp, query, {{.lowerStartCamelField}})
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{end}}
}
