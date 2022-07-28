package im

import (
    "hourglass-socket/distribution"
    "hourglass-socket/model"
)

func AuthMiddleware(msg *distribution.Message) bool {
    var ok bool
    
    msg.User, ok = msg.Conn.Attach.(*model.User)
    
    if !ok {
        message := distribution.Message{
            Event:   "reply",
            Payload: Response{Message: "未登录，试试重启？"},
        }
        data, err := message.JsonEncode()
        if err == nil {
            _ = msg.Conn.Emit(data)
        }
    }
    return ok
}

func HasRoomMiddleware(msg *distribution.Message) bool {
    if msg.User == nil || msg.User.Room == nil {
        message := distribution.Message{
            Event:   "reply",
            Payload: Response{Message: "你不在任何房间里，这应该是个BUG，快去找作者反馈"},
        }
        data, err := message.JsonEncode()
        if err == nil {
            _ = msg.Conn.Emit(data)
        }
        
        return false
    }
    
    return true
}
