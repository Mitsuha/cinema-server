package im

func (i *Im) BroadcastToRoom(room *Room, event string, message interface{}) []error {
	errs := make([]error, 0)
	for _, user := range room.Users {
		if user.Conn != nil {
			err := i.distribution.Send(user.Conn, event, message)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
