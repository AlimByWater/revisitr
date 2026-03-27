package entity

import (
	"encoding/json"
	"fmt"
	"time"
)

// ── RFM Segment constants (v2: 7 segments) ─────────────────────────────────

const (
	RFMSegmentNew       = "new"
	RFMSegmentPromising = "promising"
	RFMSegmentRegular   = "regular"
	RFMSegmentVIP       = "vip"
	RFMSegmentRareValue = "rare_valuable"
	RFMSegmentChurnRisk = "churn_risk"
	RFMSegmentLost      = "lost"
)

// Deprecated v1 segment constants are in segment.go (RFMChampions, etc.)

// AllRFMSegments returns the 7 v2 segments in priority order.
func AllRFMSegments() []string {
	return []string{
		RFMSegmentNew,
		RFMSegmentLost,
		RFMSegmentChurnRisk,
		RFMSegmentVIP,
		RFMSegmentRegular,
		RFMSegmentRareValue,
		RFMSegmentPromising,
	}
}

// SegmentNames maps segment keys to human-readable Russian names.
var SegmentNames = map[string]string{
	RFMSegmentNew:       "Новые",
	RFMSegmentPromising: "Перспективные",
	RFMSegmentRegular:   "Регулярные",
	RFMSegmentVIP:       "VIP / Ядро",
	RFMSegmentRareValue: "Редкие, но ценные",
	RFMSegmentChurnRisk: "На грани оттока",
	RFMSegmentLost:      "Потерянные",
}

// ── RFM Templates ───────────────────────────────────────────────────────────

const (
	TemplateTypeStandard = "standard"
	TemplateTypeCustom   = "custom"
)

// RFMTemplate defines scoring thresholds for Recency and Frequency.
//
// RThresholds: [R5_max, R4_max, R3_max, R2_max] (days, ascending)
//   - R=5 if days <= RThresholds[0]
//   - R=4 if days <= RThresholds[1]
//   - R=3 if days <= RThresholds[2]
//   - R=2 if days <= RThresholds[3]
//   - R=1 otherwise
//
// FThresholds: [F5_min, F4_min, F3_min, F2_min] (visits, descending)
//   - F=5 if count >= FThresholds[0]
//   - F=4 if count >= FThresholds[1]
//   - F=3 if count >= FThresholds[2]
//   - F=2 if count >= FThresholds[3]
//   - F=1 otherwise
type RFMTemplate struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RThresholds [4]int `json:"r_thresholds"`
	FThresholds [4]int `json:"f_thresholds"`
}

// StandardTemplates contains the 4 built-in templates.
var StandardTemplates = map[string]RFMTemplate{
	"coffeegng": {
		Key:         "coffeegng",
		Name:        "Кофейни и Grab&Go",
		Description: "Высокая частота коротких визитов, небольшой средний чек",
		RThresholds: [4]int{3, 7, 14, 30},
		FThresholds: [4]int{12, 8, 4, 2},
	},
	"qsr": {
		Key:         "qsr",
		Name:        "Быстрое питание",
		Description: "Средне-высокая частота, быстрые повторные визиты",
		RThresholds: [4]int{5, 10, 21, 45},
		FThresholds: [4]int{9, 6, 3, 2},
	},
	"tsr": {
		Key:         "tsr",
		Name:        "Кафе и рестораны",
		Description: "Средняя частота визитов, более высокий средний чек",
		RThresholds: [4]int{10, 21, 45, 90},
		FThresholds: [4]int{6, 4, 3, 2},
	},
	"bar": {
		Key:         "bar",
		Name:        "Бары и пабы",
		Description: "Вечерний формат, событийные и выходные визиты",
		RThresholds: [4]int{7, 21, 45, 75},
		FThresholds: [4]int{8, 5, 3, 2},
	},
}

// StandardTemplateKeys returns template keys in display order.
func StandardTemplateKeys() []string {
	return []string{"coffeegng", "qsr", "tsr", "bar"}
}

// ── RFM Config ──────────────────────────────────────────────────────────────

// RFMConfig — per-org configuration for RFM calculation.
type RFMConfig struct {
	ID               int        `db:"id"                json:"id"`
	OrgID            int        `db:"org_id"            json:"org_id"`
	PeriodDays       int        `db:"period_days"       json:"period_days"`
	RecalcInterval   string     `db:"recalc_interval"   json:"recalc_interval"`
	LastCalcAt       *time.Time `db:"last_calc_at"      json:"last_calc_at,omitempty"`
	ClientsProcessed int        `db:"clients_processed" json:"clients_processed"`
	CreatedAt        time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"        json:"updated_at"`

	// Template fields (added in migration 00030)
	ActiveTemplateType string          `db:"active_template_type" json:"active_template_type"`
	ActiveTemplateKey  string          `db:"active_template_key"  json:"active_template_key"`
	CustomTemplateName *string         `db:"custom_template_name" json:"custom_template_name,omitempty"`
	CustomRThresholds  json.RawMessage `db:"custom_r_thresholds"  json:"custom_r_thresholds,omitempty"`
	CustomFThresholds  json.RawMessage `db:"custom_f_thresholds"  json:"custom_f_thresholds,omitempty"`
}

