package caching_chat

import (
	"context"
	"errors"
	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/chat_io"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"

	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

type Usecase struct {
	base.Port
	expiredListener func(isExpired bool)
	chatRepo        *mongodb.BaseRepo[domain.CacheChat]
	cacheRoomRepo   *mongodb.BaseRepo[domain.RoomSession]
}

func NewUsecase(mongoDb *mongodb.Conn, prt base.Port) Usecase {
	return Usecase{
		Port:          prt,
		cacheRoomRepo: mongodb.NewBaseRepo(mongoDb, domain.RoomSession{}),
		chatRepo:      mongodb.NewBaseRepo(mongoDb, domain.CacheChat{}),
	}
}

func (u Usecase) CacheRoom(ctx context.Context, request CacheRoomRequest, cancle context.CancelFunc) (CacheRoomResponse, error) {
	defer cancle()

	room, err := u.cacheRoomRepo.FindOneByFilter(
		ctx,
		map[string]any{
			"book_code": request.RoomID,
		},
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return CacheRoomResponse{}, localerror.InvalidData(constant.RoomChatNotFound)
		}
		return CacheRoomResponse{}, u.ErrHandler.ErrorReturn(err)
	}
	var createNew = room.RoomIsLive == false

	if !room.IsValid {
		return CacheRoomResponse{}, localerror.InvalidData(constant.RoomNotValid)
	}

	switch {
	case room.UserRef == request.Actor:
		room.UserIsLive = request.ActorStatus
		break
	case room.EmployeeRef == request.Actor:
		room.EmployeeIsLive = request.ActorStatus
		break
	default:
		return CacheRoomResponse{}, localerror.AccessNotAllowed(constant.UserNotFound)
	}

	room.RoomIsLive = room.UserIsLive || room.EmployeeIsLive
	_, err = u.cacheRoomRepo.Update(ctx, room)
	if err != nil {
		return CacheRoomResponse{}, u.ErrHandler.ErrorReturn(err)
	}
	return CacheRoomResponse{
		NewRoom:      createNew,
		FromRef:      request.Actor,
		RoomIsLive:   room.RoomIsLive,
		UserFullName: room.GetActorName(request.RoomID),
		ToRef:        room.GetToSocketID(request.Actor),
	}, nil
}

func (u Usecase) CacheChat(ctx context.Context, request CacheChatRequest) {
	store, err := u.chatRepo.Store(ctx, domain.CacheChat{
		FromID:  request.From,
		ToID:    request.To,
		Message: request.Message,
		RoomID:  request.RoomID,
	})
	if err != nil {
		u.ErrHandler.ErrorPrint(err)
	}
	u.ErrHandler.DebugPrint("value => %v", store)

	select {
	case <-ctx.Done():
	}
}

func (u Usecase) SetRoomExpired(ctx context.Context, isExpired bool) {
	defer ctx.Done()
	u.expiredListener(isExpired)
}

func (u Usecase) GetCached(ctx context.Context, request GetCachedRequest) (dto.PaginationResponse[CachedChat], error) {
	cacheChats, err := u.chatRepo.FindAllByFilterPaged(
		ctx,
		map[string]any{"room_id": request.RoomId},
		mongodb.PaginationQuery{Page: int64(request.Filter.Page), Limit: int64(request.Filter.PerPage)})
	if err != nil {
		return dto.PaginationResponse[CachedChat]{}, err
	}

	var result = make([]CachedChat, len(cacheChats.Data))
	for i, cacheChat := range cacheChats.Data {
		result[i] = CachedChat{
			FromID:  cacheChat.FromID,
			ToID:    cacheChat.ToID,
			Message: cacheChat.Message,
			RoomID:  cacheChat.RoomID,
		}
	}

	return dto.NewPagination(result, int(cacheChats.Total), int(cacheChats.Limit), int(cacheChats.Page)), nil
}

// Store is Deprecated
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
