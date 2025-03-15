package models

const (
	Done  = "done"
	Retry = "retry"
	Delay = "delay"
)

type ConsumeResult struct {
	Result  string
	DelayMs int
}

func NewConsumeResultDone() *ConsumeResult {
	return &ConsumeResult{
		Result: Done,
	}
}

func NewConsumeResultRetry(delayMs int) *ConsumeResult {
	return &ConsumeResult{
		Result:  Retry,
		DelayMs: delayMs,
	}
}

func NewConsumeResultDelay(delayMs int) *ConsumeResult {
	return &ConsumeResult{
		Result:  Delay,
		DelayMs: delayMs,
	}
}
