package auth_client

import "context"

type Mock struct {
	UserID     string
	Err        error
	GetUserIDC int
}

func (m *Mock) GetUserID(_ context.Context) (string, error) {
	m.GetUserIDC++
	if m.Err != nil {
		return "", m.Err
	}
	return m.UserID, nil
}
