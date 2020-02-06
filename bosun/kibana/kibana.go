package kibana

type API interface {
	Info() (Info, error)

	Indices(filter string) ([]Index, error)

	IndexPatternFields(filter string) ([]IndexPattern, error)

	IndexPatterns(filter string) ([]IndexPattern, error)
}
