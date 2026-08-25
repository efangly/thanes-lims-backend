package testresult

import "gorm.io/gorm"

type Model struct {
	ID       string `gorm:"primaryKey"`
	SampleID string
	TestName string
	Analyst  string
	Result   string
	Flag     string
	RefRange string
	Status   string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteTestResult use case exists yet, but the column is in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "test_results" }
