package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueryItem struct {
	Type            string   `json:"type"`
	Id              string   `json:"id"`
	DmOnly          bool     `json:"dmOnly"`
	Url             *string  `json:"url,omitempty"`
	Name            *string  `json:"name,omitempty"`
	DisplayUrl      *string  `json:"displayUrl,omitempty"`
	MaxPrice        *float64 `json:"maxPrice,omitempty"`
	MinPrice        *float64 `json:"minPrice,omitempty"`
	MinFloat        *float64 `json:"minFloat,omitempty"`
	MaxFloat        *float64 `json:"maxFloat,omitempty"`
	RequiredPhrases *string  `json:"requiredPhrases,omitempty"`
	ExcludePhrases  *string  `json:"excludePhrases,omitempty"`
	ScanMode        *string  `json:"scanMode,omitempty"`
}

func (db *DB) ListSavedQueries(ctx context.Context, queryType string) ([]QueryItem, error) {
	var list []QueryItem

	// CashConverters
	if queryType == "" || queryType == "cashConverters" {
		type Result struct {
			CashConverters
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("CashConverters").
			Select("CashConverters.*, Query.dmOnly").
			Joins("JOIN Query ON CashConverters.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				urlVal := r.Url
				reqPhrases := r.RequiredPhrases
				exclPhrases := r.ExcludePhrases
				scanMode := r.ScanMode
				list = append(list, QueryItem{
					Type:            "cashConverters",
					Id:              r.Url,
					DmOnly:          r.DmOnly,
					Url:             &urlVal,
					RequiredPhrases: &reqPhrases,
					ExcludePhrases:  &exclPhrases,
					ScanMode:        &scanMode,
					MaxPrice:        r.MaxPrice,
				})
			}
		}
	}

	// Ebay
	if queryType == "" || queryType == "ebay" {
		type Result struct {
			Ebay
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("Ebay").
			Select("Ebay.*, Query.dmOnly").
			Joins("JOIN Query ON Ebay.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				urlVal := r.Url
				maxPrice := r.MaxPrice
				list = append(list, QueryItem{
					Type:     "ebay",
					Id:       r.Url,
					DmOnly:   r.DmOnly,
					Url:      &urlVal,
					MaxPrice: &maxPrice,
				})
			}
		}
	}

	// Gumtree
	if queryType == "" || queryType == "gumtree" {
		type Result struct {
			Gumtree
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("Gumtree").
			Select("Gumtree.*, Query.dmOnly").
			Joins("JOIN Query ON Gumtree.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				urlVal := r.Url
				maxPrice := r.MaxPrice
				list = append(list, QueryItem{
					Type:     "gumtree",
					Id:       r.Url,
					DmOnly:   r.DmOnly,
					Url:      &urlVal,
					MaxPrice: &maxPrice,
				})
			}
		}
	}

	// Salvos
	if queryType == "" || queryType == "salvos" {
		type Result struct {
			Salvos
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("Salvos").
			Select("Salvos.*, Query.dmOnly").
			Joins("JOIN Query ON Salvos.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				nameVal := r.Name
				minPrice := r.MinPrice
				maxPrice := r.MaxPrice
				list = append(list, QueryItem{
					Type:     "salvos",
					Id:       r.Name,
					DmOnly:   r.DmOnly,
					Name:     &nameVal,
					MinPrice: &minPrice,
					MaxPrice: &maxPrice,
				})
			}
		}
	}

	// CsMarket
	if queryType == "" || queryType == "csMarket" {
		type Result struct {
			CsMarket
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("CsMarket").
			Select("CsMarket.*, Query.dmOnly").
			Joins("JOIN Query ON CsMarket.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				urlVal := r.Url
				displayUrl := r.DisplayUrl
				maxPrice := r.MaxPrice
				maxFloat := r.MaxFloat
				list = append(list, QueryItem{
					Type:       "csMarket",
					Id:         r.Url,
					DmOnly:     r.DmOnly,
					Url:        &urlVal,
					DisplayUrl: &displayUrl,
					MaxPrice:   &maxPrice,
					MaxFloat:   &maxFloat,
				})
			}
		}
	}

	// SteamMarket
	if queryType == "" || queryType == "steamMarket" {
		type Result struct {
			SteamMarket
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("SteamMarket").
			Select("SteamMarket.*, Query.dmOnly").
			Joins("JOIN Query ON SteamMarket.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				nameVal := r.Name
				displayUrl := r.DisplayUrl
				maxPrice := r.MaxPrice
				list = append(list, QueryItem{
					Type:       "steamMarket",
					Id:         r.Name,
					DmOnly:     r.DmOnly,
					Name:       &nameVal,
					DisplayUrl: &displayUrl,
					MaxPrice:   &maxPrice,
				})
			}
		}
	}

	// CsTradeBot
	if queryType == "" || queryType == "csTradeBot" {
		type Result struct {
			CsTradeBot
			DmOnly bool `gorm:"column:dmOnly"`
		}
		var results []Result
		err := db.WithContext(ctx).Table("CsTradeBot").
			Select("CsTradeBot.*, Query.dmOnly").
			Joins("JOIN Query ON CsTradeBot.queryId = Query.id").
			Find(&results).Error
		if err == nil {
			for _, r := range results {
				nameVal := r.Name
				maxPrice := r.MaxPrice
				minFloat := r.MinFloat
				maxFloat := r.MaxFloat
				list = append(list, QueryItem{
					Type:     "csTradeBot",
					Id:       r.Name,
					DmOnly:   r.DmOnly,
					Name:     &nameVal,
					MaxPrice: &maxPrice,
					MinFloat: &minFloat,
					MaxFloat: &maxFloat,
				})
			}
		}
	}

	return list, nil
}

