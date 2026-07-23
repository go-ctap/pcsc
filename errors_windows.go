//go:build windows

package pcsc

import "errors"

const unsupportedFeatureCode = uint32(0x80100022)

var (
	ErrPINCacheExpired   = errors.New("pcsc: PIN cache expired")
	ErrNoPINCache        = errors.New("pcsc: PIN cache unavailable")
	ErrReadOnlyCard      = errors.New("pcsc: read-only card")
	ErrCacheItemNotFound = errors.New("pcsc: cache item not found")
	ErrCacheItemStale    = errors.New("pcsc: cache item stale")
	ErrCacheItemTooBig   = errors.New("pcsc: cache item too big")
)

var platformErrorsByCode = map[uint32]errorInfo{
	0x80100022: {name: "unsupported feature", err: ErrUnsupportedFeature},
	0x80100032: {name: "PIN cache expired", err: ErrPINCacheExpired},
	0x80100033: {name: "PIN cache unavailable", err: ErrNoPINCache},
	0x80100034: {name: "read-only card", err: ErrReadOnlyCard},
	0x80100070: {name: "cache item not found", err: ErrCacheItemNotFound},
	0x80100071: {name: "cache item stale", err: ErrCacheItemStale},
	0x80100072: {name: "cache item too big", err: ErrCacheItemTooBig},
}
