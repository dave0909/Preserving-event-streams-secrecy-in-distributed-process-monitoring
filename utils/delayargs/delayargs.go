package delayargs

type ArrivalArgs struct {
	EventCode int
}

type CompletionArgs struct {
	EventCode int
}

type TimestampResponse struct {
	Timestamp int64
}
type MemoryData struct {
	Timestamp   int
	MemoryUsage int
}
