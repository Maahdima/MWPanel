package common

import "errors"

var (
	ErrPeerNotFound  = errors.New("insufficient balance")
	ErrPeerNotShared = errors.New("peer is not shared")
)
