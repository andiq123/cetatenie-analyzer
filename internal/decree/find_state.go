package decree

type FindState int

const (
	StateNotFound FindState = iota
	StateFoundButNotResolved
	StateFoundAndResolved
)

func (s FindState) String() string {
	switch s {
	case StateNotFound:
		return "Not Found"
	case StateFoundButNotResolved:
		return "Found but not resolved"
	case StateFoundAndResolved:
		return "Found and resolved"
	default:
		return "Unknown state"
	}
}
