package dbflex

// DbIndex struct to build index on database
type DbIndex struct {
	Name     string
	IsUnique bool
	Fields   []string
}
