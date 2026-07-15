package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dealscanner/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueryItem struct {
	QueryId           string   `json:"queryId"`
	Type              string   `json:"type"`
	Id                string   `json:"id"`
	DmOnly            bool     `json:"dmOnly"`
	Url               *string  `json:"url,omitempty"`
	Name              *string  `json:"name,omitempty"`
	DisplayUrl        *string  `json:"displayUrl,omitempty"`
	MaxPrice          *float64 `json:"maxPrice,omitempty"`
	LastPrice         *float64 `json:"lastPrice,omitempty"`
	MinPrice          *float64 `json:"minPrice,omitempty"`
	MinFloat          *float64 `json:"minFloat,omitempty"`
	MaxFloat          *float64 `json:"maxFloat,omitempty"`
	RequiredPhrases   *string  `json:"requiredPhrases,omitempty"`
	RequiredMatchMode *string  `json:"requiredMatchMode,omitempty"`
	ExcludePhrases    *string  `json:"excludePhrases,omitempty"`
	ExcludeMatchMode  *string  `json:"excludeMatchMode,omitempty"`
	// Deprecated: compatibility fields for columns that should be merged into
	// RequiredPhrases/ExcludePhrases and removed in a future schema migration.
	RequiredInDescription *string `json:"requiredInDescription,omitempty"`
	ExcludeInDescription  *string `json:"excludeInDescription,omitempty"`
	ScanMode              *string `json:"scanMode,omitempty"`
}

type CashConvertersQueryInput struct {
	Url               string   `json:"url"`
	ScanMode          string   `json:"scanMode"`
	RequiredPhrases   string   `json:"requiredPhrases"`
	RequiredMatchMode string   `json:"requiredMatchMode"`
	ExcludePhrases    string   `json:"excludePhrases"`
	ExcludeMatchMode  string   `json:"excludeMatchMode"`
	MinPrice          *float64 `json:"minPrice"`
	MaxPrice          *float64 `json:"maxPrice"`
}

