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

func (s StateMachine) State(chatID int64) *StateInfo {
	stateInfo, ok := s[chatID]
	if !ok || stateInfo == nil {
		stateInfo = &StateInfo{State: NothingGot}
		s[chatID] = stateInfo
	}
	return stateInfo
}

func (s StateMachine) HasActiveState(chatID int64) bool {
	stateInfo, ok := s[chatID]
	return ok && stateInfo != nil && stateInfo.State != NothingGot
}

func (s StateMachine) Reset(chatID int64) {
	delete(s, chatID)
}

func (s StateMachine) StartTrack(chatID int64) {
	stateInfo := s.State(chatID)
	stateInfo.State = TrackCommandGot
	stateInfo.URL = ""
}
