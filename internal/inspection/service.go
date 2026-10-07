package inspection

import (
	"time"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/store"
)

// Service 接收巡检、完成范围校验并派生关注项。
type Service struct {
	store *store.Memory
	clock func() time.Time
}

func New(memory *store.Memory, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: memory, clock: clock}
}

type RecordInput struct {
	ObservedAt     time.Time
	ColonyStrength int
	BroodPattern   string
	Temper         string
	MiteCount      int
	Note           string
}

func (s *Service) Record(hiveID string, input RecordInput) (domain.Inspection, error) {
	observedAt := input.ObservedAt
	if observedAt.IsZero() {
		observedAt = s.clock()
	}
	record, err := domain.NewInspection(
		hiveID,
		observedAt,
		input.ColonyStrength,
		domain.BroodPattern(domain.TrimText(input.BroodPattern)),
		domain.Temper(domain.TrimText(input.Temper)),
		input.MiteCount,
		input.Note,
	)
	if err != nil {
		return domain.Inspection{}, err
	}
	// 在提交时按蜂箱所在片区生效的门槛派生关注项，并随巡检固化；
	// 之后的门槛调整只影响新提交的巡检，不会回改这里保存的关注项。
	thresholds, _, err := s.store.ThresholdsForHive(hiveID)
	if err != nil {
		return domain.Inspection{}, err
	}
	record.Concerns = domain.DeriveConcerns(record, thresholds)
	saved, err := s.store.AddInspection(hiveID, record)
	if err != nil {
		return domain.Inspection{}, err
	}
	return saved, nil
}

func (s *Service) List(hiveID string) ([]domain.Inspection, error) {
	return s.store.ListInspections(hiveID)
}

// ThresholdView 是片区关注项门槛的读取结果，source 标明门槛来自覆盖还是默认。
type ThresholdView struct {
	District   string                   `json:"district"`
	Source     string                   `json:"source"`
	Thresholds domain.ConcernThresholds `json:"thresholds"`
	UpdatedAt  *time.Time               `json:"updatedAt,omitempty"`
}

const (
	thresholdSourceOverride = "override"
	thresholdSourceDefault  = "default"
)

// SetDistrictThresholds 为片区设定关注项门槛覆盖。
// 片区必须已经登记过蜂场；门槛由领域层完整校验，非法时返回校验错误而不会退回默认门槛。
func (s *Service) SetDistrictThresholds(district string, thresholds domain.ConcernThresholds) (ThresholdView, error) {
	district = domain.TrimText(district)
	if err := domain.ValidateText("district", district, 1, 60); err != nil {
		return ThresholdView{}, err
	}
	if !s.store.DistrictExists(district) {
		return ThresholdView{}, domain.NotFound("片区", district)
	}
	entry, err := domain.NewDistrictThresholds(district, thresholds, s.clock())
	if err != nil {
		return ThresholdView{}, err
	}
	s.store.PutDistrictThresholds(entry)
	updatedAt := entry.UpdatedAt
	return ThresholdView{
		District:   entry.District,
		Source:     thresholdSourceOverride,
		Thresholds: entry.Thresholds,
		UpdatedAt:  &updatedAt,
	}, nil
}

// EffectiveDistrictThresholds 返回片区当前生效的关注项门槛：
// 有覆盖时返回覆盖内容，没有覆盖时明确标注 source=default，而不是伪造一份覆盖记录。
func (s *Service) EffectiveDistrictThresholds(district string) (ThresholdView, error) {
	district = domain.TrimText(district)
	if err := domain.ValidateText("district", district, 1, 60); err != nil {
		return ThresholdView{}, err
	}
	if entry, ok := s.store.GetDistrictThresholds(district); ok {
		updatedAt := entry.UpdatedAt
		return ThresholdView{
			District:   entry.District,
			Source:     thresholdSourceOverride,
			Thresholds: entry.Thresholds,
			UpdatedAt:  &updatedAt,
		}, nil
	}
	return ThresholdView{
		District:   district,
		Source:     thresholdSourceDefault,
		Thresholds: domain.DefaultConcernThresholds(),
	}, nil
}
