package inspection

import "example.com/apiary-colony-service/internal/domain"

const (
	ConcernMiteLoad    = domain.ConcernMiteLoad
	ConcernBroodGap    = domain.ConcernBroodGap
	ConcernTemperRisk  = domain.ConcernTemperRisk
	ConcernLowStrength = domain.ConcernLowStrength
)

const (
	MiteLoadThreshold  = domain.DefaultMiteLoadThreshold
	LowStrengthCeiling = domain.DefaultLowStrengthCeiling
)

// DeriveConcerns 按默认门槛从一次巡检派生关注项，顺序稳定。
// 服务流程会改用具片区覆盖门槛的 domain.DeriveConcerns。
func DeriveConcerns(record domain.Inspection) []string {
	return domain.DeriveConcerns(record, domain.DefaultConcernThresholds())
}
