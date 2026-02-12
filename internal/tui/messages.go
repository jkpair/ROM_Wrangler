package tui

// NavigateMsg requests navigation to a screen.
type NavigateMsg struct {
	Screen ScreenID
}

// NavigateBackMsg requests going back one screen.
type NavigateBackMsg struct{}

// ErrorMsg represents an error to display.
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string { return e.Err.Error() }

// ProgressMsg reports progress from a background operation.
type ProgressMsg struct {
	ID       string
	Current  int64
	Total    int64
	Filename string
	Done     bool
	Err      error
}

// StatusMsg sets the status bar text.
type StatusMsg struct {
	Text string
}

// ScreenID identifies TUI screens.
type ScreenID int

const (
	ScreenHome ScreenID = iota
	ScreenManage
	ScreenDecompress
	ScreenArchive
	ScreenConvert
	ScreenTransfer
	ScreenSettings
	ScreenSetup
	ScreenReplayOS
	ScreenBIOS
	ScreenM3U
)
