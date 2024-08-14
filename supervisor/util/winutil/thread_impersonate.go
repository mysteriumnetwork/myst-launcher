package winutil

import (
	"unsafe"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
)

var kernel32 = windows.MustLoadDLL("kernel32.dll")

// RunAsUserInThread
// runs a code closure (callback) in a thread impersonating a user.
// The user is represented by sessionId.
func RunAsUserInThread(sessionId int, closure func()) bool {

	session := uint32(sessionId)
	var userToken windows.Token
	err := windows.WTSQueryUserToken(session, &userToken)
	if err != nil {
		log.Info().Msgf("err> %v", err)
		return false
	}
	if !userToken.IsElevated() {
		log.Info().Msg("Elevate token")
		userToken = tokenElevate(userToken)
		if userToken == windows.Token(windows.InvalidHandle) {
			return false
		}
	}

	var duplicatedToken windows.Token
	err = windows.DuplicateTokenEx(userToken, 0, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &duplicatedToken)
	if err != nil {
		log.Info().Msgf("DuplicateTokenEx err: %v", err)
		return false
	}

	args := threadArgs{
		token:   duplicatedToken,
		closure: closure,
	}
	createThreadProc := kernel32.MustFindProc("CreateThread")
	r, _, err := createThreadProc.Call(0, 0, windows.NewCallback(threadProc), uintptr(unsafe.Pointer(&args)), 0, 0)
	if r == 0 {
		log.Error().Msgf("CreateThead failed, %v", err)
		return false
	}
	h := windows.Handle(r)
	defer windows.CloseHandle(h)
	windows.WaitForSingleObject(h, 0)

	return true
}

type threadArgs struct {
	token   windows.Token
	closure func()
}

func threadProc(p uintptr) uintptr {
	args := (*threadArgs)(unsafe.Pointer(p))

	err := windows.SetThreadToken(nil, windows.Token(args.token))
	if err != nil {
		log.Info().Msgf("threadProc err: %v", err)
		return p
	}
	args.closure()
	return p
}

func tokenElevate(userToken windows.Token) (t windows.Token) {
	elevatedToken, err := userToken.GetLinkedToken()
	userToken.Close()
	if err != nil {
		log.Printf("Unable to elevate token: %v", err)
		return windows.Token(windows.InvalidHandle)
	}
	if !elevatedToken.IsElevated() {
		elevatedToken.Close()
		log.Info().Msg("Linked token is not elevated")
		return windows.Token(windows.InvalidHandle)
	}
	return elevatedToken
}
