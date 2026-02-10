package devices

import "github.com/kurlmarx/romwrangler/internal/systems"

// ReplayOS implements the Device interface for ReplayOS on Raspberry Pi.
type ReplayOS struct{}

func NewReplayOS() *ReplayOS {
	return &ReplayOS{}
}

func (r *ReplayOS) Name() string {
	return "ReplayOS"
}

func (r *ReplayOS) ROMBasePath() string {
	return "/roms"
}

func (r *ReplayOS) FolderForSystem(id systems.SystemID) (string, bool) {
	return systems.FolderForSystem(id)
}

func (r *ReplayOS) SupportedSystems() []systems.SystemID {
	result := make([]systems.SystemID, 0, len(systems.ReplayOSFolders))
	for id := range systems.ReplayOSFolders {
		result = append(result, id)
	}
	return result
}

func (r *ReplayOS) DefaultConnection() ConnectionInfo {
	return ConnectionInfo{
		Host:     "replayos.local",
		Port:     22,
		User:     "root",
		Password: "replayos",
	}
}
