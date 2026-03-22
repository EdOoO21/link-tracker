package bot

type State int64

type StateInfo struct {
	State State
	URL   string
}

type StateMachine map[int64]*StateInfo

const (
	NothingGot      State = 0 // no track command got / default value
	TrackCommandGot State = 1 // track command got
	LinkGot         State = 2 // track and link got
)
