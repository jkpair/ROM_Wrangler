package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BIOSFolders lists all BIOS subdirectories that should exist under the bios root.
var BIOSFolders = []string{
	"dc",
	"dc/data",
	"fbneo",
	"fbneo/samples",
	"hatari",
	"hatari/tos",
	"keropi",
	"Machines",
	"Machines/COL - ColecoVision",
	"Machines/COL - ColecoVision with Opcode Memory Extension",
	"Machines/COL - Spectravideo SVI-603 Coleco",
	"Machines/MSX",
	"Machines/MSX - Arabic",
	"Machines/MSX - Brazilian",
	"Machines/MSX - Canon V-20",
	"Machines/MSX - C-BIOS",
	"Machines/MSX - Daewoo DPC-100",
	"Machines/MSX - Daewoo DPC-180",
	"Machines/MSX - Daewoo DPC-200",
	"Machines/MSX - French",
	"Machines/MSX - German",
	"Machines/MSX - Goldstar FC-200",
	"Machines/MSX - Gradiente Expert 1.0",
	"Machines/MSX - Gradiente Expert 1.1",
	"Machines/MSX - Gradiente Expert DDPlus",
	"Machines/MSX - Gradiente Expert Plus",
	"Machines/MSX - Japanese",
	"Machines/MSX - JVC HC-7GB",
	"Machines/MSX - Korean",
	"Machines/MSX - Mitsubishi ML-F80",
	"Machines/MSX - Mitsubishi ML-FX1",
	"Machines/MSX - National CF-1200",
	"Machines/MSX - National CF-2000",
	"Machines/MSX - National CF-2700",
	"Machines/MSX - National CF-3000",
	"Machines/MSX - National CF-3300",
	"Machines/MSX - National FS-1300",
	"Machines/MSX - National FS-4000",
	"Machines/MSX - Philips NMS-801",
	"Machines/MSX - Philips VG-8020",
	"Machines/MSX - Russian",
	"Machines/MSX - Sanyo MPC-100",
	"Machines/MSX - Sharp Epcom HotBit 1.1",
	"Machines/MSX - Sharp Epcom HotBit 1.2",
	"Machines/MSX - Sony HB-201",
	"Machines/MSX - Sony HB-201P",
	"Machines/MSX - Sony HB-501P",
	"Machines/MSX - Sony HB-75D",
	"Machines/MSX - Sony HB-75P",
	"Machines/MSX - Spanish",
	"Machines/MSX - Spectravideo SVI-728",
	"Machines/MSX - Spectravideo SVI-738",
	"Machines/MSX - Spectravideo SVI-738 Henrik Gilvad",
	"Machines/MSX - Spectravideo SVI-738 Swedish",
	"Machines/MSX - Swedish",
	"Machines/MSX - Talent DPC-200",
	"Machines/MSX - Toshiba HX-10",
	"Machines/MSX - Toshiba HX-20",
	"Machines/MSX - Yamaha CX5M",
	"Machines/MSX - Yamaha CX5M-128",
	"Machines/MSX2",
	"Machines/MSX2+",
	"Machines/MSX2 - Arabic",
	"Machines/MSX2+ - Brazilian",
	"Machines/MSX2 - Brazilian",
	"Machines/MSX2+ - C-BIOS",
	"Machines/MSX2 - C-BIOS",
	"Machines/MSX2+ - Ciel Expert 3",
	"Machines/MSX2 - Daewoo CPC-300",
	"Machines/MSX2 - Daewoo CPC-400",
	"Machines/MSX2 - Daewoo CPC-400S",
	"Machines/MSX2+ - European",
	"Machines/MSX2 - French",
	"Machines/MSX2 - German",
	"Machines/MSX2 - Gradiente Expert 2.0",
	"Machines/MSX2 - Japanese",
	"Machines/MSX2 - Korean",
	"Machines/MSX2 - National FS-4500",
	"Machines/MSX2 - National FS-4600",
	"Machines/MSX2 - National FS-4700",
	"Machines/MSX2 - National FS-5000",
	"Machines/MSX2 - National FS-5500",
	"Machines/MSX2 - Only PSG",
	"Machines/MSX2 - Panasonic FS-A1",
	"Machines/MSX2 - Panasonic FS-A1 MK2",
	"Machines/MSX2 - Panasonic FS-A1F",
	"Machines/MSX2 - Panasonic FS-A1FM",
	"Machines/MSX2+ - Panasonic FS-A1FX",
	"Machines/MSX2+ - Panasonic FS-A1WSX",
	"Machines/MSX2+ - Panasonic FS-A1WX",
	"Machines/MSX2 - Philips NMS-8220",
	"Machines/MSX2 - Philips NMS-8245",
	"Machines/MSX2 - Philips NMS-8250",
	"Machines/MSX2 - Philips NMS-8255",
	"Machines/MSX2 - Philips NMS-8280",
	"Machines/MSX2 - Philips VG-8235",
	"Machines/MSX2 - Philips VG-8240",
	"Machines/MSX2 - Russian",
	"Machines/MSX2 - Sanyo Wavy PHC-23",
	"Machines/MSX2+ - Sanyo Wavy PHC-35J",
	"Machines/MSX2+ - Sanyo Wavy PHC-70FD1",
	"Machines/MSX2+ - Sanyo Wavy PHC-70FD2",
	"Machines/MSX2 - Sharp Epcom HotBit 2.0",
	"Machines/MSX2 - Sony HB-F1",
	"Machines/MSX2 - Sony HB-F1II",
	"Machines/MSX2 - Sony HB-F1XD",
	"Machines/MSX2+ - Sony HB-F1XDJ",
	"Machines/MSX2 - Sony HB-F1XDMK2",
	"Machines/MSX2+ - Sony HB-F1XV",
	"Machines/MSX2 - Sony HB-F500",
	"Machines/MSX2 - Sony HB-F500P",
	"Machines/MSX2 - Sony HB-F700D",
	"Machines/MSX2 - Sony HB-F700P",
	"Machines/MSX2 - Sony HB-F900",
	"Machines/MSX2 - Sony HB-F9P",
	"Machines/MSX2 - Sony HB-G900P",
	"Machines/MSX2 - Spanish",
	"Machines/MSX2 - Swedish",
	"Machines/MSX2 - Talent TPC-310",
	"Machines/MSX2 - Yamaha CX7M-128",
	"Machines/MSXturboR",
	"Machines/SEGA - SC-3000",
	"Machines/SEGA - SF-7000",
	"Machines/SEGA - SG-1000",
	"Machines/Shared Roms",
	"Machines/SVI - Spectravideo SVI-318",
	"Machines/SVI - Spectravideo SVI-328",
	"Machines/SVI - Spectravideo SVI-328 80 Column",
	"Machines/SVI - Spectravideo SVI-328 80 Swedish",
	"Machines/SVI - Spectravideo SVI-328 MK2",
	"Machines/Turbo-R",
	"Machines/Turbo-R - European",
	"Machines/Turbo-R - Panasonic FS-A1GT",
	"Machines/Turbo-R - Panasonic FS-A1ST",
	"mame",
	"mame/hiscore",
	"mame/ini",
	"mame/plugins",
	"mame/plugins/hiscore",
	"mame2003-plus",
	"mame2003-plus/artwork",
	"melonDS DS",
	"neocd",
	"same_cdi",
	"same_cdi/bios",
	"scummvm",
	"scummvm/extra",
	"scummvm/soundfonts",
	"scummvm/theme",
}

