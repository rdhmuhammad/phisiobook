package booking

import (
	"context"

	"github.com/rdhmuhammad/phisiobook/pkg/db"
)

func (uc Usecase) GetTherapist(ctx context.Context, cityId uint) ([]TherapistDropdownResponse, error) {
	therapist, _, err := uc.therapistRepo.FindPagedByExpressionJoin(
		ctx,
		db.Query(db.Equal(cityId, "city_id")),
		db.PaginationQuery{Page: 1, PerPage: 5},
		[]string{"TherapyType"}, nil,
		db.ExpressionOr,
	)
	if err != nil {
		return nil, uc.ErrHandler.ErrorReturn(err)
	}

	var result = make([]TherapistDropdownResponse, len(therapist))
	for i, t := range therapist {
		result[i] = TherapistDropdownResponse{
			Code:        t.Code,
			Name:        t.Name,
			Field:       t.TherapyType.Name,
			Profile:     t.GetProfile(),
			YearOnField: t.ExperienceYear,
			Price:       uc.FormatRupiah(t.Price),
			Rating:      float32(t.Rating),
		}
	}

	return result, nil
}
