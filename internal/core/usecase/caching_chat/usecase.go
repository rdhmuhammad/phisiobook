package caching_chat

import (
	"base-be-golang/internal/core/domain"
	"base-be-golang/pkg/chat_io"
	"base-be-golang/pkg/dto"
	"base-be-golang/pkg/mongodb"
	"context"
	"log"
)

type Usecase struct {
	chatRepo *mongodb.BaseRepo[domain.CacheChat]
}

func NewUsecase(mongoDb mongodb.Conn) Usecase {
	return Usecase{
		chatRepo: mongodb.NewBaseRepo(mongoDb, domain.CacheChat{}),
	}
}

func (u Usecase) GetCached(ctx context.Context, request GetCachedRequest) (dto.PaginationResponse[CachedChat], error) {
	cacheChats, err := u.chatRepo.FindAllByFilterPaged(
		ctx,
		map[string]any{"room_id": request.RoomId, "actor_id": request.ActorId},
		mongodb.PaginationQuery{Page: int64(request.Filter.Page), Limit: int64(request.Filter.PerPage)})
	if err != nil {
		return dto.PaginationResponse[CachedChat]{}, err
	}

	var result = make([]CachedChat, len(cacheChats.Data))
	for i, cacheChat := range cacheChats.Data {
		result[i] = CachedChat{
			ToID:    cacheChat.ToID,
			Message: cacheChat.Message,
			RoomID:  cacheChat.RoomID,
		}
	}

	return dto.NewPagination(result, int(cacheChats.Total), int(cacheChats.Limit), int(cacheChats.Page)), nil
}

func (u Usecase) Store(ctx context.Context, cancel context.CancelFunc, payload chat_io.Transporter) {
	defer cancel()
	var data = domain.CacheChat{
		ToID:    payload.ToID,
		Message: payload.Message,
		RoomID:  payload.RoomID,
	}
	_, err := u.chatRepo.Store(ctx, data)
	if err != nil {
		log.Println(err)
		return
	}
}
