package domain

import "time"

// ConcernThresholds 是巡检关注项派生时生效的门槛集合。
// brood_gap 与 temper_risk 按固定枚举派生，因此可覆盖的只有两个数值门槛。
type ConcernThresholds struct {
	// MiteLoadAt 为记 mite_load 的螨计数门槛，螨计数大于等于它时记关注项。
	MiteLoadAt int `json:"miteLoadThreshold"`
	// LowStrengthCeiling 为记 low_strength 的群势上限，群势小于等于它时记关注项。
	LowStrengthCeiling int `json:"lowStrengthCeiling"`
}

const (
	// DefaultMiteLoadThreshold 是片区没有覆盖门槛时使用的默认螨计数门槛。
	DefaultMiteLoadThreshold = 9
	// DefaultLowStrengthCeiling 是片区没有覆盖门槛时使用的默认群势上限。
	DefaultLowStrengthCeiling = 2

	MinMiteLoadThreshold  = MinMiteCount
	MaxMiteLoadThreshold  = MaxMiteCount
	MinLowStrengthCeiling = MinColonyStrength
	MaxLowStrengthCeiling = MaxColonyStrength
)

// DistrictThresholds 是协调人为某个片区设定的关注项门槛覆盖。
type DistrictThresholds struct {
	District   string            `json:"district"`
	Thresholds ConcernThresholds `json:"thresholds"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

func DefaultConcernThresholds() ConcernThresholds {
	return ConcernThresholds{
		MiteLoadAt:         DefaultMiteLoadThreshold,
		LowStrengthCeiling: DefaultLowStrengthCeiling,
	}
}

// NewDistrictThresholds 校验片区名与各门槛取值，非法时返回带字段名的校验错误，
// 绝不静默退回默认门槛。
func NewDistrictThresholds(district string, thresholds ConcernThresholds, updatedAt time.Time) (DistrictThresholds, error) {
	if err := ValidateText("district", district, 1, 60); err != nil {
		return DistrictThresholds{}, err
	}
	if err := ValidateConcernThresholds(thresholds); err != nil {
		return DistrictThresholds{}, err
	}
	return DistrictThresholds{
		District:   TrimText(district),
		Thresholds: thresholds,
		UpdatedAt:  updatedAt.UTC(),
	}, nil
}

// ValidateConcernThresholds 校验各门槛都落在对应量测的合法范围内。
func ValidateConcernThresholds(thresholds ConcernThresholds) error {
	if err := ValidateIntRange(
		"miteLoadThreshold", thresholds.MiteLoadAt, MinMiteLoadThreshold, MaxMiteLoadThreshold,
	); err != nil {
		return err
	}
	if err := ValidateIntRange(
		"lowStrengthCeiling", thresholds.LowStrengthCeiling, MinLowStrengthCeiling, MaxLowStrengthCeiling,
	); err != nil {
		return err
	}
	return nil
}

// DeriveConcerns 按给定门槛从一次巡检派生出关注项，顺序稳定。
// 关注项只在巡检提交时固化一次，门槛调整不会影响已经保存的巡检。
func DeriveConcerns(record Inspection, thresholds ConcernThresholds) []string {
	concerns := make([]string, 0, 4)
	if record.MiteCount >= thresholds.MiteLoadAt {
		concerns = append(concerns, ConcernMiteLoad)
	}
	if record.BroodPattern == BroodNone {
		concerns = append(concerns, ConcernBroodGap)
	}
	if record.Temper == TemperAggressive {
		concerns = append(concerns, ConcernTemperRisk)
	}
	if record.ColonyStrength <= thresholds.LowStrengthCeiling {
		concerns = append(concerns, ConcernLowStrength)
	}
	return concerns
}
