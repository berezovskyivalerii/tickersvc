package lists

import "context"

type Def struct {
	ID          int16
	Slug        string
	SourceID    int16
	SourceSlug  string
	TargetID    int16
	TargetSlug  string
}

type DefsRepo interface {
    // уже есть:
    Find(ctx context.Context, sourceSlug, targetSlug *string) ([]Def, error)
    GetByID(ctx context.Context, id int16) (Def, error)

    IDsBySlugs(ctx context.Context, slugs []string) (map[string]int16, error)
}

type QueryRepo interface {
	// List text (ready strings "spot, futures|none"), sorted
	GetTextBySlug(ctx context.Context, slug string) ([]string, error)
	// For /lists/:target - return by sources
	GetTextByTarget(ctx context.Context, targetSlug string) (map[string][]string, error)
	// For /lists - return nested structure target -> source -> lines
	GetAllText(ctx context.Context) (map[string]map[string][]string, error)
	
	GetRowsBySlug(ctx context.Context, slug string) ([]Row, error)
}
