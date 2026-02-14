# ReplayOS Technical Reference

## Overview

ReplayOS is a lightweight, highly optimized Linux distribution built around a libretro-based frontend (RetroArch cores). It is specifically designed for accurate, low-latency emulation of classic consoles, arcade systems, and retro computers on **Raspberry Pi** hardware. The system emphasizes performance, minimal input lag, and seamless support for both **LCD** and **CRT** displays.

ReplayOS is **not** a general-purpose operating system. It is restricted to supported Raspberry Pi models and does not support PC/x86 hardware, handheld devices (Anbernic, Powkiddy, etc.), or non-Raspberry Pi platforms.

## Hardware Compatibility

- **Supported devices**: Raspberry Pi Zero 2, 3A/3B/3B+, 4B, 5B, Pi 500, Compute Module 5 (performance varies by model)
- **Minimum RAM**: 1 GB (performance is CPU/GPU-bound)
- **Storage**:
  - MicroSD card for OS boot (A2 U3 recommended, 16 GB+, official Raspberry Pi cards preferred)
  - Optional: USB 3.0 drive, NVMe (Pi 5), or NFS share for ROMs, BIOS, saves, and config
- **Display**:
  - LCD: HDMI up to 1920×1080@60 (experimental 2560×1440 on Pi 5)
  - CRT: 15/25/31 kHz, progressive/interlaced, arcade/consumer monitors
  - Dual-screen modes (cloned, side-by-side, stacked)
  - Limited DPI (GPIO) support on Pi 5 (legacy RGB-Pi adapter, no audio)
- **Input**: >700 controllers supported natively, real GunCon2 lightgun (CRT only), six-player support
- **Audio**: Low-latency resampler (~32 ms), normalization, external USB/GPIO DAC support

## Core Features

- Ultra-low input lag (0–1 frame default, no runahead required)
- **DynaRes 2.0** engine: dynamic timings, Calamity Modeline Calculator, interlaced flicker reduction, CRT profiles, CSYNC modes
- Automatic core and settings adaptation based on hardware, display type, and emulated system
- Frontend enhancements:
  - Adaptive integer scaling
  - UI rotation (90/180/270°)
  - Favorites & Recent lists
  - Arcade game database with proper naming
  - Advanced filtering (players, buttons, orientation, system type)
  - Kiosk mode
  - Autostart games on boot
  - Coin-op timer mode
  - AmbiScan dynamic colored borders
- Additional capabilities:
  - Virtual disk engine
  - Alpha Player (video/audio media playback)
  - Halt state (H key) for CRT photography
  - Gamma, RGB, and XY position correction
  - Full/limited RGB range support for DACs

## Supported Systems and File Formats

ReplayOS includes a fixed set of optimized libretro cores. Custom core additions are not supported.

**ZIP support** (no extraction required; ZIP is the only permitted compression format for these systems):

- Arcade (FBNeo)
- Arcade (MAME)
- Arcade (MAME 2K3+)
- Arcade SEGA Naomi/Atomis
- SNK NEO-GEO
- IBM PC (MS-DOS)

**IBM PC (MS-DOS)** additionally supports: `.dosz`, `.exe`, `.com`, `.bat`, `.iso`, `.cue`, `.img`, `.m3u`, `.m3u8`

**Complete list of supported systems and file extensions**:

