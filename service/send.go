package service

import "hourglass-socket/distribution"

func (s *WatchService) BroadcastToRoom(room *Room, event string, message interface{}) []error {
	errs := make([]error, 0)
	for _, user := range room.Users {
		if user.Conn != nil {
			if err := distribution.Emit(user.Conn, event, message); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (s *WatchService) BroadcastWithoutMaster(room *Room, event string, message interface{}) []error {
	errs := make([]error, 0)
	for _, user := range room.Users {
		if user.Conn != nil && user != room.Master {
			if err := distribution.Emit(user.Conn, event, message); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
