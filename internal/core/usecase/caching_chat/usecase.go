package caching_chat

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rdhmuhammad/phisiobook/internal/constant"
	"github.com/rdhmuhammad/phisiobook/internal/core/domain"
	"github.com/rdhmuhammad/phisiobook/pkg/chat_io"
	"github.com/rdhmuhammad/phisiobook/pkg/db"
	"github.com/rdhmuhammad/phisiobook/pkg/localerror"
	"github.com/rdhmuhammad/phisiobook/pkg/mongodb"
	"github.com/rdhmuhammad/phisiobook/shared/base"
	dto "github.com/rdhmuhammad/phisiobook/shared/payload"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

type Usecase struct {
	base.Port
	expiredListener func(isExpired bool)
	chatRepo        *mongodb.BaseRepo[domain.CacheChat]
	cacheRoomRepo   *mongodb.BaseRepo[domain.RoomSession]
	userRepo        db.GenericRepository[domain.UserExtended]
	userAdminRepo   db.GenericRepository[domain.UserAdminExtended]
}

func NewUsecase(dbConn *gorm.DB, mongoDb *mongodb.Conn, prt base.Port) Usecase {
	return Usecase{
		Port:          prt,
		userRepo:      db.NewGenericeRepo(dbConn, domain.UserExtended{}),
		userAdminRepo: db.NewGenericeRepo(dbConn, domain.UserAdminExtended{}),
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

func (u Usecase) GetChatRoom(ctx context.Context) ([]ChatRoomListResponse, error) {

	var filter = make(map[string]any)
	var ref = chatRef{}
	var err error
	userContext := u.Security.GetUserContext(ctx)
	switch userContext.RoleName {
	case constant.RoleIsUser:
		filter, ref, err = getFilter(ctx, u.userRepo, userContext.UserId, "user_ref")
		if err != nil {
			return nil, u.ErrHandler.ErrorReturn(err)
		}
		break
	case constant.RolesIsTerapis:
		filter, ref, err = getFilter(ctx, u.userAdminRepo, userContext.UserId, "employee_ref")
		if err != nil {
			return nil, u.ErrHandler.ErrorReturn(err)
		}
	}

	roomSessions, err := u.cacheRoomRepo.FindAllByFilter(ctx, filter)
	if err != nil {
		return nil, u.ErrHandler.ErrorReturn(err)
	}

	var result = make([]ChatRoomListResponse, len(roomSessions))
	var roomIds = make([]string, len(roomSessions))
	for i, roomSession := range roomSessions {
		result[i] = ChatRoomListResponse{
			RoomID:         roomSession.BookCode,
			UserRefID:      ref.ChatRef,
			ProfilePicture: ref.Profile,
			Name:           ref.Name,
		}
		roomIds[i] = roomSession.BookCode
	}

	cacheChats, err := u.chatRepo.FindAllByFilter(ctx,
		mongodb.Query(mongodb.In("room_id", roomIds)),
		mongodb.Order(mongodb.DESC("created_at")),
	)
	if err != nil {
		return nil, u.ErrHandler.ErrorReturn(err)
	}
	var chatMaps = make(map[string][]domain.CacheChat)
	for _, cacheChat := range cacheChats {
		chatMaps[cacheChat.RoomID] = append(chatMaps[cacheChat.RoomID], cacheChat)
	}

	for _, dt := range result {
		dt.LatestMessage = chatMaps[dt.RoomID][0].Message
		var unread int
		for _, chat := range chatMaps[dt.RoomID] {
			if !chat.Read {
				unread++
			}
		}
		dt.NotifCount = unread
	}

	return result, nil

}

type chatRef struct {
	ChatRef string `gorm:"column:chat_ref" json:"chatRef"`
	Profile string `gorm:"column:profile" json:"profile"`
	Name    string `gorm:"column:name" json:"name"`
}

func getFilter[T schema.Tabler](
	ctx context.Context,
	repo db.GenericRepository[T],
	authCode string,
	colRef string,
) (map[string]any, chatRef, error) {
	var r chatRef
	err := repo.FindOneByExpSelection(ctx, &r, db.Query(db.Equal(authCode, "auth_code")))
	if err != nil {
		return nil, chatRef{}, err
	}

	return map[string]any{
		colRef: r.ChatRef,
	}, r, nil
}

func (u Usecase) CacheChat(ctx context.Context, request CacheChatRequest) {
	store, err := u.chatRepo.Store(ctx, domain.CacheChat{
		FromID:  request.From,
		ToID:    request.To,
		Message: request.Message,
		Read:    request.Read,
		ReadAt:  request.ReadAt,
		RoomID:  request.RoomID,
	})
	if err != nil {
		u.ErrHandler.ErrorPrint(err)
	}
	u.ErrHandler.DebugPrint("value => %v", store)

	<-ctx.Done()
}

func (u Usecase) SetRoomExpired(ctx context.Context, isExpired bool) {
	defer ctx.Done()
	u.expiredListener(isExpired)
}

func (u Usecase) GetCached(ctx context.Context, request GetCachedRequest) (dto.PaginationResponse[CachedChat], error) {
	userContext := u.Security.GetUserContext(ctx)

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

	go func(tz *time.Location) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		var toUpdate = cacheChats.Data
		for i := range toUpdate {
			toUpdate[i].Read = true
			toUpdate[i].ReadAt = time.Now().In(tz)
		}

		err = u.chatRepo.BulkUpdate(bgCtx, toUpdate)
		if err != nil {
			u.ErrHandler.ErrorPrint(err)
		}
	}(userContext.Tz)

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
