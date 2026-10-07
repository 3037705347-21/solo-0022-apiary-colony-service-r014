package domain

// ConcernThresholds 是派生巡检关注项使用的数值门槛。
// 未设置片区覆盖时巡检按默认门槛派生；覆盖一旦保存，只影响之后提交的巡检，
// 已经固化的巡检关注项不随阈值调整而改变。
type ConcernThresholds struct {
	MiteLoadThreshold  int `json:"miteLoadThreshold"`
	LowStrengthCeiling int `json:"lowStrengthCeiling"`
}

// NewConcernThresholds 校验并建立一份关注项阈值覆盖。
// 任一门槛越界都会返回校验错误，调用方不得在此之外悄悄退回默认门槛。
func NewConcernThresholds(miteLoadThreshold, lowStrengthCeiling int) (ConcernThresholds, error) {
	if err := ValidateIntRange("miteLoadThreshold", miteLoadThreshold, MinMiteCount, MaxMiteCount); err != nil {
		return ConcernThresholds{}, err
	}
	if err := ValidateIntRange("lowStrengthCeiling", lowStrengthCeiling, MinColonyStrength, MaxColonyStrength); err != nil {
		return ConcernThresholds{}, err
	}
	return ConcernThresholds{
		MiteLoadThreshold:  miteLoadThreshold,
		LowStrengthCeiling: lowStrengthCeiling,
	}, nil
}