// RootDeviceFolders are the top-level folders on the ReplayOS device.
var RootDeviceFolders = []string{
	"bios",
	"captures",
	"config",
	"roms",
	"saves",
}

// BIOSFolderStatus describes the state of a BIOS folder.
type BIOSFolderStatus struct {
	Folder    string
	FullPath  string
	Exists    bool
	FileCount int
}

// CheckBIOSFolders checks which BIOS folders exist under biosDir.
func CheckBIOSFolders(biosDir string) []BIOSFolderStatus {
	var statuses []BIOSFolderStatus

	for _, folder := range BIOSFolders {
		fullPath := filepath.Join(biosDir, folder)
		status := BIOSFolderStatus{
			Folder:   folder,
			FullPath: fullPath,
		}

		info, err := os.Stat(fullPath)
		if err == nil && info.IsDir() {
			status.Exists = true
			entries, _ := os.ReadDir(fullPath)
			status.FileCount = len(entries)
		}

		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Folder < statuses[j].Folder
	})

	return statuses
}

// GenerateBIOSFolders creates all BIOS folders under biosDir.
func GenerateBIOSFolders(biosDir string) (created int, errs []error) {
	if err := os.MkdirAll(biosDir, 0755); err != nil {
		return 0, []error{fmt.Errorf("create bios dir: %w", err)}
	}

	for _, folder := range BIOSFolders {
		fullPath := filepath.Join(biosDir, folder)
		if _, err := os.Stat(fullPath); err == nil {
			continue
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			errs = append(errs, fmt.Errorf("create %s: %w", folder, err))
			continue
		}
		created++
	}
	return created, errs
}

