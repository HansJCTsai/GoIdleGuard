//go:build windows
// +build windows

package preventidle

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/HanksJCTsai/goidleguard/pkg/logger"
)

var (
	modkernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procSetThreadExecutionState = modkernel32.NewProc("SetThreadExecutionState")
	procGetTickCount            = modkernel32.NewProc("GetTickCount")

	user32               = syscall.NewLazyDLL("user32.dll")
	procGetLastInputInfo = user32.NewProc("GetLastInputInfo")
	procSendInput        = user32.NewProc("SendInput")
)

const (
	ES_CONTINUOUS       = 0x80000000
	ES_SYSTEM_REQUIRED  = 0x00000001
	ES_DISPLAY_REQUIRED = 0x00000002

	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	// keyboard event flags
	KEYEVENTF_KEYUP = 0x0002

	// virtual key codes
	VK_SPACE = 0x20
)

// PreventSleep 建立「PreventUserIdleSystemSleep」宣告
func PreventSleep() error {
	flags := ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED
	ret, _, err := procSetThreadExecutionState.Call(uintptr(flags))
	if ret == 0 {
		return fmt.Errorf("SetThreadExecutionState failed: %v", err)
	}
	logger.LogInfo("Windows: PreventSleep asserted")
	return nil
}

// AllowIdle 恢復系統與顯示器閒置行為
func AllowIdle() error {
	ret, _, err := procSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
	if ret == 0 {
		return fmt.Errorf("SetThreadExecutionState clear failed: %v", err)
	}
	logger.LogInfo("Windows: AllowSleep restored")
	return nil
}

// CallSendInput 使用 Core Graphics API 模擬鍵盤或滑鼠事件。
// 當 mode 為 "key" 時，以 Shift 鍵 (key code 56) 為示例產生鍵盤按下與釋放事件；
// 當 mode 為 "mouse" 時，取得目前滑鼠位置，並向右平移 1 像素後建立滑鼠移動事件
func CallSendInput(mode string) error {
	switch mode {
	case "key":
		// key down
		ki := CallKeyboardInput{
			Type: INPUT_KEYBOARD,
			Ki: KeyboardInput{
				WVk:         VK_SPACE,
				WScan:       0,
				DwFlags:     0,
				Time:        0,
				DwExtraInfo: 0,
			},
		}
		n, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&ki)), unsafe.Sizeof(ki))
		if n == 0 {
			return fmt.Errorf("SendInput key down failed: %v", err)
		}
		// key up
		ki.Ki.DwFlags = KEYEVENTF_KEYUP
		n, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&ki)), unsafe.Sizeof(ki))
		if n == 0 {
			return fmt.Errorf("SendInput key up failed: %v", err)
		}
		logger.LogInfo("Windows: simulated key press (space)")
		return nil

	case "mouse":
		// get current cursor position and move by +1 X
		// here we simply send relative move
		mi := CallMouseInput{
			Type: INPUT_MOUSE,
			Mi: MouseInput{
				dx:          10,
				dy:          0,
				dwFlags:     0x0001, // MOUSEEVENTF_MOVE
				dwExtraInfo: 0,
			},
		}
		n, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&mi)), unsafe.Sizeof(mi))
		if n == 0 {
			return fmt.Errorf("SendInput mouse move failed: %v", err)
		}

		mi = CallMouseInput{
			Type: INPUT_MOUSE,
			Mi: MouseInput{
				dx:          10,
				dy:          0,
				dwFlags:     0x0001, // MOUSEEVENTF_MOVE
				dwExtraInfo: 0,
			},
		}
		n, _, err = procSendInput.Call(1, uintptr(unsafe.Pointer(&mi)), unsafe.Sizeof(mi))
		if n == 0 {
			return fmt.Errorf("SendInput mouse move failed: %v", err)
		}

		logger.LogInfo("Windows: simulated mouse move")
		return nil

	default:
		return fmt.Errorf("unsupported mode for CallSendInput on Windows: %s", mode)
	}
}

// GetIdleTime 使用 CGEventSourceSecondsSinceLastEventType 取得系統閒置時間（以秒計），並轉換為 time.Duration。
func GetIdleTime() (time.Duration, error) {
	// 先取得 LASTINPUTINFO
	var li LastInputInfo
	li.cbSize = uint32(unsafe.Sizeof(li))
	ret, _, err := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&li)))
	if ret == 0 {
		return 0, fmt.Errorf("GetLastInputInfo failed: %v", err)
	}

	// 再呼叫 GetTickCount
	retTick, _, err := procGetTickCount.Call()
	if retTick == 0 {
		return 0, fmt.Errorf("GetTickCount failed: %v", err)
	}
	uptimeMs := uint32(retTick)

	// 計算閒置毫秒數
	idleMs := uptimeMs - li.dwTime
	return time.Duration(idleMs) * time.Millisecond, nil
}
