package models

// Response shapes for the master dashboard (GET /api/master/dashboard and /dashboard/trend).
// Dates are Malaysian-local calendar days formatted "YYYY-MM-DD". Slices are always non-nil so
// they serialise as [] rather than null; values that can be undefined (no data to divide by,
// nothing resolved yet) are pointers and serialise as null.

// NotRecordedBrand is the label for warranties/claims with no brand captured. Brand capture only
// started recently, so most historic rows fall in this bucket.
const NotRecordedBrand = "Not recorded"

type DashboardDateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type DashboardCount struct {
	Current  int  `json:"current"`
	Previous *int `json:"previous"` // null when there is no comparison period ("All time")
}

type DashboardMeasure struct {
	Current  *float64 `json:"current"`
	Previous *float64 `json:"previous"`
}

type BrandCoverageCount struct {
	Recorded int `json:"recorded"`
	Total    int `json:"total"`
}

type BrandCoverage struct {
	RecordedSince *string            `json:"recorded_since"` // first warranty with a brand, all time
	Current       BrandCoverageCount `json:"current"`
	Previous      BrandCoverageCount `json:"previous"`
}

// ClaimOutcomes is where the claims filed in the current period stand now. Sums to KPIs.Claims.Current.
type ClaimOutcomes struct {
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	Open     int `json:"open"` // unacknowledged or pending
}

type DashboardKPIs struct {
	Warranties             DashboardCount   `json:"warranties"`
	Claims                 DashboardCount   `json:"claims"`
	ClaimOutcomes          ClaimOutcomes    `json:"claim_outcomes"`
	MedianDaysToResolution DashboardMeasure `json:"median_days_to_resolution"`
}

type DashboardBrandRow struct {
	Brand              string   `json:"brand"`
	Warranties         int      `json:"warranties"`
	PreviousWarranties int      `json:"previous_warranties"`
	Share              float64  `json:"share"` // of warranties with a brand recorded, 0–1
	Claims             int      `json:"claims"`
	ClaimsPer100       *float64 `json:"claims_per_100"`
	Approved           int      `json:"approved"`
	Rejected           int      `json:"rejected"`
	TyresReplaced      int      `json:"tyres_replaced"`
	ApprovalRate       *float64 `json:"approval_rate"`
}

type BrandMixPoint struct {
	Start  string         `json:"start"`
	Counts map[string]int `json:"counts"` // keyed by series name
}

type BrandMix struct {
	Bucket string          `json:"bucket"` // "week" | "month"
	Series []string        `json:"series"` // top brands, then "Other", then "Not recorded"
	Points []BrandMixPoint `json:"points"`
}

type RejectionReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type DashboardResponse struct {
	Range            DashboardDateRange     `json:"range"`
	PreviousRange    *DashboardDateRange    `json:"previous_range"` // null when there is no comparison period
	BrandCoverage    BrandCoverage          `json:"brand_coverage"`
	NotRecorded      *DashboardBrandRow     `json:"not_recorded"`
	KPIs             DashboardKPIs          `json:"kpis"`
	Brands           []DashboardBrandRow    `json:"brands"`
	BrandMix         BrandMix               `json:"brand_mix"`
	RejectionReasons []RejectionReasonCount `json:"rejection_reasons"`
}

type TrendPoint struct {
	Start      string `json:"start"`
	Warranties int    `json:"warranties"`
	Claims     int    `json:"claims"`
}

type TrendResponse struct {
	Granularity string       `json:"granularity"` // "day" | "week" | "month"
	Points      []TrendPoint `json:"points"`
}
