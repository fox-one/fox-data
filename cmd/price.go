package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/fox-one/fox-data/models"
	"github.com/fox-one/gin-contrib/session"
	"github.com/shopspring/decimal"
)

// Price a asset's daily price
type Price struct {
	TimeInMS  int64
	Slug      string
	PriceBTC  decimal.Decimal
	PriceUSD  decimal.Decimal
	MarketCap decimal.Decimal
	VolumeUSD decimal.Decimal
}

// CrawlHistoryPrices crawl all assets' history price
func CrawlHistoryPrices(s *session.Session) error {
	assets, err := models.AllAssets(s)
	if err != nil {
		return err
	}
	for _, asset := range assets {
		if asset.CmcSlug == "" {
			continue
		}
		log.Printf("crawl price for %v", asset.CmcSlug)
		prices, err := GetDailyPrices(s.Context(), asset.CmcSlug)
		if err != nil {
			return err
		}
		for _, price := range prices {
			year, month, day := time.Unix(price.TimeInMS/1000, 0).Date()
			dailyAsset := models.AssetSnapshot{
				AssetID:   asset.AssetID,
				Date:      time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
				Price:     price.PriceUSD,
				MarketCap: price.MarketCap,
				Volume:    price.VolumeUSD,
			}
			dailyTable := dailyAsset.TableName(models.SnapshotIntervalDaily)
			if err := s.MysqlWrite().Table(dailyTable).Where("asset_id = ? AND date = ?", dailyAsset.AssetID, dailyAsset.Date).FirstOrCreate(&dailyAsset).Error; err != nil {
				return err
			}

			monthlyAsset := dailyAsset
			monthlyAsset.ID = 0
			monthlyAsset.Date = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			monthlyTable := monthlyAsset.TableName(models.SnapshotIntervalMonthly)
			if err := s.MysqlWrite().Table(monthlyTable).Where("asset_id = ? AND date = ?", monthlyAsset.AssetID, monthlyAsset.Date).FirstOrCreate(&monthlyAsset).Error; err != nil {
				return err
			}
		}
		// sleep to avoid 429
		time.Sleep(3 * time.Second)
	}
	return nil
}

// GetDailyPrices get a daily price
func GetDailyPrices(ctx context.Context, slug string) ([]*Price, error) {
	// https://graphs2.coinmarketcap.com/currencies/bitcoin/
	url := fmt.Sprintf("https://graphs2.coinmarketcap.com/currencies/%s", slug)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		PriceBTC  [][2]float64 `json:"price_btc"`
		PriceUSD  [][2]float64 `json:"price_usd"`
		VolumeUSD [][2]float64 `json:"volume_usd"`
		MarketCap [][2]float64 `json:"market_cap_by_available_supply"`
	}
	if err := json.Unmarshal(bt, &data); err != nil {
		return nil, err
	}

	var prices []*Price
	for i := range data.PriceBTC {
		price := Price{
			TimeInMS:  int64(data.PriceBTC[i][0]),
			Slug:      slug,
			PriceBTC:  decimal.NewFromFloat(data.PriceBTC[i][1]),
			PriceUSD:  decimal.NewFromFloat(data.PriceUSD[i][1]),
			MarketCap: decimal.NewFromFloat(data.MarketCap[i][1]),
			VolumeUSD: decimal.NewFromFloat(data.VolumeUSD[i][1]),
		}
		prices = append(prices, &price)
	}
	return prices, nil
}