func (q *Query) ListSavedQueries(ctx context.Context, queryType string) ([]QueryItem, error) {
	if cached, ok := defaultSavedQueryCache.get(queryType); ok {
		return cached, nil
	}

	list := make([]QueryItem, 0)

	// CashConverters
	if queryType == "" || queryType == "cashConverters" {
		filter := q.CashConvertersFilter
		results, err := filter.WithContext(ctx).Preload(filter.Query).Preload(filter.CashConverters).Find()
		if err != nil {
			return nil, fmt.Errorf("list cash converters queries: %w", err)
		}
		for _, r := range results {
			urlVal := r.CashConvertersUrl
			reqPhrases := r.RequiredPhrases
			exclPhrases := r.ExcludePhrases
			requiredMode := r.RequiredMatchMode
			excludeMode := r.ExcludeMatchMode
			scanMode := r.CashConverters.ScanMode
			list = append(list, QueryItem{
				QueryId:           r.QueryId,
				Type:              "cashConverters",
				Id:                r.ID,
				DmOnly:            r.Query.DmOnly,
				Url:               &urlVal,
				RequiredPhrases:   &reqPhrases,
				RequiredMatchMode: &requiredMode,
				ExcludePhrases:    &exclPhrases,
				ExcludeMatchMode:  &excludeMode,
				ScanMode:          &scanMode,
				MinPrice:          r.MinPrice,
				MaxPrice:          r.MaxPrice,
			})
		}
	}

	// Ebay
	if queryType == "" || queryType == "ebay" {
		eb := q.Ebay
		results, err := eb.WithContext(ctx).Preload(eb.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list ebay queries: %w", err)
		}
		for _, r := range results {
			urlVal := r.Url
			maxPrice := r.MaxPrice
			list = append(list, QueryItem{
				QueryId:  r.QueryId,
				Type:     "ebay",
				Id:       r.Url,
				DmOnly:   r.Query.DmOnly,
				Url:      &urlVal,
				MaxPrice: &maxPrice,
			})
		}
	}

	// Gumtree
	if queryType == "" || queryType == "gumtree" {
		gt := q.Gumtree
		results, err := gt.WithContext(ctx).Preload(gt.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list gumtree queries: %w", err)
		}
		for _, r := range results {
			urlVal := r.Url
			maxPrice := r.MaxPrice
			list = append(list, QueryItem{
				QueryId:  r.QueryId,
				Type:     "gumtree",
				Id:       r.Url,
				DmOnly:   r.Query.DmOnly,
				Url:      &urlVal,
				MaxPrice: &maxPrice,
			})
		}
	}

	// Salvos
	if queryType == "" || queryType == "salvos" {
		sa := q.Salvos
		results, err := sa.WithContext(ctx).Preload(sa.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list salvos queries: %w", err)
		}
		for _, r := range results {
			nameVal := r.Name
			minPrice := r.MinPrice
			maxPrice := r.MaxPrice
			list = append(list, QueryItem{
				QueryId:  r.QueryId,
				Type:     "salvos",
				Id:       r.Name,
				DmOnly:   r.Query.DmOnly,
				Name:     &nameVal,
				MinPrice: &minPrice,
				MaxPrice: &maxPrice,
			})
		}
	}

	// CsMarket
	if queryType == "" || queryType == "csMarket" {
		cm := q.CsMarket
		results, err := cm.WithContext(ctx).Preload(cm.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list cs market queries: %w", err)
		}
		for _, r := range results {
			urlVal := r.Url
			displayUrl := r.DisplayUrl
			maxPrice := r.MaxPrice
			maxFloat := r.MaxFloat
			list = append(list, QueryItem{
				QueryId:    r.QueryId,
				Type:       "csMarket",
				Id:         r.Url,
				DmOnly:     r.Query.DmOnly,
				Url:        &urlVal,
				DisplayUrl: &displayUrl,
				MaxPrice:   &maxPrice,
				MaxFloat:   &maxFloat,
			})
		}
	}

	// SteamMarket
	if queryType == "" || queryType == "steamMarket" {
		sm := q.SteamMarket
		results, err := sm.WithContext(ctx).Preload(sm.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list steam market queries: %w", err)
		}
		for _, r := range results {
			nameVal := r.Name
			displayUrl := r.DisplayUrl
			maxPrice := r.MaxPrice
			lastPrice := r.LastPrice
			list = append(list, QueryItem{
				QueryId:    r.QueryId,
				Type:       "steamMarket",
				Id:         r.Name,
				DmOnly:     r.Query.DmOnly,
				Name:       &nameVal,
				DisplayUrl: &displayUrl,
				MaxPrice:   &maxPrice,
				LastPrice:  &lastPrice,
			})
		}
	}

	// CsTradeBot
	if queryType == "" || queryType == "csTradeBot" {
		ct := q.CsTradeBot
		results, err := ct.WithContext(ctx).Preload(ct.Query).Find()
		if err != nil {
			return nil, fmt.Errorf("list cs trade bot queries: %w", err)
		}
		for _, r := range results {
			nameVal := r.Name
			maxPrice := r.MaxPrice
			minFloat := r.MinFloat
			maxFloat := r.MaxFloat
			list = append(list, QueryItem{
				QueryId:  r.QueryId,
				Type:     "csTradeBot",
				Id:       r.Name,
				DmOnly:   r.Query.DmOnly,
				Name:     &nameVal,
				MaxPrice: &maxPrice,
				MinFloat: &minFloat,
				MaxFloat: &maxFloat,
			})
		}
	}

	defaultSavedQueryCache.set(queryType, list)
	return list, nil
}

func (q *Query) fetchQueryItem(ctx context.Context, queryType string, queryId string) (QueryItem, error) {
	queries, err := q.ListSavedQueries(ctx, queryType)
	if err != nil || len(queries) == 0 {
		return QueryItem{}, fmt.Errorf("failed to retrieve query: %w", err)
	}

	for _, item := range queries {
		if item.Id == queryId || (item.Url != nil && *item.Url == queryId) || (item.Name != nil && *item.Name == queryId) {
			return item, nil
		}
	}
	return QueryItem{}, fmt.Errorf("query %q of type %q not found", queryId, queryType)
}

