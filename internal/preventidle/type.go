package preventidle

type IdleController struct {
	StopChan chan struct{}
	Running  bool
}

type SimulateAction struct {
	inputType  string
	actionName string
}

type LastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

type KeyboardInput struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type CallKeyboardInput struct {
	Type uint32
	_    [4]byte       // padding
	Ki   KeyboardInput // KEYBDINPUT
	_    [8]byte       // padding to fill union (32 - sizeof(KEYBDINPUT))
}

type CallMouseInput struct {
	Type uint32
	_    [4]byte    // padding
	Mi   MouseInput // KEYBDINPUT
}
