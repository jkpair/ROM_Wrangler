# ROM Wrangler

A terminal UI application for organizing, converting, and transferring ROMs and disc images to retro gaming devices. Built for [ReplayOS](https://replayos.com) on Raspberry Pi, with an extensible architecture for other targets.

ROM Wrangler handles the full workflow: scan your ROM collection, extract disc image archives, convert to CHD, clean up filenames, detect multi-disc sets and generate M3U playlists, and transfer everything to your device over SFTP or USB.

## Features

- **50+ supported systems** across 15 manufacturers (Atari, Bandai, Coleco, Commodore, GCE, Magnavox/Philips, Mattel, Microsoft, NEC, Nintendo, Panasonic, Sega, SNK, Sony, plus arcade and home computers)
- **Smart folder detection** — 100+ aliases map common names like `genesis`, `psx`, `snes` to the correct system
- **Archive extraction** — automatically extracts .zip, .7z, and .rar archives in disc-based system folders into per-game subfolders, preventing track file name collisions
- **Native ECM decompression** — decompresses .bin.ecm files to .bin with no external tools required
- **Iterative extraction** — handles nested archives (e.g. a .rar containing a .bin.ecm) by re-scanning after each pass
- **CHD conversion** — batch convert GDI, CUE/BIN, and ISO disc images to CHD via chdman, with live progress
- **CUE file auto-repair** — fixes case mismatches in FILE references (e.g. `.BIN` vs `.bin`) and patches `.bin.ecm` references after ECM decompression
- **Multi-disc detection** — automatically groups disc sets and generates M3U playlists
- **Game identification** — hash-based lookup via No-Intro/Redump DAT files and ScreenScraper API
- **Filename cleaning** — strips dump tags (`[!]`, `[b1]`, serials) while preserving region and disc info
- **Transfer** — SFTP (tuned for maximum throughput) or USB/SD card copy, with progress bars
- **Sync mode** — skip files that already exist on the destination (by size match)
- **Deferred archiving** — original disc images and extracted archives are moved to `_archive/` only after successful conversion, with optional auto-deletion
- **SQLite cache** — scraping results are cached locally so repeat lookups are instant
- **Config file** — YAML config at `~/.config/romwrangler/config.yaml`, editable in the TUI or by hand
- **No CGo** — pure Go build using modernc.org/sqlite, compiles anywhere Go runs

## Requirements

- **Go 1.23+** to build from source
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

2. **Configure source directories:** Navigate to **Settings > General** and enter the paths to your ROM folders (comma-separated). Press `Ctrl+S` to save.

3. **Organize your ROMs:** Go to **Manage ROMs**. The scanner expects your source directories to contain subfolders named after systems — for example:

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

4. **The Manage ROMs pipeline:**

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
     sega_md/
       Sonic the Hedgehog (USA).md
     sony_psx/
       Final Fantasy VII (USA) (Disc 1).chd
       Final Fantasy VII (USA) (Disc 2).chd
       Final Fantasy VII (USA) (Disc 3).chd
       Final Fantasy VII (USA).m3u
   ```

5. **Transfer to your device:** Go to **Transfer** and choose SFTP, USB, or manual. Files are transferred from your source directory directly.

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
  rom_path: /roms

scraping:
  screenscraper_user: ""
  screenscraper_pass: ""
  dat_dirs: []

transfer:
  method: sftp
  sync_mode: true
  usb_path: ""
  concurrency: 1

aliases:
  # Add custom aliases here, e.g.:
  # my_roms: sega_dc
```

Use `--config /path/to/config.yaml` to use an alternate config file.

### Configuration Options

| Option | Description |
|---|---|
| `source_dirs` | List of directories containing your ROM subfolders |
| `chdman_path` | Path to chdman binary (leave empty to auto-detect) |
| `delete_archive` | Auto-delete `_archive/` directory after organizing |
| `device.host` | Hostname or IP of your ReplayOS device |
| `device.port` | SSH port (default: 22) |
| `device.user` | SSH username (default: root) |
| `device.password` | SSH password (default: replayos) |
| `device.rom_path` | ROM directory on the device (default: /roms) |
| `transfer.sync_mode` | Skip files that already exist on the destination |
| `transfer.usb_path` | Mount path for USB/SD card transfers |

## Keybindings

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select / confirm |
| `Esc` | Go back |
| `Space` | Toggle selection |
| `a` | Select all / deselect all |
| `s` | Skip current step (extract/convert) |
| `d` | Delete archive (on results screen) |
| `Tab` | Next field (in Settings) |
| `Ctrl+S` | Save (in Settings) |
| `q` / `Ctrl+C` | Quit (from home screen) |

## Supported Systems

| Company | Systems |
|---|---|
| Atari | 2600, 5200, 7800, Lynx, Jaguar, ST |
| Bandai | WonderSwan, WonderSwan Color |
| Coleco | ColecoVision |
| Commodore | C64, Amiga |
| GCE | Vectrex |
| Magnavox/Philips | Odyssey 2, CD-i |
| Mattel | Intellivision |
| Microsoft | Xbox |
| NEC | PC Engine / TG-16, PC Engine CD / TG-CD, SuperGrafx, PC-FX |
| Nintendo | NES, FDS, SNES, N64, GameCube, Wii, Game Boy, GBC, GBA, DS, Virtual Boy, Pokemon Mini |
| Panasonic | 3DO |
| Sega | SG-1000, Master System, Genesis / Mega Drive, 32X, Sega CD / Mega CD, Saturn, Dreamcast, Game Gear |
| SNK | Neo Geo, Neo Geo CD, Neo Geo Pocket, Neo Geo Pocket Color |
| Sony | PlayStation, PS2, PSP |
| Other | DOS, ScummVM, MSX, MSX2, Arcade |

## Folder Aliases

The scanner recognizes these folder names (case-insensitive). You can also add custom aliases in the config.

<details>
<summary>Full alias list (click to expand)</summary>

| Alias | System |
|---|---|
| `2600`, `atari2600`, `vcs` | Atari 2600 |
| `5200`, `atari5200` | Atari 5200 |
| `7800`, `atari7800` | Atari 7800 |
| `lynx`, `atarilynx` | Atari Lynx |
| `jaguar`, `atarijaguar` | Atari Jaguar |
| `atarist` | Atari ST |
| `wonderswan`, `ws` | WonderSwan |
| `wonderswancolor`, `wsc` | WonderSwan Color |
| `colecovision`, `coleco` | ColecoVision |
| `c64`, `commodore64` | Commodore 64 |
| `amiga` | Amiga |
| `vectrex` | Vectrex |
| `odyssey2` | Odyssey 2 |
| `cdi`, `cd-i` | CD-i |
| `intellivision`, `intv` | Intellivision |
| `xbox` | Xbox |
| `pce`, `pcengine`, `turbografx`, `tg16` | PC Engine / TG-16 |
| `pcecd`, `pcenginecd`, `turbografxcd`, `tgcd` | PC Engine CD |
| `supergrafx`, `sgrfx` | SuperGrafx |
| `pcfx`, `pc-fx` | PC-FX |
| `nes`, `famicom`, `fc`, `nintendo` | NES |
| `fds` | Famicom Disk System |
| `snes`, `supernes`, `superfamicom`, `sfc` | SNES |
| `n64`, `nintendo64` | Nintendo 64 |
| `gamecube`, `gc`, `ngc` | GameCube |
| `wii` | Wii |
| `gb`, `gameboy` | Game Boy |
| `gbc`, `gameboycolor` | Game Boy Color |
| `gba`, `gameboyadvance` | Game Boy Advance |
| `nds`, `ds`, `nintendods` | Nintendo DS |
| `virtualboy`, `vb` | Virtual Boy |
| `pokemini` | Pokemon Mini |
| `3do` | 3DO |
| `sg1000`, `sg-1000` | SG-1000 |
| `mastersystem`, `sms` | Master System |
| `megadrive`, `genesis`, `md`, `gen` | Genesis / Mega Drive |
| `32x`, `sega32x` | 32X |
| `segacd`, `megacd` | Sega CD / Mega CD |
| `saturn`, `segasaturn` | Saturn |
| `dreamcast`, `dc` | Dreamcast |
| `gamegear`, `gg` | Game Gear |
| `neogeo`, `neo-geo` | Neo Geo |
| `neogeocd` | Neo Geo CD |
| `ngp`, `neopocket` | Neo Geo Pocket |
| `ngpc`, `neopocketcolor` | Neo Geo Pocket Color |
| `psx`, `ps1`, `playstation` | PlayStation |
| `ps2`, `playstation2` | PlayStation 2 |
| `psp` | PSP |
| `dos`, `dosbox` | DOS |
| `scummvm` | ScummVM |
| `msx` | MSX |
| `msx2` | MSX2 |
| `arcade`, `mame`, `fbneo`, `fba` | Arcade |

</details>

## Transfer to ReplayOS

### SFTP (default)

ReplayOS exposes SSH on port 22 with default credentials `root:replayos`. ROM Wrangler connects and uploads files directly to `/roms/` in the correct folder structure. The SFTP backend is tuned for maximum throughput with large packet sizes and concurrent requests.

Make sure your Pi is on the network and reachable at `replayos.local` (or set the IP in Settings).

### USB / SD Card

Set the USB mount path in Settings (e.g., `/media/you/USBDRIVE`), then use the USB transfer option. Files are copied with the same folder structure and progress tracking.

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
cmd/romwrangler/        Entry point
internal/
  config/               Config loading, system aliases
  devices/              Device interface (ReplayOS)
  systems/              System definitions, formats, folder maps
  converter/            chdman wrapper, progress parsing, batch runner
  scraper/              DAT parser, ScreenScraper API, hasher, identifier
  romdb/                SQLite cache for scraping results
  multidisc/            Disc pattern detection, M3U generation
  organizer/            Scanner, renamer, sorter, plan executor,
                        archive extraction, ECM decompression
  transfer/             SFTP and USB backends, progress tracking
  tui/                  Bubbletea app, screens, components
```

### Running Tests

```bash
go test ./...
```

Tests cover alias resolution, format validation, folder mapping, chdman progress parsing, DAT parsing, hashing, SQLite cache, multi-disc detection, M3U generation, filename cleaning, scanning, USB transfer, ECM decompression, CUE file repair, and archive extraction.

## License

MIT