func (db *DB) CreateSavedQuery(ctx context.Context, queryType string, dmOnly bool, payload map[string]interface{}) (QueryItem, error) {
	queryId := uuid.New().String()
	q := Query{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&q).Error; err != nil {
			return err
		}

		switch queryType {
		case "cashConverters":
			url, _ := payload["url"].(string)
			reqPhrases, _ := payload["requiredPhrases"].(string)
			exclPhrases, _ := payload["excludePhrases"].(string)
			scanMode, _ := payload["scanMode"].(string)
			if scanMode == "" {
				scanMode = "searchUrl"
			}
			var maxPriceVal *float64
			if mp, ok := payload["maxPrice"]; ok && mp != nil {
				if f, ok := mp.(float64); ok {
					maxPriceVal = &f
				}
			}
			cc := CashConverters{
				Url:             url,
				RequiredPhrases: reqPhrases,
				ExcludePhrases:  exclPhrases,
				MaxPrice:        maxPriceVal,
				ScanMode:        scanMode,
				QueryId:         queryId,
			}
			return tx.Create(&cc).Error

		case "ebay":
			url, _ := payload["url"].(string)
			maxPrice, _ := payload["maxPrice"].(float64)
			eb := Ebay{
				Url:      url,
				MaxPrice: maxPrice,
				QueryId:  queryId,
			}
			return tx.Create(&eb).Error

		case "gumtree":
			url, _ := payload["url"].(string)
			maxPrice, _ := payload["maxPrice"].(float64)
			gt := Gumtree{
				Url:      url,
				MaxPrice: maxPrice,
				QueryId:  queryId,
			}
			return tx.Create(&gt).Error

		case "salvos":
			name, _ := payload["name"].(string)
			minPrice, _ := payload["minPrice"].(float64)
			maxPrice, _ := payload["maxPrice"].(float64)
			sa := Salvos{
				Name:     name,
				MinPrice: minPrice,
				MaxPrice: maxPrice,
				QueryId:  queryId,
			}
			return tx.Create(&sa).Error

		case "csMarket":
			url, _ := payload["url"].(string)
			displayUrl, _ := payload["displayUrl"].(string)
			if displayUrl == "" {
				displayUrl = url
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			maxFloat, _ := payload["maxFloat"].(float64)
			cm := CsMarket{
				Url:        url,
				DisplayUrl: displayUrl,
				MaxPrice:   maxPrice,
				MaxFloat:   maxFloat,
				LastPrice:  0.0,
				QueryId:    queryId,
			}
			return tx.Create(&cm).Error

		case "steamMarket":
			name, _ := payload["name"].(string)
			displayUrl, _ := payload["displayUrl"].(string)
			if displayUrl == "" {
				displayUrl = name
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			sm := SteamMarket{
				Name:       name,
				DisplayUrl: displayUrl,
				MaxPrice:   maxPrice,
				LastPrice:  0.0,
				QueryId:    queryId,
			}
			return tx.Create(&sm).Error

		case "csTradeBot":
			name, _ := payload["name"].(string)
			maxPrice, _ := payload["maxPrice"].(float64)
			minFloat, _ := payload["minFloat"].(float64)
			maxFloat, _ := payload["maxFloat"].(float64)
			ct := CsTradeBot{
				Name:     name,
				MaxPrice: maxPrice,
				MinFloat: minFloat,
				MaxFloat: maxFloat,
				QueryId:  queryId,
			}
			return tx.Create(&ct).Error

		default:
			return fmt.Errorf("unknown query type: %s", queryType)
		}
	})

	if err != nil {
		return QueryItem{}, err
	}

	// Fetch query back for response
	queries, err := db.ListSavedQueries(ctx, queryType)
	if err != nil || len(queries) == 0 {
		return QueryItem{}, fmt.Errorf("failed to retrieve created query: %v", err)
	}

	for _, item := range queries {
		if item.Id == queryId || (item.Url != nil && *item.Url == queryId) || (item.Name != nil && *item.Name == queryId) {
			return item, nil
		}
	}
	// Fallback to returning first one (or construct the return item directly)
	return queries[len(queries)-1], nil
}

