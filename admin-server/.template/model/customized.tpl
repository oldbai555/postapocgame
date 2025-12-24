// 分页查询：根据页码和每页数量查询数据
func (m *default{{.upperStartCamelObject}}Model) FindPage(ctx context.Context, page, pageSize int64) ([]{{.upperStartCamelObject}}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	offset := (page - 1) * pageSize
	
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}Rows, "deleted_at")
	whereClause := ""
	if hasDeletedAt {
		whereClause = "where deleted_at = 0"
	}
	
	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s %s", m.table, whereClause)
	{{if .withCache}}err := m.QueryRowNoCacheCtx(ctx, &total, countQuery){{else}}err := m.conn.QueryRowCtx(ctx, &total, countQuery){{end}}
	if err != nil {
		return nil, 0, err
	}
	
	// 查询分页数据
	var list []{{.upperStartCamelObject}}
	query := fmt.Sprintf("select %s from %s %s order by id desc limit %d offset %d", {{.lowerStartCamelObject}}Rows, m.table, whereClause, pageSize, offset)
	{{if .withCache}}err = m.QueryRowsNoCacheCtx(ctx, &list, query){{else}}err = m.conn.QueryRowsCtx(ctx, &list, query){{end}}
	if err != nil {
		return nil, 0, err
	}
	
	return list, total, nil
}

// 分片查询：基于lastId的分片查询，一次查询limit条，返回数据和下次查询的lastId
// lastId=0表示第一次查询，返回的lastId用于下次查询，当返回的lastId=0或数据为空时表示无更多数据
func (m *default{{.upperStartCamelObject}}Model) FindChunk(ctx context.Context, limit int64, lastId uint64) ([]{{.upperStartCamelObject}}, uint64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	var list []{{.upperStartCamelObject}}
	var err error
	
	// 检查是否有 deleted_at 字段
	hasDeletedAt := strings.Contains({{.lowerStartCamelObject}}Rows, "deleted_at")
	whereClause := ""
	if hasDeletedAt {
		whereClause = "where deleted_at = 0"
	}
	
	if lastId == 0 {
		// 第一次查询
		query := fmt.Sprintf("select %s from %s %s order by id asc limit %d", {{.lowerStartCamelObject}}Rows, m.table, whereClause, limit)
		{{if .withCache}}err = m.QueryRowsNoCacheCtx(ctx, &list, query){{else}}err = m.conn.QueryRowsCtx(ctx, &list, query){{end}}
	} else {
		// 基于lastId的分片查询
		andClause := ""
		if hasDeletedAt {
			andClause = "where deleted_at = 0 and id > {{if .postgreSql}}$1{{else}}?{{end}}"
		} else {
			andClause = "where id > {{if .postgreSql}}$1{{else}}?{{end}}"
		}
		query := fmt.Sprintf("select %s from %s %s order by id asc limit %d", {{.lowerStartCamelObject}}Rows, m.table, andClause, limit)
		{{if .withCache}}err = m.QueryRowsNoCacheCtx(ctx, &list, query, lastId){{else}}err = m.conn.QueryRowsCtx(ctx, &list, query, lastId){{end}}
	}
	
	if err != nil {
		return nil, 0, err
	}
	
	// 返回最后一条记录的ID，用于下次查询；如果数据为空，返回0表示无更多数据
	nextLastId := uint64(0)
	if len(list) > 0 {
		nextLastId = list[len(list)-1].{{.upperStartCamelPrimaryKey}}
	}
	
	return list, nextLastId, nil
}

