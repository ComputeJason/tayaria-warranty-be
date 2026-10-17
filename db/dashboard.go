package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
	_ "time/tzdata" // embed tz database so Asia/Kuala_Lumpur loads on any host

	"tayaria-warranty-be/models"

	"github.com/jackc/pgx/v5"
)

// Dashboard aggregates for the master dashboard. Everything is computed on demand from the raw
// warranties/claims tables – volumes are small (thousands of rows), so Postgres aggregates them
// in milliseconds and there is no precomputed/summary table to keep in sync.
//
// Days, weeks and months are Malaysian calendar periods: timestamps are stored as UTC timestamptz,
// so range bounds are computed as local midnights and buckets are truncated in local time.

const dashboardTZ = "Asia/Kuala_Lumpur"

var dashboardLoc = func() *time.Location {
	loc, err := time.LoadLocation(dashboardTZ)
	if err != nil {
		panic(err) // unreachable: time/tzdata is embedded
	}
	return loc
}()

// DashboardPeriod is an inclusive range of local calendar days.
type DashboardPeriod struct {
	From time.Time
	To   time.Time
}

// ParseDashboardDate parses "YYYY-MM-DD" as a Malaysian-local calendar day.
func ParseDashboardDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, dashboardLoc)
}

func (p DashboardPeriod) start() time.Time { return p.From }
func (p DashboardPeriod) end() time.Time   { return p.To.AddDate(0, 0, 1) } // exclusive

func (p DashboardPeriod) toModel() models.DashboardDateRange {
	return models.DashboardDateRange{From: p.From.Format("2006-01-02"), To: p.To.Format("2006-01-02")}
}

// normBrand turns the various "no brand" encodings into NULL: warranties created before brand
// capture were backfilled with 'Unknown' (see migrations/2026-08-09_*), some rows hold an empty string, and
// claims created before the feature have NULL.
func normBrand(col string) string {
	return fmt.Sprintf(`CASE WHEN TRIM(COALESCE(%[1]s, '')) = '' OR LOWER(TRIM(%[1]s)) = 'unknown' THEN NULL ELSE TRIM(%[1]s) END`, col)
}

// resolvedAt is when a claim was approved/rejected. The accept/reject endpoints set date_settled,
// but the generic change-status endpoint doesn't, so fall back to updated_at for those.
const resolvedAt = `CASE WHEN c.status IN ('approved', 'rejected') THEN COALESCE(c.date_settled, c.updated_at) END`

// claimsWithBrand is a CTE of claims with a normalised brand and finalised tyre count. Claims made
// before brand capture have no claims.brand, but if a master accepted/rejected them they entered
// a brand per tyre in tyre_details – use the most common one.
var claimsWithBrand = `
	WITH cl AS (
		SELECT c.status, c.created_at, c.quantity, c.rejection_reason,
		       ` + resolvedAt + ` AS resolved_at,
		       COALESCE(` + normBrand("c.brand") + `, td.brand) AS brand,
		       COALESCE(td.tyres, 0) AS tyres
		FROM claims c
		LEFT JOIN LATERAL (
			SELECT (
				SELECT b FROM (SELECT ` + normBrand("t.brand") + ` AS b FROM tyre_details t WHERE t.claim_id = c.id) x
				WHERE b IS NOT NULL GROUP BY b ORDER BY count(*) DESC, b LIMIT 1
			) AS brand,
			(SELECT count(*) FROM tyre_details t WHERE t.claim_id = c.id) AS tyres
		) td ON true
	)`

func localDate(t time.Time) string { return t.In(dashboardLoc).Format("2006-01-02") }

func ptr[T any](v T) *T { return &v }

type brandAgg struct {
	warranties, prevWarranties                    int
	claims, prevClaims, approved, rejected, tyres int
}

// collect runs a query and scans every row with scan.
func collect[T any](ctx context.Context, query string, args []any, scan func(pgx.Rows) (T, error)) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type warrantyAggRow struct {
	brand     *string
	cur, prev int
}

