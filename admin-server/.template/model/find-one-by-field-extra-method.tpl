func (m *default{{.upperStartCamelObject}}Model) formatPrimary(primary any) string {
	return fmt.Sprintf("%s%v", {{.primaryKeyLeft}}, primary)
}

func (m *default{{.upperStartCamelObject}}Model) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}Rows, "deleted_at")
	andClause := ""
	if hasDeletedAt {
		andClause = " and deleted_at = 0"
	}
	query := fmt.Sprintf("select %s from %s where {{.originalPrimaryField}} = {{if .postgreSql}}$1{{else}}?{{end}}%s limit 1", {{.lowerStartCamelObject}}Rows, m.table, andClause)
	return conn.QueryRowCtx(ctx, v, query, primary)
}
