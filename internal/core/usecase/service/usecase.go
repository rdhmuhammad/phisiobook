package service

import (
	"context"
	"fmt"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"log"

	"gorm.io/gorm"
)

type Usecase struct {
	base.Port
	dbConn           *gorm.DB
	serviceRepo      db.GenericRepository[domain.Service]
	categoryRepo     db.GenericRepository[domain.MasterServiceCategory]
	serviceAreaRepo  db.GenericRepository[domain.ServiceArea]
	includedItemRepo db.GenericRepository[domain.ServiceIncludedItem]
}

func NewUsecase(dbConn *gorm.DB, prt base.Port) Usecase {
	return Usecase{
		Port:             prt,
		dbConn:           dbConn,
		serviceRepo:      db.NewGenericeRepo(dbConn, domain.Service{}),
		categoryRepo:     db.NewGenericeRepo(dbConn, domain.MasterServiceCategory{}),
		serviceAreaRepo:  db.NewGenericeRepo(dbConn, domain.ServiceArea{}),
		includedItemRepo: db.NewGenericeRepo(dbConn, domain.ServiceIncludedItem{}),
	}
}

// CreateService creates a new service with its areas and included items
func (uc Usecase) CreateService(ctx context.Context, request CreateServiceRequest) (ServiceDetailResponse, error) {
	var userLogin = dto.SessionDataUser{}
	err := uc.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return ServiceDetailResponse{}, err
	}

	// Validate category exists
	_, err = uc.categoryRepo.FindOneByID(ctx, request.CategoryID)
	if err != nil {
		return ServiceDetailResponse{}, localerror.InvalidDataError{Msg: "Category not found"}
	}

	// Start transaction
	tx := uc.dbConn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create service
	service := domain.Service{
		Name:        request.Name,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		Duration:    request.Duration,
		Price:       request.Price,
		Commission:  request.Commission,
	}
	service.SetCreated(userLogin.Email)

	serviceRepo := db.NewGenericeRepo(tx, domain.Service{})
	service, err = serviceRepo.Store(ctx, service)
	if err != nil {
		tx.Rollback()
		return ServiceDetailResponse{}, err
	}

	// Create service areas
	if len(request.CityIDs) > 0 {
		serviceAreaRepo := db.NewGenericeRepo(tx, domain.ServiceArea{})
		for _, cityID := range request.CityIDs {
			area := domain.ServiceArea{
				ServiceID: service.ID,
				CityID:    cityID,
			}
			area.SetCreated(userLogin.Email)
			_, err = serviceAreaRepo.Store(ctx, area)
			if err != nil {
				tx.Rollback()
				return ServiceDetailResponse{}, err
			}
		}
	}

	// Create included items
	if len(request.IncludedItems) > 0 {
		includedItemRepo := db.NewGenericeRepo(tx, domain.ServiceIncludedItem{})
		for _, itemName := range request.IncludedItems {
			item := domain.ServiceIncludedItem{
				ServiceID: service.ID,
				Name:      itemName,
			}
			item.SetCreated(userLogin.Email)
			_, err = includedItemRepo.Store(ctx, item)
			if err != nil {
				tx.Rollback()
				return ServiceDetailResponse{}, err
			}
		}
	}

	if err = tx.Commit().Error; err != nil {
		return ServiceDetailResponse{}, err
	}

	// Fetch the created service with relations
	return uc.GetServiceDetail(ctx, service.ID)
}

