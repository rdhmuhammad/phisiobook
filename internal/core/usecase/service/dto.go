package service

import (
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
)

// CreateServiceRequest represents the request body for creating a service
type CreateServiceRequest struct {
	Name          string   `json:"name" binding:"required"`
	CategoryID    uint     `json:"categoryId" binding:"required"`
	Description   string   `json:"description"`
	Duration      int      `json:"duration" binding:"required,min=1"`
	Price         float64  `json:"price" binding:"required,min=0"`
	Commission    float64  `json:"commission" binding:"required,min=0,max=100"`
	CityIDs       []uint   `json:"cityIds"`
	IncludedItems []string `json:"includedItems"`
}

// UpdateServiceRequest represents the request body for updating a service
type UpdateServiceRequest struct {
	ID            uint
	Name          string   `json:"name" binding:"required"`
	CategoryID    uint     `json:"categoryId" binding:"required"`
	Description   string   `json:"description"`
	Duration      int      `json:"duration" binding:"required,min=1"`
	Price         float64  `json:"price" binding:"required,min=0"`
	Commission    float64  `json:"commission" binding:"required,min=0,max=100"`
	CityIDs       []uint   `json:"cityIds"`
	IncludedItems []string `json:"includedItems"`
}

// ServiceListItem represents a single service in the list response
type ServiceListItem struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Duration      int     `json:"duration"`
	Price         float64 `json:"price"`
	Commission    float64 `json:"commission"`
	CoverageCount int     `json:"coverageCount"`
}

// ServiceDetailResponse represents the full detail of a service
type ServiceDetailResponse struct {
	ID            uint                          `json:"id"`
	Name          string                        `json:"name"`
	CategoryID    uint                          `json:"categoryId"`
	CategoryName  string                        `json:"categoryName"`
	Description   string                        `json:"description"`
	Duration      int                           `json:"duration"`
	Price         float64                       `json:"price"`
	Commission    float64                       `json:"commission"`
	Areas         []ServiceAreaItem             `json:"areas"`
	IncludedItems []ServiceIncludedItemResponse `json:"includedItems"`
}

// ServiceAreaItem represents an area/city where the service is available
type ServiceAreaItem struct {
	ID       uint   `json:"id"`
	CityID   uint   `json:"cityId"`
	CityName string `json:"cityName"`
}

// ServiceIncludedItemResponse represents an included item in the service
type ServiceIncludedItemResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CategoryItem represents a service category for dropdown
type CategoryItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// MapServiceToListItem converts a domain.Service to ServiceListItem
func MapServiceToListItem(service domain.Service, coverageCount int) ServiceListItem {
	return ServiceListItem{
		ID:            service.ID,
		Name:          service.Name,
		Duration:      service.Duration,
		Price:         service.Price,
		Commission:    service.Commission,
		CoverageCount: coverageCount,
	}
}

// MapServiceToDetail converts a domain.Service to ServiceDetailResponse
func MapServiceToDetail(service domain.Service) ServiceDetailResponse {
	detail := ServiceDetailResponse{
		ID:          service.ID,
		Name:        service.Name,
		CategoryID:  service.CategoryID,
		Description: service.Description,
		Duration:    service.Duration,
		Price:       service.Price,
		Commission:  service.Commission,
	}

	if service.Category != nil {
		detail.CategoryName = service.Category.Name
	}

	detail.Areas = make([]ServiceAreaItem, len(service.Areas))
	for i, area := range service.Areas {
		detail.Areas[i] = ServiceAreaItem{
			ID:     area.ID,
			CityID: area.CityID,
		}
		if area.City != nil {
			detail.Areas[i].CityName = area.City.Name
		}
	}

	detail.IncludedItems = make([]ServiceIncludedItemResponse, len(service.IncludedItems))
	for i, item := range service.IncludedItems {
		detail.IncludedItems[i] = ServiceIncludedItemResponse{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	return detail
}

// MapCategoryToItem converts a domain.MasterServiceCategory to CategoryItem
func MapCategoryToItem(category domain.MasterServiceCategory) CategoryItem {
	return CategoryItem{
		ID:   category.ID,
		Name: category.Name,
	}
}
