package handshake

import (
	"errors"
)

const (
	HandshakeReqMessageCode byte = iota
	HandshakeResMessageCode
	HandshakeFinMessageCode
)

const (
	HandshakeSuccess byte = iota
	GeneralError
	HandshakeBackwardsCompatible
	HandshakeIncompatible
	HandshakeRequestCompatible
)

var (
	UnexpectedMessageCode = errors.New("Unexpected message code")
	IncompatibleVersions  = errors.New("Incompatible tolliver version")
)