type claimAggRow struct {
	brand                                    *string
	filed, prevFiled, approved, rejected, ty int
	filedApproved, filedRejected             int // current status of claims filed in the period
}

type mixRow struct {
	bucket string
	brand  *string
	n      int
}

const topBrandsInMix = 5

// mixUnit is the brand-mix bucket size: weekly for ranges up to ~2 months, monthly otherwise.
func mixUnit(p DashboardPeriod) string {
	if p.To.Sub(p.From) <= 62*24*time.Hour {
		return "week"
	}
	return "month"
}

// GetDashboard returns all dashboard figures for the current period, compared with prev. prev is
// nil when there is nothing to compare with ("All time"): previous figures are then null/zero.
// The queries are independent, so they run concurrently – each one is a network round trip to
// the database, which dominates the response time.
func GetDashboard(cur DashboardPeriod, prevPeriod *DashboardPeriod) (*models.DashboardResponse, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Without a comparison period, use an empty window so every "previous" filter matches nothing.
	prevStart, prevEnd := cur.start(), cur.start()
	if prevPeriod != nil {
		prevStart, prevEnd = prevPeriod.start(), prevPeriod.end()
	}
	bounds := []any{cur.start(), cur.end(), prevStart, prevEnd}
	unit := mixUnit(cur)

	var (
		warrantyRows []warrantyAggRow
		claimRows    []claimAggRow
		medianCur    *float64
		medianPrev   *float64
		reasons      []models.RejectionReasonCount
		firstBranded *time.Time
		mixRows      []mixRow
	)
	queries := []struct {
		name string
		run  func() error
	}{
		// 1. Warranties per brand, current + previous period in one pass.
		{"warranties", func() (err error) {
			warrantyRows, err = collect(ctx, `
				SELECT brand,
				       count(*) FILTER (WHERE created_at >= $1 AND created_at < $2),
				       count(*) FILTER (WHERE created_at >= $3 AND created_at < $4)
				FROM (
					SELECT `+normBrand("w.brand")+` AS brand, w.created_at
					FROM warranties w
					WHERE (w.created_at >= $1 AND w.created_at < $2) OR (w.created_at >= $3 AND w.created_at < $4)
				) x
				GROUP BY brand`, bounds, func(r pgx.Rows) (v warrantyAggRow, err error) {
				err = r.Scan(&v.brand, &v.cur, &v.prev)
				return
			})
			return
		}},
		// 2. Claims per brand: filed in each period; outcomes for claims resolved in the current one.
		{"claims", func() (err error) {
			claimRows, err = collect(ctx, claimsWithBrand+`
				SELECT brand,
				       count(*) FILTER (WHERE created_at >= $1 AND created_at < $2),
				       count(*) FILTER (WHERE created_at >= $3 AND created_at < $4),
				       count(*) FILTER (WHERE status = 'approved' AND resolved_at >= $1 AND resolved_at < $2),
				       count(*) FILTER (WHERE status = 'rejected' AND resolved_at >= $1 AND resolved_at < $2),
				       COALESCE(sum(CASE WHEN tyres > 0 THEN tyres ELSE COALESCE(quantity, 0) END)
				                FILTER (WHERE status = 'approved' AND resolved_at >= $1 AND resolved_at < $2), 0)::int,
				       count(*) FILTER (WHERE status = 'approved' AND created_at >= $1 AND created_at < $2),
				       count(*) FILTER (WHERE status = 'rejected' AND created_at >= $1 AND created_at < $2)
				FROM cl
				WHERE (created_at >= $1 AND created_at < $2) OR (created_at >= $3 AND created_at < $4)
				   OR (resolved_at >= $1 AND resolved_at < $2)
				GROUP BY brand`, bounds, func(r pgx.Rows) (v claimAggRow, err error) {
				err = r.Scan(&v.brand, &v.filed, &v.prevFiled, &v.approved, &v.rejected, &v.ty, &v.filedApproved, &v.filedRejected)
				return
			})
			return
		}},
		// 3. Median days from filing to resolution (Malaysian calendar days), per period. Clamped at
		//    0 so bad data (resolved before created) can't produce a negative duration.
		{"resolution time", func() error {
			return db.QueryRow(ctx, `
				SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY days) FILTER (WHERE r >= $1 AND r < $2),
				       percentile_cont(0.5) WITHIN GROUP (ORDER BY days) FILTER (WHERE r >= $3 AND r < $4)
				FROM (
					SELECT `+resolvedAt+` AS r,
					       GREATEST((`+resolvedAt+` AT TIME ZONE '`+dashboardTZ+`')::date
					                - (c.created_at AT TIME ZONE '`+dashboardTZ+`')::date, 0)::float8 AS days
					FROM claims c
					WHERE c.status IN ('approved', 'rejected')
				) x`, bounds...).Scan(&medianCur, &medianPrev)
		}},
		// 4. Rejection reasons for claims rejected in the current period.
		{"rejection reasons", func() (err error) {
			reasons, err = collect(ctx, `
				SELECT COALESCE(NULLIF(TRIM(c.rejection_reason), ''), 'Not specified') AS reason, count(*)
				FROM claims c
				WHERE c.status = 'rejected' AND `+resolvedAt+` >= $1 AND `+resolvedAt+` < $2
				GROUP BY reason
				ORDER BY count(*) DESC, reason`, bounds[:2], func(r pgx.Rows) (v models.RejectionReasonCount, err error) {
				err = r.Scan(&v.Reason, &v.Count)
				return
			})
			return
		}},
		// 5. When brand capture started: the first warranty with a real brand, all time.
		{"brand capture start", func() error {
			return db.QueryRow(ctx, `SELECT min(w.created_at) FROM warranties w WHERE `+normBrand("w.brand")+` IS NOT NULL`).
				Scan(&firstBranded)
		}},
		// 6. Warranties per brand per week/month bucket, for the brand mix chart.
		{"brand mix", func() (err error) {
			mixRows, err = collect(ctx, fmt.Sprintf(`
				SELECT to_char(date_trunc('%s', w.created_at AT TIME ZONE '%s'), 'YYYY-MM-DD') AS bucket,
				       %s AS brand, count(*)
				FROM warranties w
				WHERE w.created_at >= $1 AND w.created_at < $2
				GROUP BY bucket, brand`, unit, dashboardTZ, normBrand("w.brand")), bounds[:2],
				func(r pgx.Rows) (v mixRow, err error) {
					err = r.Scan(&v.bucket, &v.brand, &v.n)
					return
				})
			return
		}},
	}

	errs := make([]error, len(queries))
	var wg sync.WaitGroup
	for i, q := range queries {
		wg.Add(1)
		go func(i int, run func() error) {
			defer wg.Done()
			errs[i] = run()
		}(i, q.run)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("failed to load dashboard %s: %v", queries[i].name, err)
		}
	}

	// brand → aggregates; the "" key holds warranties/claims with no brand recorded.
	aggs := map[string]*brandAgg{}
	agg := func(brand *string) *brandAgg {
		key := ""
		if brand != nil {
			key = *brand
		}
		if aggs[key] == nil {
			aggs[key] = &brandAgg{}
		}
		return aggs[key]
	}
	for _, r := range warrantyRows {
		a := agg(r.brand)
		a.warranties, a.prevWarranties = r.cur, r.prev
	}
	outcomes := models.ClaimOutcomes{}
	for _, r := range claimRows {
		a := agg(r.brand)
		a.claims, a.prevClaims, a.approved, a.rejected, a.tyres = r.filed, r.prevFiled, r.approved, r.rejected, r.ty
		outcomes.Approved += r.filedApproved
		outcomes.Rejected += r.filedRejected
		outcomes.Open += r.filed - r.filedApproved - r.filedRejected
	}

	// Shares are out of warranties that have a brand.
	coverage := models.BrandCoverage{}
	var claimsTotal, prevClaimsTotal int
	for brand, a := range aggs {
		coverage.Current.Total += a.warranties
		coverage.Previous.Total += a.prevWarranties
		if brand != "" {
			coverage.Current.Recorded += a.warranties
			coverage.Previous.Recorded += a.prevWarranties
		}
		claimsTotal += a.claims
		prevClaimsTotal += a.prevClaims
	}
	if firstBranded != nil {
		coverage.RecordedSince = ptr(localDate(*firstBranded))
	}

	toRow := func(brand string, a *brandAgg) models.DashboardBrandRow {
		row := models.DashboardBrandRow{
			Brand:              brand,
			Warranties:         a.warranties,
			PreviousWarranties: a.prevWarranties,
			Claims:             a.claims,
			Approved:           a.approved,
			Rejected:           a.rejected,
			TyresReplaced:      a.tyres,
		}
		if brand != models.NotRecordedBrand && coverage.Current.Recorded > 0 {
			row.Share = float64(a.warranties) / float64(coverage.Current.Recorded)
		}
		if a.warranties > 0 {
			row.ClaimsPer100 = ptr(float64(a.claims) / float64(a.warranties) * 100)
		}
		if decided := a.approved + a.rejected; decided > 0 {
			row.ApprovalRate = ptr(float64(a.approved) / float64(decided))
		}
		return row
	}

	brands := []models.DashboardBrandRow{}
	var notRecorded *models.DashboardBrandRow
	for brand, a := range aggs {
		active := a.warranties+a.prevWarranties+a.claims+a.approved+a.rejected > 0
		if !active {
			continue // e.g. a brand only seen in previous-period claims
		}
		if brand == "" {
			notRecorded = ptr(toRow(models.NotRecordedBrand, a))
			continue
		}
		brands = append(brands, toRow(brand, a))
	}
	sort.Slice(brands, func(i, j int) bool {
		if brands[i].Warranties != brands[j].Warranties {
			return brands[i].Warranties > brands[j].Warranties
		}
		return brands[i].Brand < brands[j].Brand
	})

	var previousRange *models.DashboardDateRange
	warrantiesPrev, claimsPrev := ptr(coverage.Previous.Total), ptr(prevClaimsTotal)
	if prevPeriod != nil {
		previousRange = ptr(prevPeriod.toModel())
	} else {
		warrantiesPrev, claimsPrev = nil, nil
	}

	return &models.DashboardResponse{
		Range:         cur.toModel(),
		PreviousRange: previousRange,
		BrandCoverage: coverage,
		NotRecorded:   notRecorded,
		KPIs: models.DashboardKPIs{
			Warranties:             models.DashboardCount{Current: coverage.Current.Total, Previous: warrantiesPrev},
			Claims:                 models.DashboardCount{Current: claimsTotal, Previous: claimsPrev},
			ClaimOutcomes:          outcomes,
			MedianDaysToResolution: models.DashboardMeasure{Current: medianCur, Previous: medianPrev},
		},
		Brands:           brands,
		BrandMix:         buildBrandMix(cur, unit, brands, notRecorded != nil && notRecorded.Warranties > 0, mixRows),
		RejectionReasons: reasons,
	}, nil
}