// ActiveTemplate returns the resolved RFMTemplate for this config.
func (c *RFMConfig) ActiveTemplate() (RFMTemplate, bool) {
	if c.ActiveTemplateType == TemplateTypeStandard {
		t, ok := StandardTemplates[c.ActiveTemplateKey]
		return t, ok
	}

	// Custom template
	var rTh, fTh [4]int
	if err := json.Unmarshal(c.CustomRThresholds, &rTh); err != nil {
		return RFMTemplate{}, false
	}
	if err := json.Unmarshal(c.CustomFThresholds, &fTh); err != nil {
		return RFMTemplate{}, false
	}

	name := c.ActiveTemplateKey
	if c.CustomTemplateName != nil {
		name = *c.CustomTemplateName
	}

	return RFMTemplate{
		Key:         "custom",
		Name:        name,
		RThresholds: rTh,
		FThresholds: fTh,
	}, true
}

type UpdateRFMConfigRequest struct {
	PeriodDays     *int    `json:"period_days,omitempty"`
	RecalcInterval *string `json:"recalc_interval,omitempty"`
}

// SetTemplateRequest is the request body for PUT /api/v1/rfm/template.
type SetTemplateRequest struct {
	TemplateType string  `json:"template_type" binding:"required,oneof=standard custom"`
	TemplateKey  string  `json:"template_key"  binding:"required_if=TemplateType standard"`
	CustomName   *string `json:"custom_name,omitempty"`
	RThresholds  *[4]int `json:"r_thresholds,omitempty"`
	FThresholds  *[4]int `json:"f_thresholds,omitempty"`
}

// ── RFM History & Dashboard ─────────────────────────────────────────────────

// RFMHistory — snapshot of segment distribution for trend tracking.
type RFMHistory struct {
	ID           int       `db:"id"            json:"id"`
	OrgID        int       `db:"org_id"        json:"org_id"`
	Segment      string    `db:"segment"       json:"segment"`
	ClientCount  int       `db:"client_count"  json:"client_count"`
	CalculatedAt time.Time `db:"calculated_at" json:"calculated_at"`
}

// RFMDashboard — response for the RFM dashboard endpoint.
type RFMDashboard struct {
	Segments []RFMSegmentSummary `json:"segments"`
	Trends   []RFMHistory        `json:"trends"`
	Config   *RFMConfig          `json:"config,omitempty"`
}

type RFMSegmentSummary struct {
	Segment     string  `json:"segment"      db:"segment"`
	ClientCount int     `json:"client_count" db:"client_count"`
	Percentage  float64 `json:"percentage"`
	AvgCheck    float64 `json:"avg_check"    db:"avg_check"`
	TotalCheck  float64 `json:"total_check"  db:"total_check"`
}

// ── RFM Segment Detail ──────────────────────────────────────────────────────

// SegmentClientRow represents a single client in the segment detail view.
type SegmentClientRow struct {
	ID                 int        `json:"id"                   db:"id"`
	FirstName          string     `json:"first_name"           db:"first_name"`
	LastName           string     `json:"last_name"            db:"last_name"`
	Phone              string     `json:"phone"                db:"phone"`
	RScore             *int       `json:"r_score"              db:"r_score"`
	FScore             *int       `json:"f_score"              db:"f_score"`
	MScore             *int       `json:"m_score"              db:"m_score"`
	RecencyDays        *int       `json:"recency_days"         db:"recency_days"`
	FrequencyCount     *int       `json:"frequency_count"      db:"frequency_count"`
	MonetarySum        *float64   `json:"monetary_sum"         db:"monetary_sum"`
	LastVisitDate      *time.Time `json:"last_visit_date"      db:"last_visit_date"`
	TotalVisitsLifetime int      `json:"total_visits_lifetime" db:"total_visits_lifetime"`
}

