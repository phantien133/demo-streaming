package streampublishsessions

import (
	"context"

	"demo-streaming/internal/database"
	"gorm.io/gorm"
)

const (
	defaultListPage  = 1
	defaultListLimit = 20
	maxListLimit     = 100
)

type GormListService struct {
	db *gorm.DB
}

func NewGormListService(db *gorm.DB) *GormListService {
	return &GormListService{db: db}
}

func (s *GormListService) Execute(ctx context.Context, input ListInput) (ListOutput, error) {
	page := input.Page
	if page <= 0 {
		page = defaultListPage
	}
	limit := input.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := (page - 1) * limit

	base := s.db.WithContext(ctx).Model(&database.StreamPublishSession{}).
		Where("streamer_user_id = ?", input.StreamerUserID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return ListOutput{}, err
	}

	var rows []database.StreamPublishSession
	if err := s.db.WithContext(ctx).
		Where("streamer_user_id = ?", input.StreamerUserID).
		Preload("Renditions", func(db *gorm.DB) *gorm.DB {
			return db.Order("rendition_name ASC")
		}).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return ListOutput{}, err
	}

	items := make([]ListItem, 0, len(rows))
	for _, row := range rows {
		renditions := make([]ListRenditionItem, 0, len(row.Renditions))
		for _, rendition := range row.Renditions {
			renditions = append(renditions, ListRenditionItem{
				Name:         rendition.RenditionName,
				PlaylistPath: rendition.PlaylistPath,
				Status:       rendition.Status,
			})
		}

		items = append(items, ListItem{
			SessionID:      row.ID,
			PlaybackID:     row.PlaybackID,
			Title:          row.Title,
			Status:         row.Status,
			PlaybackURLCDN: row.PlaybackURLCDN,
			CreatedAt:      row.CreatedAt,
			StartedAt:      row.StartedAt,
			EndedAt:        row.EndedAt,
			Renditions:     renditions,
		})
	}

	return ListOutput{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}