// buildBrandMix shapes the per-bucket counts into chart series: the top brands by volume, then
// "Other", then "Not recorded". Every bucket in the range is present, zero-filled.
func buildBrandMix(cur DashboardPeriod, unit string, brands []models.DashboardBrandRow, hasMissing bool, rows []mixRow) models.BrandMix {
	series := []string{}
	top := map[string]bool{}
	withWarranties := 0
	for _, b := range brands {
		if b.Warranties == 0 {
			continue
		}
		withWarranties++
		if len(series) < topBrandsInMix {
			series = append(series, b.Brand)
			top[b.Brand] = true
		}
	}
	if withWarranties > topBrandsInMix {
		series = append(series, "Other")
	}
	if hasMissing {
		series = append(series, models.NotRecordedBrand)
	}

	starts := bucketStarts(cur, unit)
	points := make([]models.BrandMixPoint, len(starts))
	index := map[string]int{}
	for i, s := range starts {
		counts := map[string]int{}
		for _, name := range series {
			counts[name] = 0
		}
		points[i] = models.BrandMixPoint{Start: s, Counts: counts}
		index[s] = i
	}
	for _, r := range rows {
		i, ok := index[r.bucket]
		if !ok {
			continue
		}
		key := models.NotRecordedBrand
		if r.brand != nil {
			key = "Other"
			if top[*r.brand] {
				key = *r.brand
			}
		}
		points[i].Counts[key] += r.n
	}
	return models.BrandMix{Bucket: unit, Series: series, Points: points}
}