// SegmentClientsResponse is the response for GET /segments/:segment/clients.
type SegmentClientsResponse struct {
	Segment     string             `json:"segment"`
	SegmentName string             `json:"segment_name"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PerPage     int                `json:"per_page"`
	Clients     []SegmentClientRow `json:"clients"`
}

// ── RFM Onboarding ──────────────────────────────────────────────────────────

type OnboardingQuestion struct {
	ID      int                `json:"id"`
	Text    string             `json:"text"`
	Answers []OnboardingAnswer `json:"answers"`
}

type OnboardingAnswer struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type TemplateRecommendation struct {
	Recommended *RFMTemplate   `json:"recommended"`
	Alternative *RFMTemplate   `json:"alternative,omitempty"`
	AllScores   map[string]int `json:"all_scores"`
}

// ValidateCustomThresholds checks that custom thresholds are valid.
func (r *SetTemplateRequest) ValidateCustomThresholds() error {
	if r.TemplateType != TemplateTypeCustom {
		return nil
	}
	if r.RThresholds == nil || r.FThresholds == nil {
		return fmt.Errorf("custom template requires r_thresholds and f_thresholds")
	}

	// R thresholds: strictly ascending, all >= 0
	rt := r.RThresholds
	for i := 0; i < 4; i++ {
		if rt[i] < 0 {
			return fmt.Errorf("r_thresholds[%d] must be >= 0", i)
		}
		if i > 0 && rt[i] <= rt[i-1] {
			return fmt.Errorf("r_thresholds must be strictly ascending")
		}
	}

	// F thresholds: strictly descending, all >= 1
	ft := r.FThresholds
	for i := 0; i < 4; i++ {
		if ft[i] < 1 {
			return fmt.Errorf("f_thresholds[%d] must be >= 1", i)
		}
		if i > 0 && ft[i] >= ft[i-1] {
			return fmt.Errorf("f_thresholds must be strictly descending")
		}
	}

	return nil
}

// ── RFM Onboarding Data ──────────────────────────────────────────────────────

// OnboardingQuestions returns the 3 quiz questions for template recommendation.
func GetOnboardingQuestions() []OnboardingQuestion {
	return []OnboardingQuestion{
		{
			ID:   1,
			Text: "Как чаще всего гости используют ваше заведение?",
			Answers: []OnboardingAnswer{
				{ID: 1, Text: "Берут кофе, снеки или еду с собой"},
				{ID: 2, Text: "Быстро заказывают и недолго остаются"},
				{ID: 3, Text: "Приходят посидеть, поесть, провести время"},
				{ID: 4, Text: "Чаще приходят за напитками, вечером или для общения"},
			},
		},
		{
			ID:   2,
			Text: "Как часто ваши постоянные гости обычно возвращаются?",
			Answers: []OnboardingAnswer{
				{ID: 1, Text: "Почти каждый день или несколько раз в неделю"},
				{ID: 2, Text: "2–3 раза в неделю"},
				{ID: 3, Text: "1–2 раза в неделю или реже"},
				{ID: 4, Text: "Раз в неделю, чаще по выходным или на события"},
			},
		},
		{
			ID:   3,
			Text: "Какой средний чек у вашего заведения?",
			Answers: []OnboardingAnswer{
				{ID: 1, Text: "До 300 ₽ — кофе, выпечка, мелкие покупки"},
				{ID: 2, Text: "300–600 ₽ — фастфуд, комбо, быстрый перекус"},
				{ID: 3, Text: "600–1500 ₽ — обед или ужин в кафе/ресторане"},
				{ID: 4, Text: "1000+ ₽ — бар, ресторан, вечерний формат"},
			},
		},
	}
}

// answerToTemplate maps answer IDs (1-4) to template keys for all questions.
var answerToTemplate = map[int]string{
	1: "coffeegng",
	2: "qsr",
	3: "tsr",
	4: "bar",
}

// RecommendTemplate scores answers and returns the best template recommendation.
func RecommendTemplate(answers []int) (*TemplateRecommendation, error) {
	if len(answers) != 3 {
		return nil, fmt.Errorf("expected 3 answers, got %d", len(answers))
	}

	scores := map[string]int{}
	for _, key := range StandardTemplateKeys() {
		scores[key] = 0
	}

	for _, answerID := range answers {
		key, ok := answerToTemplate[answerID]
		if !ok {
			return nil, fmt.Errorf("invalid answer ID: %d", answerID)
		}
		scores[key]++
	}

	// Find winner (max score); tie-break by Q1 answer
	bestKey := answerToTemplate[answers[0]]
	bestScore := scores[bestKey]
	for _, key := range StandardTemplateKeys() {
		if scores[key] > bestScore {
			bestScore = scores[key]
			bestKey = key
		}
	}

	recommended := StandardTemplates[bestKey]
	result := &TemplateRecommendation{
		Recommended: &recommended,
		AllScores:   scores,
	}

	// Find alternative: second highest score (if > 0 and != winner)
	var altKey string
	var altScore int
	for _, key := range StandardTemplateKeys() {
		if key == bestKey {
			continue
		}
		if scores[key] > altScore {
			altScore = scores[key]
			altKey = key
		}
	}
	if altScore > 0 {
		alt := StandardTemplates[altKey]
		result.Alternative = &alt
	}

	return result, nil
}

// ── Client RFM Stats (for recalculation) ────────────────────────────────────

// ClientRFMStats holds raw metrics for a single client, used by the RFM service.
type ClientRFMStats struct {
	ClientID           int       `db:"client_id"`
	LastVisitAt        time.Time `db:"last_visit_at"`
	FrequencyCount     int       `db:"frequency_count"`      // visits in last 90 days
	MonetarySum        float64   `db:"monetary_sum"`          // revenue in last 180 days
	TotalVisitsLifetime int      `db:"total_visits_lifetime"` // all-time visits
}

// RFMUpdateParams holds all values to update on a bot_client after RFM recalc.
// Fields are non-nullable because the caller (RFM service) always computes concrete values.
// This differs from BotClient where fields are *int/*float64 to represent "not yet calculated".
type RFMUpdateParams struct {
	ClientID        int
	RScore          int
	FScore          int
	MScore          int
	RecencyDays     int
	FrequencyCount  int
	MonetarySum     float64
	TotalVisitsLife int
	LastVisitDate   time.Time
	Segment         string
}
