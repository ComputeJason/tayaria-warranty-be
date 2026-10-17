package handlers

import (
	"fmt"
	"net/http"

	"tayaria-warranty-be/db"

	"github.com/gin-gonic/gin"
)

// maxTrendBuckets caps how many points one trend request can ask for (e.g. 3 years by day).
const maxTrendBuckets = 400

// parseDashboardPeriod reads a required inclusive date range from two query params.
func parseDashboardPeriod(c *gin.Context, fromKey, toKey string) (db.DashboardPeriod, error) {
	fromStr, toStr := c.Query(fromKey), c.Query(toKey)
	if fromStr == "" || toStr == "" {
		return db.DashboardPeriod{}, fmt.Errorf("'%s' and '%s' are required (YYYY-MM-DD)", fromKey, toKey)
	}
	from, err := db.ParseDashboardDate(fromStr)
	if err != nil {
		return db.DashboardPeriod{}, fmt.Errorf("invalid '%s' date, expected YYYY-MM-DD", fromKey)
	}
	to, err := db.ParseDashboardDate(toStr)
	if err != nil {
		return db.DashboardPeriod{}, fmt.Errorf("invalid '%s' date, expected YYYY-MM-DD", toKey)
	}
	if from.After(to) {
		return db.DashboardPeriod{}, fmt.Errorf("'%s' must not be after '%s'", fromKey, toKey)
	}
	return db.DashboardPeriod{From: from, To: to}, nil
}

// GET /api/master/dashboard?from=YYYY-MM-DD&to=YYYY-MM-DD[&prev_from=YYYY-MM-DD&prev_to=YYYY-MM-DD]
// Dashboard figures for [from, to], compared with [prev_from, prev_to]. The comparison period is
// chosen by the frontend (full previous months for presets, same length for custom ranges) and
// is omitted entirely for "All time", in which case previous figures come back as null.
func GetDashboard(c *gin.Context) {
	cur, err := parseDashboardPeriod(c, "from", "to")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var prev *db.DashboardPeriod
	if c.Query("prev_from") != "" || c.Query("prev_to") != "" {
		p, err := parseDashboardPeriod(c, "prev_from", "prev_to")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		prev = &p
	}

	data, err := db.GetDashboard(cur, prev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to load dashboard: %v", err)})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GET /api/master/dashboard/trend?from=YYYY-MM-DD&to=YYYY-MM-DD&granularity=day|week|month
// Warranties registered and claims filed per bucket, zero-filled across the range.
func GetDashboardTrend(c *gin.Context) {
	period, err := parseDashboardPeriod(c, "from", "to")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	granularity := c.Query("granularity")
	if granularity != "day" && granularity != "week" && granularity != "month" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'granularity' must be one of day, week, month"})
		return
	}
	if n := db.CountDashboardBuckets(period, granularity); n > maxTrendBuckets {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("range too long for %s granularity (%d points, max %d)", granularity, n, maxTrendBuckets)})
		return
	}

	data, err := db.GetTrend(period, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to load trend: %v", err)})
		return
	}
	c.JSON(http.StatusOK, data)
}
