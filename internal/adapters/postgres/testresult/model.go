package testresult

type Model struct {
	ID       string `gorm:"primaryKey"`
	SampleID string
	TestName string
	Analyst  string
	Result   string
	Flag     string
	RefRange string
	Status   string
}

func (Model) TableName() string { return "test_results" }
