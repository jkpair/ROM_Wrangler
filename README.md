# ROM Wrangler

A terminal UI application for organizing, converting, and transferring ROMs and disc images to retro gaming devices. Built for [ReplayOS](https://replayos.com) on Raspberry Pi, with an extensible architecture for other targets.

ROM Wrangler handles the full workflow: scan your ROM collection, extract archives, decompress ECM files, convert disc images to CHD, generate M3U playlists for multi-disc games, clean up redundant files, and transfer everything to your device over SFTP or USB.

## Features

- **50 supported systems** across 16 manufacturers — Amstrad, Atari, Commodore, Microsoft, NEC, Nintendo, Panasonic, Philips, Sega, Sharp, Sinclair, SNK, Sony, plus arcade (FBNeo, MAME, MAME 2K3+, Naomi/Atomiswave) and home computers (DOS, ScummVM)
- **Smart folder detection** — 100+ aliases map common names like `genesis`, `psx`, `snes`, `mame`, `dreamcast` to the correct ReplayOS folder structure
- **Archive extraction** — extracts .zip, .7z, and .rar archives in disc-based system folders into per-game subfolders, preventing track file name collisions
- **Native ECM decompression** — decompresses .bin.ecm files to .bin with no external tools required
- **Iterative extraction** — handles nested archives (e.g. a .rar containing a .bin.ecm) by re-scanning after each pass
- **CHD conversion** — batch convert GDI, CUE/BIN, and ISO disc images to CHD via chdman, with live per-file progress
- **CUE file auto-repair** — fixes case mismatches in FILE references and patches `.bin.ecm` references after ECM decompression
- **Multi-disc detection** — automatically groups disc sets and generates M3U playlists
- **Game identification** — hash-based lookup via No-Intro/Redump DAT files and ScreenScraper API
- **Filename cleaning** — strips dump tags (`[!]`, `[b1]`, serials) while preserving region and disc info
- **High-performance transfers** — SFTP with concurrent writes/reads and 256KB buffer pooling, or USB with 1MB buffers and Linux `fallocate` pre-allocation
- **Parallel transfers** — configurable concurrency for transferring multiple files simultaneously
- **Transfer cancellation** — press Esc during a transfer to cancel in-flight uploads
- **Sync mode** — skip files that already exist on the destination (by size match)
- **BIOS setup** — guided BIOS file organization for all supported systems
- **Deferred archiving** — original disc images and spent archives are moved to `_archive/` only after successful conversion, with optional auto-deletion
- **Redundant file cleanup** — detect and archive duplicate versions, superseded disc images, and already-extracted archives
- **SQLite cache** — scraping results are cached locally so repeat lookups are instant
- **Config file** — YAML config at `~/.config/romwrangler/config.yaml`, editable in the TUI or by hand
- **No CGo** — pure Go build using modernc.org/sqlite, compiles anywhere Go runs

## Requirements

- **Go 1.25+** to build from source
- **chdman** (optional) — for disc image conversion, part of MAME tools
- **7z** (optional) — for extracting .7z and .rar archives (p7zip or 7-zip)

### Installing optional dependencies

```
# Arch / Manjaro / CachyOS
sudo pacman -S mame-tools p7zip

# Ubuntu / Debian
sudo apt install mame-tools p7zip-full

# Fedora
sudo dnf install mame-tools p7zip
```

ECM decompression (`.bin.ecm` files) is handled natively — no external `unecm` tool is needed.

## Building from Source

```bash
git clone https://github.com/jkpair/ROM_Wrangler.git
cd ROM_Wrangler
make build
```

The binary is output to `bin/romwrangler`. Copy it somewhere on your PATH:

```bash
cp bin/romwrangler ~/.local/bin/
```

Or build and run directly:

```bash
make run
```

## Quick Start

1. **Launch the TUI:**

   ```bash
   romwrangler
   ```

2. **Configure source directories:** Navigate to **Settings > General** and enter the paths to your source folder for bios, roms, etc. (example: /home/kurlmarx/Games/replayos)

Should look like this

<img width="525" height="555" alt="Screenshot_20260211_233240" src="https://github.com/user-attachments/assets/37eb7f2f-d143-442b-8ec8-c8897ef25146" />

Press `Ctrl+S` to save.

4. **Organize your ROMs:** Go to **Manage ROMs**. The scanner expects your source directories to contain subfolders named after systems — for example:

   ```
   /home/you/roms/
     nes/
       Super Mario Bros. (USA).nes
     genesis/
       Sonic the Hedgehog (USA).md
     psx/
       Final Fantasy VII (USA) (Disc 1).cue
       Final Fantasy VII (USA) (Disc 1).bin
       Final Fantasy VII (USA) (Disc 2).cue
       Final Fantasy VII (USA) (Disc 2).bin
       Final Fantasy VII (USA) (Disc 3).cue
       Final Fantasy VII (USA) (Disc 3).bin
     sega_dc/
       SomeGame.zip      (extracted automatically)
       AnotherGame.rar   (extracted automatically)
   ```

   The scanner resolves folder names like `nes`, `genesis`, `psx`, `dreamcast`, `snes`, `gba`, `mame`, etc. to the correct system. See the full alias list below.

5. **The Manage ROMs pipeline:**

   The manage screen walks you through each step, skipping any that don't apply:

   - **Scan** — finds ROM files and archives across your source directories
   - **Review** — shows files by system, archive counts, and convertible file counts
   - **Extract** — extracts .zip/.7z/.rar archives into per-game subfolders (prevents track file collisions), decompresses .bin.ecm to .bin, with a progress bar
   - **Re-scan** — picks up newly extracted files
   - **Convert** — batch converts disc images (GDI/CUE/ISO) to CHD via chdman, with per-file progress bars
   - **Archive** — moves original archives and pre-conversion disc images to `_archive/`
   - **Sort** — reviews the sort plan, then moves/copies files into the ReplayOS folder structure with M3U playlists for multi-disc sets

   The result is your source directory organized with the correct folder structure:

   ```
   /home/you/roms/
     nintendo_nes/
       Super Mario Bros. (USA).nes
     sega_smd/
       Sonic the Hedgehog (USA).md
     sony_psx/
       Final Fantasy VII (USA) (Disc 1).chd
       Final Fantasy VII (USA) (Disc 2).chd
       Final Fantasy VII (USA) (Disc 3).chd
       Final Fantasy VII (USA).m3u
   ```

6. **Transfer to your device:** Go to **Transfer** and choose SFTP, USB, or manual. Select which folders to send (ROMs, BIOS, Saves, Config), review the transfer plan, and start. Progress is tracked per-file and overall.

## Home Screen

| Menu Item | Description |
|---|---|
| Manage ROMs | Full pipeline: scan, organize, convert disc images to CHD, and generate M3U for multi-disc games |
| Decompress Files | Extract .zip, .7z, .rar, and .ecm archives |
| Convert Files | Convert disc images to CHD format |
| Generate M3U Files | Generate M3U playlists for multi-disc games |
| Transfer | Send files to your gaming device via SFTP or USB |
| Archive Redundant Files | Clean up duplicates, superseded disc images, and spent archives |
| Settings | Configure devices, paths, and options |
| About ReplayOS | Learn more about ReplayOS and support the project |

## Configuration

Config is stored at `~/.config/romwrangler/config.yaml`. A default config is created on first launch. You can edit it in the TUI under Settings, or by hand:

```yaml
source_dirs:
  - /home/you/roms

chdman_path: ""  # leave empty to auto-detect from PATH

delete_archive: false  # set to true to auto-delete _archive/ after organizing

device:
  type: replayos
  host: replayos.local
  port: 22
  user: root
  password: replayos
  root_path: /

scraping:
  screenscraper_user: ""
  screenscraper_pass: ""
  dat_dirs: []

transfer:
  method: sftp
  sync_mode: true
  usb_path: ""
  concurrency: 1  # increase for parallel transfers

aliases:
  # Add custom aliases here, e.g.:
  # my_roms: sega_dc
```

Use `--config /path/to/config.yaml` to use an alternate config file.

### Configuration Options

| Option | Description | Default |
|---|---|---|
| `source_dirs` | List of directories containing your ROM subfolders | (none) |
| `chdman_path` | Path to chdman binary (leave empty to auto-detect) | auto |
| `delete_archive` | Auto-delete `_archive/` directory after organizing | false |
| `device.type` | Device type | replayos |
| `device.host` | Hostname or IP of your ReplayOS device | replayos.local |
| `device.port` | SSH port | 22 |
| `device.user` | SSH username | root |
| `device.password` | SSH password | replayos |
| `device.root_path` | Root path on the device | / |
| `transfer.method` | Transfer method (`sftp` or `usb`) | sftp |
| `transfer.sync_mode` | Skip files that already exist on the destination | true |
| `transfer.usb_path` | Mount path for USB/SD card transfers | (none) |
| `transfer.concurrency` | Number of parallel transfer workers | 1 |
| `scraping.screenscraper_user` | ScreenScraper API username | (none) |
| `scraping.screenscraper_pass` | ScreenScraper API password | (none) |
| `scraping.dat_dirs` | Directories containing No-Intro/Redump DAT files | (none) |

## Keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select / confirm |
| `Esc` | Go back / cancel transfer |
| `Space` | Toggle selection |
| `a` | Select all / deselect all |
| `Tab` | Next field (in Settings) |
| `Ctrl+S` | Save (in Settings) |
| `q` / `Ctrl+C` | Quit (from home screen) |

## Supported Systems

| Company | Systems |
|---|---|
| Amstrad | CPC |
| Atari | 2600, 5200, 7800, Lynx, Jaguar |
| Commodore | C64, Amiga, Amiga CD32 |
| Microsoft | MSX, MSX2 |
| NEC | PC Engine / TurboGrafx-16, PC Engine CD / TurboGrafx-CD |
| Nintendo | NES, Famicom Disk System, SNES, N64, Game Boy, Game Boy Color, Game Boy Advance, DS |
| Panasonic | 3DO |
| Philips | CD-i |
| Sega | SG-1000, Master System, Mega Drive / Genesis, 32X, Sega CD / Mega CD, Saturn, Dreamcast, Game Gear |
| Sharp | X68000 |
| Sinclair | ZX Spectrum |
| SNK | Neo Geo, Neo Geo CD, Neo Geo Pocket, Neo Geo Pocket Color |
| Sony | PlayStation |
| Arcade | FBNeo, MAME, MAME 2K3+, Naomi / Atomiswave |
| Other | DOS (DOSBox), ScummVM |

## Folder Aliases

The scanner recognizes these folder names (case-insensitive). ReplayOS folder names (e.g. `nintendo_snes`, `sega_dc`) are also recognized automatically. You can add custom aliases in the config.

<details>
<summary>Full alias list (click to expand)</summary>

| Alias | System |
|---|---|
| `arcade`, `fbneo`, `fba` | Arcade (FBNeo) |
| `mame` | Arcade (MAME) |
| `mame2003`, `mame2003plus`, `mame 2003`, `mame2k3p` | Arcade (MAME 2K3+) |
| `naomi`, `atomiswave` | Arcade (Naomi/Atomiswave) |
| `amstrad`, `cpc` | Amstrad CPC |
| `2600`, `atari2600`, `atari 2600`, `vcs` | Atari 2600 |
| `5200`, `atari5200`, `atari 5200` | Atari 5200 |
| `7800`, `atari7800`, `atari 7800` | Atari 7800 |
| `lynx`, `atarilynx`, `atari lynx` | Atari Lynx |
| `jaguar`, `atarijaguar`, `atari jaguar` | Atari Jaguar |
| `c64`, `commodore64`, `commodore 64` | Commodore 64 |
| `amiga` | Amiga |
| `amigacd`, `amigacd32`, `amiga cd32`, `cd32` | Amiga CD32 |
| `msx` | MSX |
| `msx2` | MSX2 |
| `pce`, `pcengine`, `pc engine`, `turbografx`, `turbografx16`, `turbografx-16`, `tg16` | PC Engine / TurboGrafx-16 |
| `pcecd`, `pcenginecd`, `turbografxcd`, `tgcd` | PC Engine CD / TurboGrafx-CD |
| `nes`, `famicom`, `fc`, `nintendo` | NES |
| `fds` | Famicom Disk System |
| `snes`, `supernes`, `superfamicom`, `sfc` | SNES |
| `n64`, `nintendo64`, `nintendo 64` | Nintendo 64 |
| `gb`, `gameboy`, `game boy` | Game Boy |
| `gbc`, `gameboycolor`, `game boy color` | Game Boy Color |
| `gba`, `gameboyadvance`, `game boy advance` | Game Boy Advance |
| `nds`, `ds`, `nintendods`, `nintendo ds` | Nintendo DS |
| `3do` | 3DO |
| `cdi`, `cd-i` | CD-i |
| `sg1000`, `sg-1000` | SG-1000 |
| `mastersystem`, `master system`, `sms` | Master System |
| `megadrive`, `mega drive`, `genesis`, `md`, `gen`, `smd` | Mega Drive / Genesis |
| `32x`, `sega32x` | 32X |
| `segacd`, `sega cd`, `megacd`, `mega cd` | Sega CD / Mega CD |
| `saturn`, `segasaturn`, `sega saturn` | Saturn |
| `dreamcast`, `dc` | Dreamcast |
| `gamegear`, `game gear`, `gg` | Game Gear |
| `x68000`, `x68k` | Sharp X68000 |
| `zxspectrum`, `zx spectrum`, `spectrum` | ZX Spectrum |
| `neogeo`, `neo geo`, `neo-geo` | Neo Geo |
| `neogeocd`, `neo geo cd` | Neo Geo CD |
| `ngp`, `neopocket` | Neo Geo Pocket |
| `ngpc`, `neopocketcolor` | Neo Geo Pocket Color |
| `psx`, `ps1`, `playstation`, `playstation1`, `playstation 1` | PlayStation |
| `dos`, `dosbox` | DOS |
| `mediaplayer`, `media` | Media Player |

</details>

## Transfer to ReplayOS

### SFTP (default)

ReplayOS exposes SSH on port 22 with default credentials `root:replayos`. ROM Wrangler connects and uploads files directly in the correct folder structure. The SFTP backend is tuned for maximum throughput with:

- 256KB packet sizes and 64 concurrent requests per file
- Concurrent writes and reads enabled
- Buffer pooling to reduce allocations
- Configurable parallel file transfers (`transfer.concurrency`)
- Context-aware cancellation (press Esc to stop)

Make sure your Pi is on the network and reachable at `replayos.local` (or set the IP in Settings).

### USB / SD Card

Set the USB mount path in Settings (e.g., `/media/you/USBDRIVE`), then use the USB transfer option. The USB backend uses:

- 1MB buffer pooling for maximum throughput
- Linux `fallocate` pre-allocation to reduce fragmentation
- `fsync` after each file to ensure data is written to disk
- Parallel transfers and sync mode support

### Manual

If neither SFTP nor USB works, organize your ROMs with Manage ROMs, then manually copy the organized files to your device.

## Development

```bash
make build    # Build binary to bin/romwrangler
make run      # Build and run
make test     # Run all tests
make lint     # Run golangci-lint
make clean    # Remove build artifacts
```

### Project Structure

```
cmd/romwrangler/        Entry point, screen factory
internal/
  config/               Config loading, system aliases (100+)
  devices/              Device interface (ReplayOS)
  systems/              50 system definitions, formats, folder maps
  converter/            chdman wrapper, progress parsing, batch runner
  scraper/              DAT parser, ScreenScraper API, hasher, identifier
  romdb/                SQLite cache for scraping results
  multidisc/            Disc pattern detection, M3U generation
  organizer/            Scanner, renamer, sorter, plan executor,
                        archive extraction, ECM decompression,
                        BIOS folder setup, redundant file detection
  transfer/             SFTP and USB backends with context cancellation,
                        buffer pooling, parallel execution, progress tracking
  tui/                  Bubbletea app, theme, keys, screen management
    screens/            Home, manage, decompress, convert, M3U, transfer,
                        archive, settings, setup, BIOS setup, ReplayOS info
    components/         Reusable header and status bar components
```

### Running Tests

```bash
go test ./...
```

Tests cover alias resolution, format validation, folder mapping, chdman progress parsing, DAT parsing, hashing, SQLite cache, multi-disc detection (including revision suffix handling), M3U generation, filename cleaning, scanning, USB transfer with context cancellation, ECM decompression, CUE file repair, and archive extraction.

## License

MIT