| System                          | Supported File Formats                                                            |
|---------------------------------|-----------------------------------------------------------------------------------|
| Arcade (FBNeo)                  | zip                                                                               |
| Arcade (MAME)                   | zip                                                                               |
| Arcade (MAME 2K3+)              | zip                                                                               |
| Arcade SEGA Naomi/Atomis        | zip                                                                               |
| Atari 2600/VCS                  | a26, bin                                                                          |
| Atari 5200                      | a52, bin                                                                          |
| Atari 7800                      | a78, bin, cdf                                                                     |
| Atari Jaguar                    | j64, jag                                                                          |
| Atari Lynx                      | lnx                                                                               |
| NEC TurboGrafx-16 / PC Engine   | pce, sgx, toc                                                                     |
| NEC TurboGrafx-CD / CD-ROM²     | cue, ccd, chd, m3u                                                                |
| Nintendo NES / Famicom          | fds, nes, unf, unif                                                               |
| Nintendo SNES / Super Famicom   | smc, sfc, swc, fig, bs, st                                                        |
| Nintendo 64                     | n64, v64, z64, bin, u1                                                            |
| Nintendo Game Boy               | gb, sgb                                                                           |
| Nintendo Game Boy Color         | gbc, sgbc                                                                         |
| Nintendo Game Boy Advance       | gba                                                                               |
| Nintendo DS                     | nds                                                                               |
| SEGA SG-1000                    | sg                                                                                |
| SEGA Game Gear                  | gg                                                                                |
| SEGA Master System / Mark III   | sms                                                                               |
| SEGA Megadrive / Genesis        | md, smd, gen, bin                                                                 |
| SEGA Mega-CD / Sega CD          | m3u, cue, iso, chd                                                                |
| SEGA 32X                        | 32x                                                                               |
| SEGA Saturn                     | cue, ccd, chd, toc, m3u                                                           |
| SEGA Dreamcast                  | chd, cdi, elf, cue, gdi, lst, dat, m3u                                            |
| SNK NEO-GEO                     | zip                                                                               |
| SNK NEO-GEO CD                  | cue, chd                                                                          |
| SNK NEO-GEO Pocket              | ngp, ngc, ngpc, npc                                                               |
| SONY PlayStation                | exe, psexe, cue, img, iso, chd, pbp, ecm, mds, psf, m3u                           |
| Panasonic 3DO                   | iso, chd, cue                                                                     |
| Philips CD-i                    | iso, chd, cue                                                                     |
| Amstrad CPC                     | dsk, sna, tap, cdt, voc, cpr, m3u                                                 |
| Commodore 64                    | d64, d71, d80, d81, d82, g64, g41, x64, t64, tap, prg, p00, crt, bin, gz, …, m3u  |
| Commodore Amiga                 | adf, adz, dms, fdi, raw, hdf, hdz, lha, slave, info, uae, m3u                     |
| Commodore Amiga CD32            | cue, ccd, nrg, mds, iso, chd, m3u                                                 |
| Sharp X68000                    | dim, img, d88, 88d, hdm, dup, 2hd, xdf, hdf, cmd, m3u                             |
| Microsoft MSX                   | rom, ri, mx1, mx2, dsk, col, sg, sc, sf, cas, m3u                                 |
| Sinclair ZX Spectrum            | tzx, tap, z80, rzx, scl, trd, dsk, dck, sna, szx                                  |
| IBM PC (MS-DOS)                 | zip, dosz, exe, com, bat, iso, cue, img, m3u, m3u8                                |
| ScummVM                         | scummvm, svm                                                                      |
| Alpha Player (media)            | mkv, avi, f4v, mp4, mp3, flac, ogg, wav, … (wide range of video/audio formats)    |

**Note**: Files must match the listed formats for each system. If a format is not supported natively, use ZIP compression only for systems that explicitly allow it.

## Folder Structure

Here is the exact folder structure of the ReplayOS environment:
```
├── bios
├── captures
├── config
│   ├── input
│   │   ├── game
│   │   │   ├── crt
│   │   │   └── lcd
│   │   └── system
│   │       ├── crt
│   │       └── lcd
│   └── settings
│       ├── game
│       │   ├── crt
│       │   └── lcd
│       └── system
│           ├── crt
│           └── lcd
├── roms
│   ├── amstrad_cpc
│   ├── arcade_dc
│   ├── arcade_fbneo
│   ├── arcade_mame
│   ├── arcade_mame_2k3p
│   ├── atari_2600
│   ├── atari_5200
│   ├── atari_7800
│   ├── atari_jaguar
│   ├── atari_lynx
│   ├── _autostart
│   ├── commodore_amiga
│   ├── commodore_amigacd
│   ├── commodore_c64
│   ├── _extra
│   ├── _favorites
│   ├── ibm_pc
│   ├── media_player
│   ├── microsoft_msx
│   ├── nec_pce
│   ├── nec_pcecd
│   ├── nintendo_ds
│   ├── nintendo_gb
│   ├── nintendo_gba
│   ├── nintendo_n64
│   ├── nintendo_nes
│   ├── nintendo_snes
│   ├── panasonic_3do
│   ├── philips_cdi
│   ├── _recent
│   ├── scummvm
│   ├── sega_32x
│   ├── sega_cd
│   ├── sega_dc
│   ├── sega_gg
│   ├── sega_sg
│   ├── sega_smd
│   ├── sega_sms
│   ├── sega_st
│   ├── sharp_x68k
│   ├── sinclair_zx
│   ├── snk_ng
│   ├── snk_ngcd
│   ├── snk_ngp
│   └── sony_psx
└── saves
```
### Directory Purposes