// UpdateService updates an existing service with its areas and included items
func (uc Usecase) UpdateService(ctx context.Context, request UpdateServiceRequest) (ServiceDetailResponse, error) {
	var userLogin = dto.SessionDataUser{}
	err := uc.Security.GetSessionLogin(ctx, &userLogin)
	if err != nil {
		return ServiceDetailResponse{}, err
	}

	// Check if service exists
	existingService, err := uc.serviceRepo.FindOneByID(ctx, request.ID)
	if err != nil {
		return ServiceDetailResponse{}, localerror.InvalidDataError{Msg: "Service not found"}
	}

	// Validate category exists
	_, err = uc.categoryRepo.FindOneByID(ctx, request.CategoryID)
	if err != nil {
		return ServiceDetailResponse{}, localerror.InvalidDataError{Msg: "Category not found"}
	}

	// Start transaction
	tx := uc.dbConn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update service
	existingService.Name = request.Name
	existingService.CategoryID = request.CategoryID
	existingService.Description = request.Description
	existingService.Duration = request.Duration
	existingService.Price = request.Price
	existingService.Commission = request.Commission
	existingService.SetUpdated(userLogin.Email)

	serviceRepo := db.NewGenericeRepo(tx, domain.Service{})
	err = serviceRepo.Update(ctx, existingService)
	if err != nil {
		tx.Rollback()
		return ServiceDetailResponse{}, err
	}

	// Delete existing service areas and create new ones
	serviceAreaRepo := db.NewGenericeRepo(tx, domain.ServiceArea{})
	err = serviceAreaRepo.DeleteByExpression(ctx, db.Query(db.Equal(request.ID, "service_id")))
	if err != nil {
		tx.Rollback()
		return ServiceDetailResponse{}, err
	}

	if len(request.CityIDs) > 0 {
		for _, cityID := range request.CityIDs {
			area := domain.ServiceArea{
				ServiceID: request.ID,
				CityID:    cityID,
			}
			area.SetCreated(userLogin.Email)
			_, err = serviceAreaRepo.Store(ctx, area)
			if err != nil {
				tx.Rollback()
				return ServiceDetailResponse{}, err
			}
		}
	}

	// Delete existing included items and create new ones
	includedItemRepo := db.NewGenericeRepo(tx, domain.ServiceIncludedItem{})
	err = includedItemRepo.DeleteByExpression(ctx, db.Query(db.Equal(request.ID, "service_id")))
	if err != nil {
		tx.Rollback()
		return ServiceDetailResponse{}, err
	}

	if len(request.IncludedItems) > 0 {
		for _, itemName := range request.IncludedItems {
			item := domain.ServiceIncludedItem{
				ServiceID: request.ID,
				Name:      itemName,
			}
			item.SetCreated(userLogin.Email)
			_, err = includedItemRepo.Store(ctx, item)
			if err != nil {
				tx.Rollback()
				return ServiceDetailResponse{}, err
			}
		}
	}

	if err = tx.Commit().Error; err != nil {
		return ServiceDetailResponse{}, err
	}

	// Fetch the updated service with relations
	return uc.GetServiceDetail(ctx, request.ID)
}

// DeleteService deletes a service and its related data
func (uc Usecase) DeleteService(ctx context.Context, id uint) error {
	// Check if service exists
	_, err := uc.serviceRepo.FindOneByID(ctx, id)
	if err != nil {
		return localerror.InvalidDataError{Msg: "Service not found"}
	}

	// Delete service (cascade will delete related areas and items)
	return uc.serviceRepo.DeleteByID(ctx, id)
}

// GetServiceDetail returns the full detail of a service
func (uc Usecase) GetServiceDetail(ctx context.Context, id uint) (ServiceDetailResponse, error) {
	service, err := uc.serviceRepo.FindOneByExpressionAndJoin(
		ctx,
		db.Query(db.Equal(id, "services.id")),
		[]string{"Category"},
		[]string{"Areas", "Areas.City", "IncludedItems"},
	)
	if err != nil {
		return ServiceDetailResponse{}, localerror.InvalidDataError{Msg: "Service not found"}
	}

	return MapServiceToDetail(service), nil
}

// GetServiceList returns a paginated list of services
func (uc Usecase) GetServiceList(ctx context.Context, query dto.GetListQueryNoPeriod) (dto.PaginationResponse[ServiceListItem], error) {
	var expressions []interface{}

	// Build search expression if search term is provided
	if query.Search != "" {
		expressions = append(expressions, db.Search(query.Search, "name", "description"))
	}

	// Get paginated services
	services, total, err := uc.serviceRepo.FindPagedByExpressionJoin(
		ctx,
		nil,
		db.PaginationQuery{Page: query.Page, PerPage: query.PerPage},
		nil,
		nil,
		db.ExpressionAnd,
	)
	if err != nil {
		return dto.PaginationResponse[ServiceListItem]{}, err
	}

	// Map to response items with coverage count
	items := make([]ServiceListItem, len(services))
	for i, service := range services {
		coverageCount, err := uc.serviceAreaRepo.CountByExpression(ctx, db.Query(db.Equal(service.ID, "service_id")))
		if err != nil {
			log.Printf("Error counting service areas for service %d: %v", service.ID, err)
			coverageCount = 0
		}
		items[i] = MapServiceToListItem(service, coverageCount)
	}

	return dto.NewPagination(items, total, query.PerPage, query.Page), nil
}

// GetCategories returns all service categories
func (uc Usecase) GetCategories(ctx context.Context) ([]CategoryItem, error) {
	categories, err := uc.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	items := make([]CategoryItem, len(categories))
	for i, category := range categories {
		items[i] = MapCategoryToItem(category)
	}

	return items, nil
}
