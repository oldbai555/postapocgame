type (
	{{.lowerStartCamelObject}}Model interface {
		{{.method}}
		FindPage(ctx context.Context, page, pageSize int64) ([]{{.upperStartCamelObject}}, int64, error)
		FindChunk(ctx context.Context, limit int64, lastId uint64) ([]{{.upperStartCamelObject}}, uint64, error)
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}sqlc.CachedConn{{else}}conn sqlx.SqlConn{{end}}
		table string
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
