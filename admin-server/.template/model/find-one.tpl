func (m *default{{.upperStartCamelObject}}Model) FindOne(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}Rows, "deleted_at")
	whereClause := "where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}"
	if hasDeletedAt {
		whereClause += " and deleted_at = 0"
	}
	{{if .withCache}}{{.cacheKey}}
	var resp {{.upperStartCamelObject}}
	err := m.QueryRowCtx(ctx, &resp, {{.cacheKeyVariable}}, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s %s limit 1", {{.lowerStartCamelObject}}Rows, m.table, whereClause)
		return conn.QueryRowCtx(ctx, v, query, {{.lowerStartCamelPrimaryKey}})
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{else}}query := fmt.Sprintf("select %s from %s %s limit 1", {{.lowerStartCamelObject}}Rows, m.table, whereClause)
	var resp {{.upperStartCamelObject}}
	err := m.conn.QueryRowCtx(ctx, &resp, query, {{.lowerStartCamelPrimaryKey}})
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{end}}
}