func (q *Query) CreateCashConvertersQuery(ctx context.Context, dmOnly bool, input CashConvertersQueryInput) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}

	err := q.Transaction(func(tx *Query) error {
		scanMode := input.ScanMode
		if scanMode == "" {
			scanMode = "searchUrl"
		}
		if _, err := tx.CashConverters.WithContext(ctx).Where(tx.CashConverters.Url.Eq(input.Url)).
			Attrs(tx.CashConverters.ScanMode.Value(scanMode)).FirstOrCreate(); err != nil {
			return err
		}
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		filter := models.CashConvertersFilter{
			ID: queryId, CashConvertersUrl: input.Url, QueryId: queryId,
			RequiredPhrases:   input.RequiredPhrases,
			RequiredMatchMode: normalizeMatchMode(input.RequiredMatchMode, "all"),
			ExcludePhrases:    input.ExcludePhrases,
			ExcludeMatchMode:  normalizeMatchMode(input.ExcludeMatchMode, "any"),
			MinPrice:          input.MinPrice, MaxPrice: input.MaxPrice, CreatedAt: time.Now().UTC(),
		}
		return tx.CashConvertersFilter.WithContext(ctx).Create(&filter)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "cashConverters", queryId)
}

func normalizeMatchMode(value string, fallback string) string {
	if value == "any" || value == "all" {
		return value
	}
	return fallback
}

func (q *Query) CreateEbayQuery(ctx context.Context, dmOnly bool, eb *models.Ebay) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		eb.QueryId = queryId
		return tx.Ebay.WithContext(ctx).Create(eb)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "ebay", eb.Url)
}

func (q *Query) CreateGumtreeQuery(ctx context.Context, dmOnly bool, gt *models.Gumtree) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		gt.QueryId = queryId
		return tx.Gumtree.WithContext(ctx).Create(gt)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "gumtree", gt.Url)
}

func (q *Query) CreateSalvosQuery(ctx context.Context, dmOnly bool, sa *models.Salvos) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		sa.QueryId = queryId
		return tx.Salvos.WithContext(ctx).Create(sa)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "salvos", sa.Name)
}

func (q *Query) CreateCsMarketQuery(ctx context.Context, dmOnly bool, cm *models.CsMarket) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		cm.QueryId = queryId
		if cm.DisplayUrl == "" {
			cm.DisplayUrl = cm.Url
		}
		return tx.CsMarket.WithContext(ctx).Create(cm)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "csMarket", cm.Url)
}

func (q *Query) CreateSteamMarketQuery(ctx context.Context, dmOnly bool, sm *models.SteamMarket) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		sm.QueryId = queryId
		if sm.DisplayUrl == "" {
			sm.DisplayUrl = sm.Name
		}
		return tx.SteamMarket.WithContext(ctx).Create(sm)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "steamMarket", sm.Name)
}

func (q *Query) CreateCsTradeBotQuery(ctx context.Context, dmOnly bool, ct *models.CsTradeBot) (QueryItem, error) {
	queryId := uuid.New().String()
	qParent := models.SearchQuery{
		ID:        queryId,
		DmOnly:    dmOnly,
		CreatedAt: time.Now().UTC(),
	}
	err := q.Transaction(func(tx *Query) error {
		if err := tx.SearchQuery.WithContext(ctx).Create(&qParent); err != nil {
			return err
		}
		ct.QueryId = queryId
		return tx.CsTradeBot.WithContext(ctx).Create(ct)
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "csTradeBot", ct.Name)
}

func (q *Query) UpdateCashConvertersQuery(ctx context.Context, id string, dmOnly bool, input CashConvertersQueryInput) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		targetURL := input.Url
		if targetURL == "" {
			targetURL = existing.CashConvertersUrl
		}
		scanMode := input.ScanMode
		if scanMode == "" {
			scanMode = "searchUrl"
		}
		if _, err := tx.CashConverters.WithContext(ctx).Where(tx.CashConverters.Url.Eq(targetURL)).
			Attrs(tx.CashConverters.ScanMode.Value(scanMode)).FirstOrCreate(); err != nil {
			return err
		}
		if _, err = tx.CashConverters.WithContext(ctx).Where(tx.CashConverters.Url.Eq(targetURL)).Update(tx.CashConverters.ScanMode, scanMode); err != nil {
			return err
		}
		_, err = tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.ID.Eq(id)).Updates(map[string]interface{}{
			"cashConvertersUrl": targetURL,
			"requiredPhrases":   input.RequiredPhrases,
			"requiredMatchMode": normalizeMatchMode(input.RequiredMatchMode, "all"),
			"excludePhrases":    input.ExcludePhrases,
			"excludeMatchMode":  normalizeMatchMode(input.ExcludeMatchMode, "any"),
			"minPrice":          input.MinPrice,
			"maxPrice":          input.MaxPrice,
		})
		if err != nil {
			return err
		}
		if existing.CashConvertersUrl != targetURL {
			count, countErr := tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.CashConvertersUrl.Eq(existing.CashConvertersUrl)).Count()
			if countErr != nil {
				return countErr
			}
			if count == 0 {
				_, err = tx.CashConverters.WithContext(ctx).Where(tx.CashConverters.Url.Eq(existing.CashConvertersUrl)).Delete()
			}
		}
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "cashConverters", id)
}

