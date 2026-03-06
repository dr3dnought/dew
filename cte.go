package dew

type cteClause struct {
	name      string
	query     Expression
	recursive bool
}

func CTE(name string, query Expression) cteClause {
	return cteClause{name: name, query: query}
}

func RecursiveCTE(name string, query Expression) cteClause {
	return cteClause{name: name, query: query, recursive: true}
}
