package testresult

type Flag string

const (
	FlagHi Flag = "hi"
	FlagLo Flag = "lo"
	FlagOk Flag = "ok"
)

type Status string

const (
	StatusAnalyzing           Status = "analyzing"
	StatusPendingVerification Status = "pending_verification"
	StatusApproved            Status = "approved"
)

type TestResult struct {
	ID       string
	SampleID string
	TestName string
	Analyst  string
	Result   string
	Flag     Flag
	RefRange string
	Status   Status
}