func (q *Query) UpdateEbayQuery(ctx context.Context, id string, dmOnly bool, eb *models.Ebay) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.Ebay.WithContext(ctx).Where(tx.Ebay.Url.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if eb.Url == "" {
			eb.Url = id
		}
		_, err = tx.Ebay.WithContext(ctx).Where(tx.Ebay.Url.Eq(id)).Updates(map[string]interface{}{
			"url":      eb.Url,
			"maxPrice": eb.MaxPrice,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "ebay", eb.Url)
}

func (q *Query) UpdateGumtreeQuery(ctx context.Context, id string, dmOnly bool, gt *models.Gumtree) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.Gumtree.WithContext(ctx).Where(tx.Gumtree.Url.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if gt.Url == "" {
			gt.Url = id
		}
		_, err = tx.Gumtree.WithContext(ctx).Where(tx.Gumtree.Url.Eq(id)).Updates(map[string]interface{}{
			"url":      gt.Url,
			"maxPrice": gt.MaxPrice,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "gumtree", gt.Url)
}

func (q *Query) UpdateSalvosQuery(ctx context.Context, id string, dmOnly bool, sa *models.Salvos) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.Salvos.WithContext(ctx).Where(tx.Salvos.Name.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if sa.Name == "" {
			sa.Name = id
		}
		_, err = tx.Salvos.WithContext(ctx).Where(tx.Salvos.Name.Eq(id)).Updates(map[string]interface{}{
			"name":     sa.Name,
			"minPrice": sa.MinPrice,
			"maxPrice": sa.MaxPrice,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "salvos", sa.Name)
}

func (q *Query) UpdateCsMarketQuery(ctx context.Context, id string, dmOnly bool, cm *models.CsMarket) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.CsMarket.WithContext(ctx).Where(tx.CsMarket.Url.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if cm.Url == "" {
			cm.Url = id
		}
		if cm.DisplayUrl == "" {
			cm.DisplayUrl = cm.Url
		}
		_, err = tx.CsMarket.WithContext(ctx).Where(tx.CsMarket.Url.Eq(id)).Updates(map[string]interface{}{
			"url":        cm.Url,
			"displayUrl": cm.DisplayUrl,
			"maxPrice":   cm.MaxPrice,
			"maxFloat":   cm.MaxFloat,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "csMarket", cm.Url)
}

func (q *Query) UpdateSteamMarketQuery(ctx context.Context, id string, dmOnly bool, sm *models.SteamMarket) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.SteamMarket.WithContext(ctx).Where(tx.SteamMarket.Name.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if sm.Name == "" {
			sm.Name = id
		}
		if sm.DisplayUrl == "" {
			sm.DisplayUrl = sm.Name
		}
		_, err = tx.SteamMarket.WithContext(ctx).Where(tx.SteamMarket.Name.Eq(id)).Updates(map[string]interface{}{
			"name":       sm.Name,
			"displayUrl": sm.DisplayUrl,
			"maxPrice":   sm.MaxPrice,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "steamMarket", sm.Name)
}

func (q *Query) UpdateCsTradeBotQuery(ctx context.Context, id string, dmOnly bool, ct *models.CsTradeBot) (QueryItem, error) {
	err := q.Transaction(func(tx *Query) error {
		existing, err := tx.CsTradeBot.WithContext(ctx).Where(tx.CsTradeBot.Name.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(existing.QueryId)).Update(tx.SearchQuery.DmOnly, dmOnly); err != nil {
			return err
		}
		if ct.Name == "" {
			ct.Name = id
		}
		_, err = tx.CsTradeBot.WithContext(ctx).Where(tx.CsTradeBot.Name.Eq(id)).Updates(map[string]interface{}{
			"name":     ct.Name,
			"maxPrice": ct.MaxPrice,
			"minFloat": ct.MinFloat,
			"maxFloat": ct.MaxFloat,
		})
		return err
	})
	if err != nil {
		return QueryItem{}, err
	}
	defaultSavedQueryCache.invalidate()
	return q.fetchQueryItem(ctx, "csTradeBot", ct.Name)
}

func (q *Query) DeleteSavedQuery(ctx context.Context, queryType string, id string) error {
	err := q.Transaction(func(tx *Query) error {
		var queryId string
		var cashConvertersURL string
		var err error

		switch queryType {
		case "cashConverters":
			var filter *models.CashConvertersFilter
			filter, err = tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.ID.Eq(id)).First()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				filters, findErr := tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.CashConvertersUrl.Eq(id)).Find()
				if findErr != nil {
					return findErr
				}
				switch len(filters) {
				case 0:
					return gorm.ErrRecordNotFound
				case 1:
					filter = filters[0]
					err = nil
				default:
					return fmt.Errorf("multiple Cash Converters queries use URL %q; delete by ID instead", id)
				}
			}
			if err == nil {
				queryId = filter.QueryId
				cashConvertersURL = filter.CashConvertersUrl
			}
		case "ebay":
			var eb *models.Ebay
			eb, err = tx.Ebay.WithContext(ctx).Where(tx.Ebay.Url.Eq(id)).First()
			if err == nil {
				queryId = eb.QueryId
			}
		case "gumtree":
			var gt *models.Gumtree
			gt, err = tx.Gumtree.WithContext(ctx).Where(tx.Gumtree.Url.Eq(id)).First()
			if err == nil {
				queryId = gt.QueryId
			}
		case "salvos":
			var sa *models.Salvos
			sa, err = tx.Salvos.WithContext(ctx).Where(tx.Salvos.Name.Eq(id)).First()
			if err == nil {
				queryId = sa.QueryId
			}
		case "csMarket":
			var cm *models.CsMarket
			cm, err = tx.CsMarket.WithContext(ctx).Where(tx.CsMarket.Url.Eq(id)).First()
			if err == nil {
				queryId = cm.QueryId
			}
		case "steamMarket":
			var sm *models.SteamMarket
			sm, err = tx.SteamMarket.WithContext(ctx).Where(tx.SteamMarket.Name.Eq(id)).First()
			if err == nil {
				queryId = sm.QueryId
			}
		case "csTradeBot":
			var ct *models.CsTradeBot
			ct, err = tx.CsTradeBot.WithContext(ctx).Where(tx.CsTradeBot.Name.Eq(id)).First()
			if err == nil {
				queryId = ct.QueryId
			}
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // Already deleted
			}
			return err
		}

		if cashConvertersURL != "" {
			if _, err = tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.ID.Eq(id)).Delete(); err != nil {
				return err
			}
		}
		if _, err = tx.SearchQuery.WithContext(ctx).Where(tx.SearchQuery.ID.Eq(queryId)).Delete(); err != nil {
			return err
		}
		if cashConvertersURL != "" {
			count, countErr := tx.CashConvertersFilter.WithContext(ctx).Where(tx.CashConvertersFilter.CashConvertersUrl.Eq(cashConvertersURL)).Count()
			if countErr != nil {
				return countErr
			}
			if count == 0 {
				_, err = tx.CashConverters.WithContext(ctx).Where(tx.CashConverters.Url.Eq(cashConvertersURL)).Delete()
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	defaultSavedQueryCache.invalidate()
	return nil
}

type FoundItem struct {
	Title       string
	Url         string
	Price       float64
	LastMatched time.Time
}

func (q *Query) GetLastFoundItems(ctx context.Context, queryId string, limit int) ([]FoundItem, error) {
	var results []struct {
		Title                  string
		CanonicalUrl           string
		LastNotifiedTotalPrice *float64
		LowestObservedPrice    *float64
		LastMatchedAt          *time.Time
	}

	err := q.db.WithContext(ctx).
		Table("QueryListingState").
		Select("Listing.title, Listing.canonicalUrl, QueryListingState.lastNotifiedTotalPrice, QueryListingState.lowestObservedPrice, QueryListingState.lastMatchedAt").
		Joins("JOIN Listing ON QueryListingState.listingId = Listing.id").
		Where("QueryListingState.queryId = ? AND QueryListingState.status IN (?)", queryId, []ListingStateStatus{ListingStateStatusNotified, ListingStateStatusMatched}).
		Order("QueryListingState.lastMatchedAt DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var items []FoundItem
	for _, r := range results {
		price := 0.0
		if r.LastNotifiedTotalPrice != nil {
			price = *r.LastNotifiedTotalPrice
		} else if r.LowestObservedPrice != nil {
			price = *r.LowestObservedPrice
		}

		matchedAt := time.Time{}
		if r.LastMatchedAt != nil {
			matchedAt = *r.LastMatchedAt
		}

		items = append(items, FoundItem{
			Title:       r.Title,
			Url:         r.CanonicalUrl,
			Price:       price,
			LastMatched: matchedAt,
		})
	}
	return items, nil
}

func (q *Query) GetLastFoundItemsBatch(ctx context.Context, queryIds []string, limit int) (map[string][]FoundItem, error) {
	if len(queryIds) == 0 {
		return make(map[string][]FoundItem), nil
	}

	var results []struct {
		QueryId                string
		Title                  string
		CanonicalUrl           string
		LastNotifiedTotalPrice *float64
		LowestObservedPrice    *float64
		LastMatchedAt          *time.Time
	}

	subQuery := q.db.Table("QueryListingState").
		Select(`
			QueryListingState.queryId,
			Listing.title,
			Listing.canonicalUrl,
			QueryListingState.lastNotifiedTotalPrice,
			QueryListingState.lowestObservedPrice,
			QueryListingState.lastMatchedAt,
			ROW_NUMBER() OVER (
				PARTITION BY QueryListingState.queryId 
				ORDER BY QueryListingState.lastMatchedAt DESC
			) as rn
		`).
		Joins("JOIN Listing ON QueryListingState.listingId = Listing.id").
		Where("QueryListingState.queryId IN ? AND QueryListingState.status IN ?", queryIds, []ListingStateStatus{ListingStateStatusNotified, ListingStateStatusMatched})

	err := q.db.WithContext(ctx).
		Table("(?) as Ranked", subQuery).
		Select("queryId, title, canonicalUrl, lastNotifiedTotalPrice, lowestObservedPrice, lastMatchedAt").
		Where("rn <= ?", limit).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	itemsMap := make(map[string][]FoundItem)
	for _, qId := range queryIds {
		itemsMap[qId] = []FoundItem{}
	}

	for _, r := range results {
		price := 0.0
		if r.LastNotifiedTotalPrice != nil {
			price = *r.LastNotifiedTotalPrice
		} else if r.LowestObservedPrice != nil {
			price = *r.LowestObservedPrice
		}

		matchedAt := time.Time{}
		if r.LastMatchedAt != nil {
			matchedAt = *r.LastMatchedAt
		}

		item := FoundItem{
			Title:       r.Title,
			Url:         r.CanonicalUrl,
			Price:       price,
			LastMatched: matchedAt,
		}
		itemsMap[r.QueryId] = append(itemsMap[r.QueryId], item)
	}

	return itemsMap, nil
}
