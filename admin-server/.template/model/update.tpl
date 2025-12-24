func (m *default{{.upperStartCamelObject}}Model) Update(ctx context.Context, {{if .containsIndexCache}}newData{{else}}data{{end}} *{{.upperStartCamelObject}}) error {
	{{if .withCache}}{{if .containsIndexCache}}data, err:=m.FindOne(ctx, newData.{{.upperStartCamelPrimaryKey}})
	if err!=nil{
		return err
	}

{{end}}	{{.keys}}
    _, {{if .containsIndexCache}}err{{else}}err:{{end}}= m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		// 自动设置更新时间
		{{if .containsIndexCache}}newData{{else}}data{{end}}.UpdatedAt = time.Now().Unix()
		// 检查是否有 deleted_at 字段
		hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}RowsWithPlaceHolder, "deleted_at")
		// 手动构建包含 updated_at 的更新语句
		whereClause := "where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}"
		if hasDeletedAt {
			whereClause += " and deleted_at = 0"
		}
		query := fmt.Sprintf("update %s set %s, `updated_at` = %d %s", m.table, {{.lowerStartCamelObject}}RowsWithPlaceHolder, {{if .containsIndexCache}}newData{{else}}data{{end}}.UpdatedAt, whereClause)
		return conn.ExecCtx(ctx, query, {{.expressionValues}})
	}, {{.keyValues}}){{else}}// 自动设置更新时间
	data.UpdatedAt = time.Now().Unix()
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}RowsWithPlaceHolder, "deleted_at")
	// 手动构建包含 updated_at 的更新语句
	whereClause := "where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}"
	if hasDeletedAt {
		whereClause += " and deleted_at = 0"
	}
	query := fmt.Sprintf("update %s set %s, `updated_at` = %d %s", m.table, {{.lowerStartCamelObject}}RowsWithPlaceHolder, data.UpdatedAt, whereClause)
    _,err:=m.conn.ExecCtx(ctx, query, {{.expressionValues}}){{end}}
	return err
}
