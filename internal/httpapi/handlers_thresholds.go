package httpapi

import (
	"net/http"

	"example.com/apiary-colony-service/internal/domain"
)

type districtThresholdsRequest struct {
	// 指针字段配合 DisallowUnknownFields：缺字段时保持 nil，入口层据此拒绝请求，
	// 避免缺门槛被当成零值写入、再悄悄退回默认门槛。
	MiteLoadThreshold  *int `json:"miteLoadThreshold"`
	LowStrengthCeiling *int `json:"lowStrengthCeiling"`
}

func (s *Server) handlePutDistrictThresholds(w http.ResponseWriter, r *http.Request) {
	var request districtThresholdsRequest
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
	view, err := s.deps.Inspection.SetDistrictThresholds(r.PathValue("district"), domain.ConcernThresholds{
		MiteLoadAt:         *request.MiteLoadThreshold,
		LowStrengthCeiling: *request.LowStrengthCeiling,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	// 返回完整的生效门槛与 source=override，让调用方能确认覆盖确实生效，
	// 而不是误以为部分门槛被默认值替代。
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleGetDistrictThresholds(w http.ResponseWriter, r *http.Request) {
	view, err := s.deps.Inspection.EffectiveDistrictThresholds(r.PathValue("district"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
