package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fox-one/fox-data/models"
	"github.com/fox-one/gin-contrib/session"
)

// WhiteIcon if a asset do not hava a icon in Mixin, it has a white icon
const WhiteIcon = "https://images.mixin.one/yH_I5b0GiV2zDmvrXRyr3bK5xusjfy5q7FX3lw3mM2Ryx4Dfuj6Xcw8SHNRnDKm7ZVE3_LvpKlLdcLrlFQUBhds=s128"

// MatchSlug add slug to table assets
func MatchSlug(s *session.Session) error {
	assets, err := models.AllAssets(s)
	if err != nil {
		return err
	}
	assetsBySymbol := make(map[string][]*models.Asset)
	for _, asset := range assets {
		assetsBySymbol[asset.Symbol] = append(assetsBySymbol[asset.Symbol], asset)
	}

	coins, err := GetCoinInfo(s.Context())
	if err != nil {
		return err
	}
	conisBySymbol := make(map[string][]*Coin)
	for _, coin := range coins {
		conisBySymbol[coin.Symbol] = append(conisBySymbol[coin.Symbol], coin)
	}

	for _, asset := range assets {
		coins, ok := conisBySymbol[asset.Symbol]
		if !ok {
			continue
		}

		var (
			index int
			max   float64
		)

		// 1. symbol唯一，取图标非空的
		// 2. 不唯一，取name匹配度大于90的
		matched := false
		if len(coins) == 1 {
			if len(assetsBySymbol[asset.Symbol]) == 1 || asset.IconURL != WhiteIcon {
				matched = true
			}
		} else {
			for i, coin := range coins {
				var percent float64
				SimilarText(asset.Name, coin.Name, &percent)
				if percent > max {
					max = percent
					index = i
				}
			}

			if max > 90 {
				matched = true
			}
		}

		if matched {
			if err := s.MysqlWrite().Model(&models.Asset{}).Where("asset_id = ?", asset.AssetID).Update("cmc_slug", coins[index].Slug).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// SimilarText similar_text()
func SimilarText(first, second string, percent *float64) int {
	var similarText func(string, string, int, int) int
	similarText = func(str1, str2 string, len1, len2 int) int {
		var sum, max int
		pos1, pos2 := 0, 0

		// Find the longest segment of the same section in two strings
		for i := 0; i < len1; i++ {
			for j := 0; j < len2; j++ {
				for l := 0; (i+l < len1) && (j+l < len2) && (str1[i+l] == str2[j+l]); l++ {
					if l+1 > max {
						max = l + 1
						pos1 = i
						pos2 = j
					}
				}
			}
		}

		if sum = max; sum > 0 {
			if pos1 > 0 && pos2 > 0 {
				sum += similarText(str1, str2, pos1, pos2)
			}
			if (pos1+max < len1) && (pos2+max < len2) {
				s1 := []byte(str1)
				s2 := []byte(str2)
				sum += similarText(string(s1[pos1+max:]), string(s2[pos2+max:]), len1-pos1-max, len2-pos2-max)
			}
		}

		return sum
	}

	l1, l2 := len(first), len(second)
	if l1+l2 == 0 {
		return 0
	}
	sim := similarText(first, second, l1, l2)
	if percent != nil {
		*percent = float64(sim*200) / float64(l1+l2)
	}
	return sim
}

var httpClient = http.Client{Timeout: 10 * time.Second}

// Coin the cmc coin struct
type Coin struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	ID     int    `json:"id"`
	Rank   int    `json:"rank"`
}

// GetCoinInfo get coinmarketcap's coin defination
func GetCoinInfo(ctx context.Context) ([]*Coin, error) {
	url := "https://s2.coinmarketcap.com/generated/search/quick_search.json"
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

	var coins []*Coin
	err = json.Unmarshal(bt, &coins)
	return coins, err
}
