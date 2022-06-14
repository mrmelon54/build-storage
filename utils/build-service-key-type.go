package utils

type BuildServiceKeyType int

const (
	KeyUser = BuildServiceKeyType(iota)
	KeyState
	KeyAccessToken
	KeyRefreshToken
)
