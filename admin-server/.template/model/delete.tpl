func (m *default{{.upperStartCamelObject}}Model) Delete(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	{{if .withCache}}{{if .containsIndexCache}}data, err:=m.FindOne(ctx, {{.lowerStartCamelPrimaryKey}})
	if err!=nil{
		return err
	}

{{end}}	{{.keys}}
    _, err {{if .containsIndexCache}}={{else}}:={{end}} m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		// 检查是否有 deleted_at 字段：通过检查结构体字段名列表
		fieldNames := builder.RawFieldNames(&{{.upperStartCamelObject}}{})
		fieldNamesStr := strings.Join(fieldNames, ",")
		hasDeletedAt := strings.Contains(fieldNamesStr, "deleted_at")
		
		var query string
		if hasDeletedAt {
			// 软删除：设置 deleted_at 为当前时间戳
			query = fmt.Sprintf("update %s set deleted_at = UNIX_TIMESTAMP(), updated_at = UNIX_TIMESTAMP() where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}} and deleted_at = 0", m.table)
		} else {
			// 物理删除：直接删除记录
			query = fmt.Sprintf("delete from %s where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}", m.table)
		}
		return conn.ExecCtx(ctx, query, {{.lowerStartCamelPrimaryKey}})
	}, {{.keyValues}}){{else}}// 检查是否有 deleted_at 字段：通过检查结构体字段名列表
	fieldNames := builder.RawFieldNames(&{{.upperStartCamelObject}}{})
	fieldNamesStr := strings.Join(fieldNames, ",")
	hasDeletedAt := strings.Contains(fieldNamesStr, "deleted_at")
	
	var query string
	if hasDeletedAt {
		// 软删除：设置 deleted_at 为当前时间戳
		query = fmt.Sprintf("update %s set deleted_at = UNIX_TIMESTAMP(), updated_at = UNIX_TIMESTAMP() where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}} and deleted_at = 0", m.table)
	} else {
		// 物理删除：直接删除记录
		query = fmt.Sprintf("delete from %s where {{.originalPrimaryKey}} = {{if .postgreSql}}$1{{else}}?{{end}}", m.table)
	}
	_,err:=m.conn.ExecCtx(ctx, query, {{.lowerStartCamelPrimaryKey}}){{end}}
	return err
}