// GenerateRootDeviceFolders creates the top-level device structure folders
// under rootDir (bios, captures, config, roms, saves).
func GenerateRootDeviceFolders(rootDir string) (created int, errs []error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return 0, []error{fmt.Errorf("create root dir: %w", err)}
	}

	for _, folder := range RootDeviceFolders {
		fullPath := filepath.Join(rootDir, folder)
		if _, err := os.Stat(fullPath); err == nil {
			continue
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			errs = append(errs, fmt.Errorf("create %s: %w", folder, err))
			continue
		}
		created++
	}
	return created, errs
}

// BIOSFileMatch maps a known BIOS filename to its target folder within the bios directory.
type BIOSFileMatch struct {
	Filename   string
	TargetDir  string // relative to bios root, empty string = bios root
	SourcePath string // populated during scan
}

// biosFileMap maps lowercase filenames to their target directories.
var biosFileMap = map[string]string{
	// Root-level BIOS files
	"5200.rom":               "",
	"7800 bios (u).rom":      "",
	"bios_cd_e.bin":          "",
	"bios_cd_j.bin":          "",
	"bios_cd_u.bin":          "",
	"bios_e.sms":             "",
	"bios.gg":                "",
	"bios_j.sms":             "",
	"bios_md.bin":            "",
	"bios.sms":               "",
	"bios_u.sms":             "",
	"bs-x.bin":               "",
	"disksys.rom":            "",
	"dosboxpuremidicache.txt": "",
	"gba_bios.bin":           "",
	"gb_bios.bin":            "",
	"gbc_bios.bin":           "",
	"gexpress.pce":           "",
	"kick33180.a500":         "",
	"kick34005.a500":         "",
	"kick34005.cdtv":         "",
	"kick37175.a500":         "",
	"kick37350.a600":         "",
	"kick39106.a1200":        "",
	"kick39106.a4000":        "",
	"kick40060.cd32":         "",
	"kick40060.cd32.ext":     "",
	"kick40063.a600":         "",
	"kick40068.a1200":        "",
	"kick40068.a4000":        "",
	"lynxboot.img":           "",
	"mpr-17933.bin":          "",
	"panafz10.bin":           "",
	"scph5500.bin":           "",
	"scph5501.bin":           "",
	"scph5502.bin":           "",
	"scummvm.ini":            "",
	"sega_101.bin":           "",
	"syscard1.pce":           "",
	"syscard2.pce":           "",
	"syscard3.pce":           "",

	// dc/
	"airlbios.zip":  "dc",
	"awbios.zip":    "dc",
	"dc_boot.bin":   "dc",
	"f355bios.zip":  "dc",
	"f355dlx.zip":   "dc",
	"f355.zip":      "dc",
	"hod2bios.zip":  "dc",
	"naomi2.zip":    "dc",
	"naomi.zip":     "dc",
	"segasp.zip":    "dc",

	// fbneo/
	"000-lo.lo":     "fbneo",
	"bubsys.zip":    "fbneo",
	"cchip.zip":     "fbneo",
	"coleco.zip":    "fbneo",
	"decocass.zip":  "fbneo",
	"front-sp1.bin": "fbneo",
	"hiscore.dat":   "fbneo",
	"isgsm.zip":     "fbneo",
	"m68705p5.zip":  "fbneo",
	"midssio.zip":   "fbneo",
	"msx.zip":       "fbneo",
	"namcoc69.zip":  "fbneo",
	"namcoc70.zip":  "fbneo",
	"namcoc75.zip":  "fbneo",
	"neocdz.zip":    "fbneo",
	"neogeo.zip":    "fbneo",
	"nmk004.zip":    "fbneo",
	"pgm.zip":       "fbneo",
	"phoenix.key":   "fbneo",
	"skns.zip":      "fbneo",
	"top-sp1.bin":   "fbneo",
	"ym2608.zip":    "fbneo",

	// hatari/tos/
	"tos.img": "hatari/tos",

	// keropi/
	"cgrom.dat":    "keropi",
	"iplrom30.dat": "keropi",
	"iplromco.dat": "keropi",
	"iplrom.dat":   "keropi",
	"iplromxv.dat": "keropi",

	// melonDS DS/
	"bios7.bin":      "melonDS DS",
	"bios9.bin":      "melonDS DS",
	"dsi_bios7.bin":  "melonDS DS",
	"dsi_bios9.bin":  "melonDS DS",
	"dsi_firmware.bin": "melonDS DS",
	"dsi_nand.bin":   "melonDS DS",
	"firmware.bin":   "melonDS DS",

	// neocd/
	"neocd.srm":      "neocd",
	"neocd_z.rom":    "neocd",
	"uni-bioscd.rom": "neocd",

	// same_cdi/bios/
	"cdibios.zip":  "same_cdi/bios",
	"cdimono1.zip": "same_cdi/bios",
	"cdimono2.zip": "same_cdi/bios",
}

