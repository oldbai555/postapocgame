func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, data *{{.upperStartCamelObject}}) (sql.Result,error) {
	// 自动设置创建时间和更新时间
	if data.CreatedAt == 0 {
		data.CreatedAt = time.Now().Unix()
	}
	if data.UpdatedAt == 0 {
		data.UpdatedAt = time.Now().Unix()
	}
	// 注意：如果表没有 deleted_at 字段，data.DeletedAt 将不存在，不能访问
	// deleted_at 字段如果存在，它已经在 RowsExpectAutoSet 中，不需要特殊处理
	{{if .withCache}}{{.keys}}
    ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		// 手动构建包含 created_at、updated_at 的插入语句
		// 如果表有 deleted_at 字段，它已经在 RowsExpectAutoSet 中，不需要重复添加
		query := fmt.Sprintf("insert into %s (%s, `created_at`, `updated_at`) values ({{.expression}}, ?, ?)", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, {{.expressionValues}}, data.CreatedAt, data.UpdatedAt)
	}, {{.keyValues}}){{else}}// 手动构建包含 created_at、updated_at 的插入语句
	// 如果表有 deleted_at 字段，它已经在 RowsExpectAutoSet 中，不需要重复添加
	query := fmt.Sprintf("insert into %s (%s, `created_at`, `updated_at`) values ({{.expression}}, ?, ?)", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
    ret,err:=m.conn.ExecCtx(ctx, query, {{.expressionValues}}, data.CreatedAt, data.UpdatedAt){{end}}
	return ret,err
}
