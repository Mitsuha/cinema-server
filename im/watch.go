package im

import (
    "encoding/json"
    "hourglass-socket/distribution"
    "hourglass-socket/model"
    "log"
)

func (i *Im) registerUser(message *distribution.Message) {
    user := model.User{}
    
    if err := json.Unmarshal(message.Origin, &user); err != nil {
        log.Fatalln(err)
        return
    }
    
    if u, ok := i.Users[user.ID]; ok {
        u.Association(message.Conn)
    } else {
        user.Association(message.Conn)
        i.Users[user.ID] = &user
    }
}

func (i *Im) disconnect(msg *distribution.Message) {
    if msg.User != nil && msg.User.Room != nil {
        i.leaveRoom(msg)
    }
}

func (i *Im) createRoom(msg *distribution.Message) {
    if msg.User.Room != nil {
        i.leaveRoom(msg)
    }
    
    room := i.NewRoom(msg.User)
    room.Playlist = msg.Origin
    
    i.Rooms[room.ID], msg.User.Room = room, room
    
    err := i.Reply(msg, room)
    if err != nil {
        log.Fatalln(err)
    }
}

func (i *Im) NewRoom(master *model.User) *model.Room {
    i.roomCreating.Lock()
    defer i.roomCreating.Unlock()
    
    var id = 21066
    //for true {
    //	id = rand.Intn(99999) + 10000
    //	if _, ok := i.Rooms[id]; !ok {
    //		break
    //	}
    //}
    
    return &model.Room{ID: id, Master: master, Users: []*model.User{master}, Speed: 1}
}

func (i *Im) leaveRoom(msg *distribution.Message) {
    if msg.User.Room.Master.ID == msg.User.ID {
        delete(i.Rooms, msg.User.Room.ID)
        i.BroadcastToRoom(msg.User.Room, "dismiss", msg.User.Room)
        msg.User.Room.Dismiss()
        return
    } else {
        i.BroadcastToRoom(msg.User.Room, "leaveRoom", msg.User)
    }
    
    msg.User.Room.RemoveUser(msg.User)
    
    msg.User.Room = nil
}

func (i *Im) joinRoom(msg *distribution.Message) {
    room := model.Room{}
    err := json.Unmarshal(msg.Origin, &room)
    if err != nil {
        _ = i.Reply(msg, &Response{Message: "无法解析的请求，也许是因为软件要更新了"})
        return
    }
    if room, ok := i.Rooms[room.ID]; ok {
        room.AddUser(msg.User)
        _ = i.Reply(msg, &Response{Success: true, Message: "加入成功"})
        i.BroadcastToRoom(room, "joinRoom", msg.User)
        return
    }
    _ = i.Reply(msg, &Response{Message: "房间不存在"})
}

func (i *Im) roomInfo(msg *distribution.Message) {
    var room model.Room
    if err := json.Unmarshal(msg.Origin, &room); err != nil {
        log.Println(err)
        return
    }
    if room, ok := i.Rooms[room.ID]; ok {
        if err := i.Reply(msg, room); err != nil {
            log.Println(err)
        }
        return
    }
    _ = i.Reply(msg, Response{Message: "房间不存在"})
}

func (i *Im) syncPlayList(msg *distribution.Message) {
    if msg.Origin == nil {
        _ = i.Send(msg.Conn, "syncPlayList", msg.Origin)
    } else {
        msg.User.Room.Playlist = msg.Origin
        i.BroadcastToRoom(msg.User.Room, "syncPlayList", msg.Origin)
    }
}

func (i *Im) syncEpisode(msg *distribution.Message) {
    var data = struct {
        Index int `json:"index"`
    }{}
    if err := json.Unmarshal(msg.Origin, &data); err != nil {
        return
    }
    
    msg.User.Room.Episode = data.Index
    i.BroadcastToRoom(msg.User.Room, "syncEpisode", data)
}

func (i *Im) syncDuration(msg *distribution.Message) {
    var data = struct {
        Duration int `json:"duration"`
        Time     int `json:"time"`
    }{}
    
    if err := json.Unmarshal(msg.Origin, &data); err != nil {
        return
    }
    
    msg.User.Room.Duration, msg.User.Room.SyncTime = data.Duration, data.Time
    
    i.BroadcastToRoom(msg.User.Room, "syncDuration", data)
}

func (i *Im) syncPlayingStatus(msg *distribution.Message) {
    var data = struct {
        Playing bool `json:"playing"`
    }{}
    
    if err := json.Unmarshal(msg.Origin, &data); err != nil {
        return
    }
    
    msg.User.Room.IsPlaying = data.Playing
    
    i.BroadcastToRoom(msg.User.Room, "syncPlayingStatus", data)
}
