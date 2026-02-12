package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kurlmarx/romwrangler/internal/config"
	"github.com/kurlmarx/romwrangler/internal/tui"
	"github.com/kurlmarx/romwrangler/internal/tui/screens"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	factory := func(id tui.ScreenID, cfg *config.Config, width, height int) tui.Screen {
		switch id {
		case tui.ScreenHome:
			return screens.NewHomeScreen(cfg, width, height)
		case tui.ScreenManage:
			return screens.NewManageScreen(cfg, width, height)
		case tui.ScreenDecompress:
			return screens.NewDecompressScreen(cfg, width, height)
		case tui.ScreenArchive:
			return screens.NewArchiveScreen(cfg, width, height)
		case tui.ScreenConvert:
			return screens.NewConvertScreen(cfg, width, height)
		case tui.ScreenTransfer:
			return screens.NewTransferScreen(cfg, width, height)
		case tui.ScreenSettings:
			return screens.NewSettingsScreen(cfg, width, height)
		case tui.ScreenSetup:
			return screens.NewSetupScreen(cfg, width, height)
		case tui.ScreenReplayOS:
			return screens.NewReplayOSScreen(width, height)
		case tui.ScreenBIOS:
			return screens.NewBIOSSetupScreen(cfg, width, height)
		case tui.ScreenM3U:
			return screens.NewM3UScreen(cfg, width, height)
		default:
			return screens.NewHomeScreen(cfg, width, height)
		}
	}

	app := tui.NewApp(cfg, factory)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