// ScanBIOSFiles scans sourceDir for known BIOS files and returns matches
// with their target directories.
func ScanBIOSFiles(sourceDir string) []BIOSFileMatch {
	if sourceDir == "" {
		return nil
	}

	// Build a set of known filenames (lowercase)
	known := make(map[string]string, len(biosFileMap))
	for k, v := range biosFileMap {
		known[k] = v
	}

	var matches []BIOSFileMatch
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "_archive" {
				return filepath.SkipDir
			}
			return nil
		}

		nameLower := strings.ToLower(info.Name())
		if targetDir, ok := known[nameLower]; ok {
			matches = append(matches, BIOSFileMatch{
				Filename:   info.Name(),
				TargetDir:  targetDir,
				SourcePath: path,
			})
		}
		return nil
	})

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Filename < matches[j].Filename
	})

	return matches
}

// OrganizeBIOSFiles moves detected BIOS files to their correct locations
// under biosDir. Returns the number of files moved and any errors.
func OrganizeBIOSFiles(matches []BIOSFileMatch, biosDir string) (moved int, errs []error) {
	for _, m := range matches {
		targetPath := filepath.Join(biosDir, m.TargetDir, m.Filename)

		// Skip if already in correct location
		if m.SourcePath == targetPath {
			continue
		}

		// Skip if target already exists
		if _, err := os.Stat(targetPath); err == nil {
			continue
		}

		// Ensure target directory exists
		targetDirPath := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDirPath, 0755); err != nil {
			errs = append(errs, fmt.Errorf("create dir for %s: %w", m.Filename, err))
			continue
		}

		if err := os.Rename(m.SourcePath, targetPath); err != nil {
			errs = append(errs, fmt.Errorf("move %s: %w", m.Filename, err))
			continue
		}
		moved++
	}
	return moved, errs
}
