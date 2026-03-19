package entity

import "time"

// ClientTxStats holds aggregated transaction data per client (used by RFM service).
type ClientTxStats struct {
	ClientID    int       `db:"client_id"`
	LastTxAt    time.Time `db:"last_tx_at"`
	TxCount     int       `db:"tx_count"`
	TotalAmount float64   `db:"total_amount"`
}

// AnalyticsFilter — unified filters for all analytics endpoints.
type AnalyticsFilter struct {
	OrgID     int
	BotID     *int
	POSID     *int
	SegmentID *int
	From      time.Time
	To        time.Time
}

// --- Sales ---

type SalesMetrics struct {
	TransactionCount int64   `json:"transaction_count" db:"transaction_count"`
	UniqueClients    int64   `json:"unique_clients"    db:"unique_clients"`
	TotalAmount      float64 `json:"total_amount"      db:"total_amount"`
	AvgAmount        float64 `json:"avg_amount"        db:"avg_amount"`
	BuyFrequency     float64 `json:"buy_frequency"     db:"buy_frequency"`
}

type SalesChartPoint struct {
	Day   time.Time `json:"day"   db:"day"`
	Value float64   `json:"value" db:"value"`
}

type SalesAnalytics struct {
	Metrics    SalesMetrics                 `json:"metrics"`
	Charts     map[string][]SalesChartPoint `json:"charts"` // "transactions","revenue","avg_amount"
	Comparison *LoyaltyComparison           `json:"comparison,omitempty"`
}

type LoyaltyComparison struct {
	ParticipantsAvgAmount    float64 `json:"participants_avg_amount"`
	NonParticipantsAvgAmount float64 `json:"non_participants_avg_amount"`
}

// --- Loyalty ---

type LoyaltyAnalytics struct {
	NewClients    int64              `json:"new_clients"`
	ActiveClients int64              `json:"active_clients"`
	BonusEarned   float64            `json:"bonus_earned"`
	BonusSpent    float64            `json:"bonus_spent"`
	Demographics  ClientDemographics `json:"demographics"`
	BotFunnel     []FunnelStep       `json:"bot_funnel"`
}

type ClientDemographics struct {
	ByGender       []PieSlice `json:"by_gender"`
	ByAgeGroup     []PieSlice `json:"by_age_group"`
	ByOS           []PieSlice `json:"by_os"`
	LoyaltyPercent float64    `json:"loyalty_percent"`
}

type PieSlice struct {
	Label   string  `json:"label"`
	Value   int64   `json:"value"`
	Percent float64 `json:"percent"`
}

type FunnelStep struct {
	Step    string  `json:"step"`
	Count   int64   `json:"count"`
	Percent float64 `json:"percent"`
}

// --- Campaigns ---

type CampaignAnalytics struct {
	TotalSent   int64          `json:"total_sent"`
	TotalOpened int64          `json:"total_opened"`
	OpenRate    float64        `json:"open_rate"`
	Conversions int64          `json:"conversions"`
	ConvRate    float64        `json:"conv_rate"`
	ByCampaign  []CampaignStat `json:"by_campaign"`
}

type CampaignStat struct {
	CampaignID   int     `json:"campaign_id"   db:"campaign_id"`
	CampaignName string  `json:"campaign_name" db:"campaign_name"`
	Sent         int64   `json:"sent"          db:"sent"`
	OpenRate     float64 `json:"open_rate"     db:"open_rate"`
	Conversions  int64   `json:"conversions"   db:"conversions"`
}
