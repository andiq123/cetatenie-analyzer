package decree

type FindState int

const (
	StateNotFound FindState = iota
	StateFoundButNotResolved
	StateFoundAndResolved
)