func (db *DB) UpdateSavedQuery(ctx context.Context, queryType string, id string, dmOnly bool, payload map[string]interface{}) (QueryItem, error) {
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var queryId string
		var err error

		switch queryType {
		case "cashConverters":
			var cc CashConverters
			err = tx.Where("url = ?", id).First(&cc).Error
			queryId = cc.QueryId
		case "ebay":
			var eb Ebay
			err = tx.Where("url = ?", id).First(&eb).Error
			queryId = eb.QueryId
		case "gumtree":
			var gt Gumtree
			err = tx.Where("url = ?", id).First(&gt).Error
			queryId = gt.QueryId
		case "salvos":
			var sa Salvos
			err = tx.Where("name = ?", id).First(&sa).Error
			queryId = sa.QueryId
		case "csMarket":
			var cm CsMarket
			err = tx.Where("url = ?", id).First(&cm).Error
			queryId = cm.QueryId
		case "steamMarket":
			var sm SteamMarket
			err = tx.Where("name = ?", id).First(&sm).Error
			queryId = sm.QueryId
		case "csTradeBot":
			var ct CsTradeBot
			err = tx.Where("name = ?", id).First(&ct).Error
			queryId = ct.QueryId
		}

		if err != nil {
			return err
		}

		// Update parent
		if err := tx.Model(&Query{ID: queryId}).Update("dmOnly", dmOnly).Error; err != nil {
			return err
		}

		switch queryType {
		case "cashConverters":
			url, _ := payload["url"].(string)
			if url == "" {
				url = id
			}
			reqPhrases, _ := payload["requiredPhrases"].(string)
			exclPhrases, _ := payload["excludePhrases"].(string)
			scanMode, _ := payload["scanMode"].(string)
			if scanMode == "" {
				scanMode = "searchUrl"
			}
			var maxPriceVal *float64
			if mp, ok := payload["maxPrice"]; ok && mp != nil {
				if f, ok := mp.(float64); ok {
					maxPriceVal = &f
				}
			}
			return tx.Model(&CashConverters{}).Where("url = ?", id).Updates(map[string]interface{}{
				"url":             url,
				"requiredPhrases": reqPhrases,
				"excludePhrases":  exclPhrases,
				"maxPrice":        maxPriceVal,
				"scanMode":        scanMode,
			}).Error

		case "ebay":
			url, _ := payload["url"].(string)
			if url == "" {
				url = id
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			return tx.Model(&Ebay{}).Where("url = ?", id).Updates(map[string]interface{}{
				"url":      url,
				"maxPrice": maxPrice,
			}).Error

		case "gumtree":
			url, _ := payload["url"].(string)
			if url == "" {
				url = id
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			return tx.Model(&Gumtree{}).Where("url = ?", id).Updates(map[string]interface{}{
				"url":      url,
				"maxPrice": maxPrice,
			}).Error

		case "salvos":
			name, _ := payload["name"].(string)
			if name == "" {
				name = id
			}
			minPrice, _ := payload["minPrice"].(float64)
			maxPrice, _ := payload["maxPrice"].(float64)
			return tx.Model(&Salvos{}).Where("name = ?", id).Updates(map[string]interface{}{
				"name":     name,
				"minPrice": minPrice,
				"maxPrice": maxPrice,
			}).Error

		case "csMarket":
			url, _ := payload["url"].(string)
			if url == "" {
				url = id
			}
			displayUrl, _ := payload["displayUrl"].(string)
			if displayUrl == "" {
				displayUrl = url
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			maxFloat, _ := payload["maxFloat"].(float64)
			return tx.Model(&CsMarket{}).Where("url = ?", id).Updates(map[string]interface{}{
				"url":        url,
				"displayUrl": displayUrl,
				"maxPrice":   maxPrice,
				"maxFloat":   maxFloat,
			}).Error

		case "steamMarket":
			name, _ := payload["name"].(string)
			if name == "" {
				name = id
			}
			displayUrl, _ := payload["displayUrl"].(string)
			if displayUrl == "" {
				displayUrl = name
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			return tx.Model(&SteamMarket{}).Where("name = ?", id).Updates(map[string]interface{}{
				"name":       name,
				"displayUrl": displayUrl,
				"maxPrice":   maxPrice,
			}).Error

		case "csTradeBot":
			name, _ := payload["name"].(string)
			if name == "" {
				name = id
			}
			maxPrice, _ := payload["maxPrice"].(float64)
			minFloat, _ := payload["minFloat"].(float64)
			maxFloat, _ := payload["maxFloat"].(float64)
			return tx.Model(&CsTradeBot{}).Where("name = ?", id).Updates(map[string]interface{}{
				"name":     name,
				"maxPrice": maxPrice,
				"minFloat": minFloat,
				"maxFloat": maxFloat,
			}).Error
		}

		return nil
	})

	if err != nil {
		return QueryItem{}, err
	}

	queries, err := db.ListSavedQueries(ctx, queryType)
	if err != nil || len(queries) == 0 {
		return QueryItem{}, fmt.Errorf("failed to retrieve updated query")
	}

	for _, item := range queries {
		if item.Id == id || (item.Url != nil && *item.Url == id) || (item.Name != nil && *item.Name == id) {
			return item, nil
		}
	}
	return queries[0], nil
}

func (db *DB) DeleteSavedQuery(ctx context.Context, queryType string, id string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var queryId string
		var err error

		switch queryType {
		case "cashConverters":
			var cc CashConverters
			err = tx.Where("url = ?", id).First(&cc).Error
			queryId = cc.QueryId
		case "ebay":
			var eb Ebay
			err = tx.Where("url = ?", id).First(&eb).Error
			queryId = eb.QueryId
		case "gumtree":
			var gt Gumtree
			err = tx.Where("url = ?", id).First(&gt).Error
			queryId = gt.QueryId
		case "salvos":
			var sa Salvos
			err = tx.Where("name = ?", id).First(&sa).Error
			queryId = sa.QueryId
		case "csMarket":
			var cm CsMarket
			err = tx.Where("url = ?", id).First(&cm).Error
			queryId = cm.QueryId
		case "steamMarket":
			var sm SteamMarket
			err = tx.Where("name = ?", id).First(&sm).Error
			queryId = sm.QueryId
		case "csTradeBot":
			var ct CsTradeBot
			err = tx.Where("name = ?", id).First(&ct).Error
			queryId = ct.QueryId
		}

		if err != nil {
			if gorm.ErrRecordNotFound == err {
				return nil // Already deleted
			}
			return err
		}

		// Delete parent Query row. Since DB constraint foreign keys delete cascade, GORM deletes child rows automatically.
		return tx.Delete(&Query{ID: queryId}).Error
	})
}
