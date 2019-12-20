package server

import (
	"strconv"
	"strings"
	"time"

	"github.com/fox-one/fox-data/models"
	"github.com/fox-one/gin-contrib/gin_helper"
	"github.com/gin-gonic/gin"
)

// RegisterPriceRoute register price router to gin.Engine
func RegisterPriceRoute(r *gin.Engine) {
	imp := &priceImp{}
	r.GET("/data/price/lastest", imp.lastestPrice)
	r.GET("/data/price/history", imp.historyPrice)
}

type priceImp struct{}

func (imp *priceImp) lastestPrice(c *gin.Context) {
	assets, err := models.GetLastestAssetSnapshots(FoxSession(c))
	if err != nil {
		gin_helper.FailError(c, err)
		return
	}
	gin_helper.Data(c, assets)
}

func (imp *priceImp) historyPrice(c *gin.Context) {
	interval := models.SnapshotIntervalDaily
	if c.Query("interval") == "monthly" {
		interval = models.SnapshotIntervalMonthly
	}

	ids := strings.Fields(c.Query("asset_id"))
	from := time.Time{}
	if fr, _ := strconv.ParseInt(c.Query("from"), 10, 64); fr > 0 {
		from = time.Unix(fr, 0)
	}
	to := time.Now()
	if t, _ := strconv.ParseInt(c.Query("to"), 10, 64); t > 0 {
		to = time.Unix(t, 0)
	}
	assets, err := models.QueryAssetSnapshots(FoxSession(c), interval, ids, from, to)
	if err != nil {
		gin_helper.FailError(c, err)
		return
	}
	gin_helper.Data(c, assets)
}