// bucketStarts lists the local start date of every day/week/month bucket overlapping the period.
// Weeks start on Monday, matching Postgres date_trunc('week').
func bucketStarts(p DashboardPeriod, unit string) []string {
	s := p.From
	switch unit {
	case "week":
		s = s.AddDate(0, 0, -((int(s.Weekday()) + 6) % 7))
	case "month":
		s = time.Date(s.Year(), s.Month(), 1, 0, 0, 0, 0, dashboardLoc)
	}
	out := []string{}
	for !s.After(p.To) {
		out = append(out, s.Format("2006-01-02"))
		switch unit {
		case "day":
			s = s.AddDate(0, 0, 1)
		case "week":
			s = s.AddDate(0, 0, 7)
		default:
			s = s.AddDate(0, 1, 0)
		}
	}
	return out
}

// CountDashboardBuckets is how many points GetTrend would return – used to reject huge requests.
func CountDashboardBuckets(p DashboardPeriod, unit string) int { return len(bucketStarts(p, unit)) }

// GetTrend returns warranties registered and claims filed per day/week/month, zero-filled.
func GetTrend(p DashboardPeriod, unit string) (*models.TrendResponse, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if unit != "day" && unit != "week" && unit != "month" {
		return nil, fmt.Errorf("invalid granularity %q", unit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	starts := bucketStarts(p, unit)
	points := make([]models.TrendPoint, len(starts))
	index := map[string]int{}
	for i, s := range starts {
		points[i] = models.TrendPoint{Start: s}
		index[s] = i
	}

	// unit is whitelisted above, so formatting it into the query is safe.
	rows, err := db.Query(ctx, fmt.Sprintf(`
		SELECT 'w' AS kind, to_char(date_trunc('%[1]s', w.created_at AT TIME ZONE '%[2]s'), 'YYYY-MM-DD') AS bucket, count(*)
		FROM warranties w WHERE w.created_at >= $1 AND w.created_at < $2 GROUP BY bucket
		UNION ALL
		SELECT 'c', to_char(date_trunc('%[1]s', c.created_at AT TIME ZONE '%[2]s'), 'YYYY-MM-DD'), count(*)
		FROM claims c WHERE c.created_at >= $1 AND c.created_at < $2 GROUP BY 2`, unit, dashboardTZ),
		p.start(), p.end())
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate trend: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var kind, bucket string
		var n int
		if err := rows.Scan(&kind, &bucket, &n); err != nil {
			return nil, fmt.Errorf("failed to scan trend: %v", err)
		}
		i, ok := index[bucket]
		if !ok {
			continue
		}
		if kind == "w" {
			points[i].Warranties = n
		} else {
			points[i].Claims = n
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trend: %v", err)
	}

	return &models.TrendResponse{Granularity: unit, Points: points}, nil
}
