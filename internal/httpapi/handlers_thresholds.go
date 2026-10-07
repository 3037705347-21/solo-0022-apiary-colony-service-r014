package httpapi

import (
	"net/http"

	"example.com/apiary-colony-service/internal/domain"
	"example.com/apiary-colony-service/internal/inspection"
)

// setDistrictThresholdsRequest 使用指针字段区分缺失与零值：
// 覆盖规则缺字段时必须显式报错，不能悄悄退回默认门槛。
type setDistrictThresholdsRequest struct {
	MiteLoadThreshold  *int `json:"miteLoadThreshold"`
	LowStrengthCeiling *int `json:"lowStrengthCeiling"`
}

type districtThresholdsResponse struct {
	District           string `json:"district"`
	MiteLoadThreshold  int    `json:"miteLoadThreshold"`
	LowStrengthCeiling int    `json:"lowStrengthCeiling"`
}

func (s *Server) handleSetDistrictThresholds(w http.ResponseWriter, r *http.Request) {
	var request setDistrictThresholdsRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	if request.MiteLoadThreshold == nil {
		writeError(w, domain.Invalid("miteLoadThreshold", "miteLoadThreshold 不能为空"))
		return
	}
	if request.LowStrengthCeiling == nil {
		writeError(w, domain.Invalid("lowStrengthCeiling", "lowStrengthCeiling 不能为空"))
		return
	}
	district := domain.TrimText(r.PathValue("district"))
	thresholds, err := s.deps.Inspection.SetDistrictThresholds(district, inspection.ThresholdsInput{
		MiteLoadThreshold:  *request.MiteLoadThreshold,
		LowStrengthCeiling: *request.LowStrengthCeiling,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, districtThresholdsResponse{
		District:           district,
		MiteLoadThreshold:  thresholds.MiteLoadThreshold,
		LowStrengthCeiling: thresholds.LowStrengthCeiling,
	})
}

func (s *Server) handleGetDistrictThresholds(w http.ResponseWriter, r *http.Request) {
	district := domain.TrimText(r.PathValue("district"))
	thresholds, err := s.deps.Inspection.DistrictThresholds(district)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, districtThresholdsResponse{
		District:           district,
		MiteLoadThreshold:  thresholds.MiteLoadThreshold,
		LowStrengthCeiling: thresholds.LowStrengthCeiling,
	})
}
