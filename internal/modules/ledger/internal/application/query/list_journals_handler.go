package query

import "context"

type JournalReader interface {
	ListJournals(ctx context.Context, merchantID, currency string, limit int) ([]JournalDTO, error)
}

type ListJournalsHandler struct {
	reader JournalReader
}

func NewListJournalsHandler(reader JournalReader) *ListJournalsHandler {
	return &ListJournalsHandler{reader: reader}
}

func (h *ListJournalsHandler) Handle(ctx context.Context, q ListJournals) ([]JournalDTO, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return h.reader.ListJournals(ctx, q.MerchantID, q.Currency, limit)
}
