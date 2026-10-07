package inspection

import "example.com/apiary-colony-service/internal/domain"

const (
	ConcernMiteLoad    = "mite_load"
	ConcernBroodGap    = "brood_gap"
	ConcernTemperRisk  = "temper_risk"
	ConcernLowStrength = "low_strength"
)

const (
	MiteLoadThreshold  = 9
	LowStrengthCeiling = 2
)

// DefaultThresholds 返回未设置片区覆盖时生效的默认关注项阈值。
func DefaultThresholds() domain.ConcernThresholds {
	return domain.ConcernThresholds{
		MiteLoadThreshold:  MiteLoadThreshold,
		LowStrengthCeiling: LowStrengthCeiling,
	}
}

// DeriveConcerns 按生效阈值从一次巡检派生出关注项，顺序稳定。
func DeriveConcerns(record domain.Inspection, thresholds domain.ConcernThresholds) []string {
	concerns := make([]string, 0, 4)
	if record.MiteCount >= thresholds.MiteLoadThreshold {
		concerns = append(concerns, ConcernMiteLoad)
	}
	if record.BroodPattern == domain.BroodNone {
		concerns = append(concerns, ConcernBroodGap)
	}
	if record.Temper == domain.TemperAggressive {
		concerns = append(concerns, ConcernTemperRisk)
	}
	if record.ColonyStrength <= thresholds.LowStrengthCeiling {
		concerns = append(concerns, ConcernLowStrength)
	}
	return concerns
}
