package im

import (
    "hourglass-socket/distribution"
    "hourglass-socket/model"
    "sync"
)

type Im struct {
    Distributor  *distribution.Distribution
    Rooms        map[int]*model.Room
    Users        map[string]*model.User
    roomCreating sync.Mutex
}

func Register(distributor *distribution.Distribution) *Im {
    im := &Im{
        Distributor: distributor,
        Rooms:       make(map[int]*model.Room),
        Users:       make(map[string]*model.User),
    }
    im.Register()
    
    return im
}

func (i *Im) Register() {
    i.Distributor.RegisterMany(map[string]*distribution.Listener{
        "disconnect": {Action: i.disconnect},
        "register":   {Action: i.registerUser},
        "roomInfo":   {Action: i.roomInfo},
        "createRoom": {
            Middlewares: []distribution.Middleware{AuthMiddleware},
            Action:      i.createRoom,
        },
        "joinRoom": {
            Middlewares: []distribution.Middleware{AuthMiddleware},
            Action:      i.joinRoom,
        },
        "leaveRoom": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.leaveRoom,
        },
        "syncPlayList": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.syncPlayList,
        },
        "syncEpisode": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.syncEpisode,
        },
        "syncDuration": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.syncDuration,
        },
        "syncPlayingStatus": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.syncPlayingStatus,
        },
        "syncSpeed": {
            Middlewares: []distribution.Middleware{AuthMiddleware, HasRoomMiddleware},
            Action:      i.syncSpeed,
        },
    })
}
