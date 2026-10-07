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
	saved, err := s.store.AddInspection(hiveID, record, DefaultThresholds(), func(thresholds domain.ConcernThresholds) []string {
		return DeriveConcerns(record, thresholds)
	})
	if err != nil {
		return domain.Inspection{}, err
	}
	return saved, nil
}

func (s *Service) List(hiveID string) ([]domain.Inspection, error) {
	return s.store.ListInspections(hiveID)
}

type ThresholdsInput struct {
	MiteLoadThreshold  int
	LowStrengthCeiling int
}

// SetDistrictThresholds 为片区保存关注项阈值覆盖，只影响之后提交的巡检。
// 片区名或任一门槛非法时返回校验错误，不写入任何覆盖。
func (s *Service) SetDistrictThresholds(district string, input ThresholdsInput) (domain.ConcernThresholds, error) {
	trimmed := domain.TrimText(district)
	if err := domain.ValidateText("district", trimmed, 1, 60); err != nil {
		return domain.ConcernThresholds{}, err
	}
	thresholds, err := domain.NewConcernThresholds(input.MiteLoadThreshold, input.LowStrengthCeiling)
	if err != nil {
		return domain.ConcernThresholds{}, err
	}
	s.store.SetDistrictThresholds(trimmed, thresholds)
	return thresholds, nil
}

// DistrictThresholds 读回片区当前的关注项阈值覆盖；未设置时返回未找到，
// 避免调用方把默认门槛误当成已生效的覆盖。
func (s *Service) DistrictThresholds(district string) (domain.ConcernThresholds, error) {
	trimmed := domain.TrimText(district)
	thresholds, ok := s.store.DistrictThresholds(trimmed)
	if !ok {
		return domain.ConcernThresholds{}, domain.NotFound("片区阈值覆盖", trimmed)
	}
	return thresholds, nil
}
