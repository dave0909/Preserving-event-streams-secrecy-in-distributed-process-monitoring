package delayargs

type ArrivalArgs struct {
	EventCode        int
	ArrivalTimestamp int
}

type CompletionArgs struct {
	EventCode           int
	CompletionTimestamp int
}

type TimestampResponse struct {
	Timestamp int64
}
type MemoryData struct {
	Timestamp   int
	MemoryUsage int
}
