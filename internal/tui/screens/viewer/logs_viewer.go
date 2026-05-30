package viewer

type LogsViewer struct {
	SessionID                 int
	TailLines                 int
	HistoryLoad               bool
	HistoryDone               bool
	HistoryBaseLen            int
	HistoryAppendedDuringLoad int
	HistoryNoProgressCount    int
}

func NewLogsViewer() LogsViewer {
	return LogsViewer{}
}
