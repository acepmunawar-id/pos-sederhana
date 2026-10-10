package service

import (
	"context"
	"io"
	"strings"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
)

type SettingService struct {
	repo    domain.StoreSettingRepository
	storage domain.FileStorage
}

func NewSettingService(repo domain.StoreSettingRepository, storage domain.FileStorage) *SettingService {
	return &SettingService{repo: repo, storage: storage}
}

func (s *SettingService) Get(ctx context.Context) (*domain.StoreSetting, error) {
	return s.repo.Get(ctx)
}

func (s *SettingService) Update(ctx context.Context, req dto.UpdateStoreSettingRequest) (*domain.StoreSetting, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	st.Name = strings.TrimSpace(req.Name)
	st.Address, st.Phone, st.ReceiptFooter = trimPtr(req.Address), trimPtr(req.Phone), trimPtr(req.ReceiptFooter)
	st.TaxRateBP = req.TaxRateBP
	if err := s.repo.Update(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *SettingService) UploadLogo(ctx context.Context, contentType string, r io.Reader) (*domain.StoreSetting, error) {
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return nil, apperror.Validation([]apperror.FieldError{{Field: "logo", Message: "format gambar harus JPG, PNG, atau WEBP"}})
	}
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	url, err := s.storage.Save(ctx, "logo", ext, r)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLogo(ctx, &url); err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}
	if st.LogoURL != nil {
		_ = s.storage.Delete(ctx, *st.LogoURL)
	}
	return s.repo.Get(ctx)
}