- `bios` — BIOS files, arcade samples, sound fonts, special configurations
- `captures` — Screenshots and captures
- `config` — Input mappings and core settings (split by game/system and CRT/LCD)
- `roms` — Game ROMs and disc images (subfolders correspond to supported systems)
- `saves` — Save states and in-game save data

Special-purpose subfolders under `roms`:
- `_autostart` — Games to launch automatically on boot
- `_extra` — Additional content
- `_favorites` — User-marked favorites
- `_recent` — Recently played items

### RePlay Options File

The following is a description of all available options and default values that RePlay uses for global configuration.

The configuration file is located in '/media/sd/config/replay.cfg':
```
# video_connector
## 0 = hdmi
## 1 = dpi (used for gpio)
video_connector             = "0"
# video_mode
## NRR (Native Refresh Rate)
## 0 = default
## 1 = crt 320x240@nrr (ui boots @60)
## 2 = crt 320x240@nrr (ui boots @50)
## 3 = lcd native resolution & nrr
## 4 = lcd 1920x1080@60
## 5 = lcd 1280x720@60
## 6 = lcd 1280x1024@60
## 7 = lcd 1024x768@60
## 8 = lcd 2560x1440@60
## 9 = lcd 3840x2160@60
video_mode                  = "0"
# video_monitor_multi_mode
## 0 = disabled
## 1 = dual cloned
## 2 = dual horizontal
## 3 = dual vertical
## 4 = dual smart output
video_monitor_multi_mode     = "0"
# video_lcd_type
## generic_60 = supports 55-61hz ranges
## gaming_vrr = supports 48-75hz ranges
video_lcd_type              = "generic_60"
# video_crt_type
## generic_15
## arcade_15
## arcade_15_25
## arcade_15_25_31
## arcade_31 (also used for PC)
video_crt_type              = "generic_15"
# video_crt_csync_mode (requires RGB-Pi compatible hardware)
## 0 = AND
## 1 = XOR
## 2 = separated H/V
video_crt_csync_mode        = "0"
# video_crt_rgb_range
## 0 = auto
## 1 = full (0:255)
## 2 = limited (16:235)
video_crt_rgb_range         = "0"
# video_aspect_ratio
## 0 = full screen 4:3
## 1 = full screen native
## 2 = vertical integer scaling, horizontal 4:3
## 3 = vertical integer scaling, horizontal native
## 4 = horizontal integer scaling, vertical 4:3
## 5 = horizontal integer scaling, vertical native
## 6 = full integer scaling
## 7 = full integer over scaling (only FHD TVs)
## 8 = full integer under scaling
video_aspect_ratio          = "0"
# video_crt_h_shift
## values = -16<-->16
video_crt_h_shift           = "0"
# video_crt_h_size
## values = 0.5<-->1.5
video_crt_h_size            = "1.0"
video_monitor_x             = "0"
video_monitor_y             = "0"
# video_gamma
## values = 0.5<-->1.5
video_gamma                 = "1.0"
# video_red_scale
## values = 0.0<-->1.0
video_red_scale             = "1.0"
# video_green_scale
## values = 0.0<-->1.0
video_green_scale           = "1.0"
# video_blue_scale
## values = 0.0<-->1.0
video_blue_scale            = "1.0"
# video_ui_rotation_mode
## 0 = 0
## 1 = 90
## 2 = 180
## 3 = 270
video_ui_rotation_mode      = "0"
video_show_fps              = "false"
video_show_info             = "false"
# video_filter
## 0 = none
## 1 = light scanlines
## 2 = medium scanlines
## 3 = strong scanlines
## 4 = black scanlines          
video_filter                = "0"
video_ambiscan              = "true"
# video_screen_saver
## 0 = OFF
## 60000 = 1 min
## 180000 = 3 min
## 300000 = 5 min
## 600000 = 10 min
## 900000 = 15 min
video_screen_saver          = "0"
# audio_card
## 0 = HDMI
## 1 = USB DAC
## 2 = GPIO DAC
video_hdmi_cec              = "false"
audio_card                  = "0"
audio_mono                  = "false"
audio_normalization         = "false"
# audio_system_volume
## values = 0<-->10
audio_system_volume         = "10"
# input_gcon2_flash
## 0 = disabled
## 1 = pulse
## 2 = hold
input_gcon2_flash           = "1"
input_gcon2_offscreen       = "true"
input_ui_swap_ab            = "false"
input_all_control_ui        = "false"
# input_ui_menu_btn
## 0 = home button
## 1 = select+start
## 2 = hold start
input_ui_menu_btn           = "1"
input_ui_select_fav         = "false"
# input_kbd_real_mode
## true = keyboard works in native scancode mode
## false = keyboard works in special cmd event mode
input_kbd_real_mode         = "true"
# input_kbd_menu_key
## 0 = windows (left)
## 1 = windows (right)
## 2 = play/pause
## 3 = home page
## 4 = home
input_kbd_menu_key          = "0"
system_coinop               = "false"
# system_coinop_time
## game time you get for a credit
system_coinop_time          = "180"
# system_verbose
## 0 = debug (not available for users)
## 1 = info
## 2 = warn
## 3 = error
## 4 = disabled
system_verbose              = "4"
# timezone_srv
## timezone detection server URL
timezone_srv                = "https://time.now/developer/api/ip"
system_kiosk_mode           = "false"
# system_low_latency_mode
## true = -1/0 frames input lag
## false = 0/1 frames input lag
system_low_latency_mode     = "false"
# system_skin
## 0 = replay (default)
## 1 = mega tech
## 2 = play choice
## 3 = astro
## 4 = super video
## 5 = mvs
## 6 = rpg
## 7 = fantasy
## 8 = simple purple
## 9 = metal
## 10 = unicolors
## 11-36 = for custom user skins
system_skin                 = "0"
# system_boot_to_system
## all
## arcade_fbneo
## arcade_mame
## arcade_mame_2k3p
## arcade_dc
## nintendo_nes
## nintendo_snes
## nintendo_gb
## sega_smd
## sony_psx
system_boot_to_system       = "all"
# system_storage
## sd = internal sd card
## usb = external usb drive
## nfs = network nfs share
system_storage              = "sd"
system_ui_pauses_core       = "false"
system_folder_regen         = "true"
# view_players
## 0 = show all
## 1-6 = num players
view_players                = "0"
# view_rotation
## 0 = show all
## 1 = horizontal
## 2 = vertical
view_rotation               = "0"
# view_displays
## 0 = show all
## 1 = single screen
## 2 = dual screen
view_displays               = "0"
# view_buttons
## 0 = show all
## 1-6 = N or less buttons
view_buttons                = "0"
# view_controller
## 0 = show all
## 1 = joystick (any)
## 2 = joystick (4-way)
## 3 = joystick (8-way)
## 4 = dial / paddle
## 5 = trackball / mouse
## 6 = lightgun
view_controller             = "0"
view_player                 = "true"
view_arcade                 = "true"
view_console                = "true"
view_computer               = "true"
view_handheld               = "true"
nfs_server                  = "192.168.X.X"
nfs_share                   = "/export/share"
# nfs_version
## 3 = NFSv3 (rpcbind/mountd required on server)
## 4 = NFSv4
nfs_version                 = "4"
wifi_name                   = "MyWifi"
wifi_pwd                    = "********"
wifi_country                = "ES"
# wifi_mode
## wpa2
## wpa3
## transition (for mixed wpa2 & wpa3)
wifi_mode                   = "transition"
wifi_hidden                 = "false"
# addon_retroflag_case_pi5
## 0 = disabled
## 1 = reset button for reboot
## 2 = reset button for menu
addon_retroflag_case_pi5    = "0"
# addon_tilt_input_pi5
## 0 = disabled
## 1 = +90
## 3 = +270
addon_tilt_input_pi5        = "0"
addon_gpio_joy_pi5          = "0"
addon_dpi_dac_pi5           = "0"
```

## Limitations

- Restricted to Raspberry Pi hardware only
- Fixed set of included libretro cores (no custom core support)
- No runahead feature (low latency achieved through other optimizations)
- Performance varies significantly by Raspberry Pi model
- LCD resolution capped at 1920×1080 (except experimental higher modes on Pi 5)
- DPI (GPIO) video output is limited (Pi 5 only, no audio)
- BIOS files required for many systems and must be correctly placed
- NFS share support requires manual configuration in `replay.cfg`

This document serves as a complete reference for ReplayOS behavior and structure, particularly useful when developing tools, launchers, managers, or configuration utilities that interact with the filesystem.
