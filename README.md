![](gato-banner.png)

# Gato OS

An elegant WIP distro that will feel like magic. Currently? just half-vibecoded slop from hell while I focus more on the general direction I want + I learn everything I need to to make this a reality.

## Installation

1. Install [Fedora COSMIC Atomic](https://fedoraproject.org/atomic-desktops/cosmic/)
2. Rebase to Gato OS:
    ```bash
    rpm-ostree rebase ostree-unverified-registry:ghcr.io/DCCXXV/gato-os:latest
    systemctl reboot
    ```

## Features

### Gato Dark Theme

A custom colorscheme with themes for COSMIC and Zed (and more to come!).

### Gato Daemon & CLI

**Intelligent Folders** - Automatically process files dropped into designated folders:

```bash
gato f a ~/Photos -k              # Compress images (keep original)
gato f a ~/Videos -a convert-mp4  # Convert to MP4
gato f a ~/Resize -c 'convert {} -resize 50% {}'  # Custom command
gato f ls                         # List configured folders
gato f rm ~/Photos                # Remove folder
```

**Available actions:** `compress`, `convert-webp`, `convert-mp4`, `convert-mp3`, `resize-50`, `resize-25`

The CLI will have a GUI too.

### Soar Integration

Uses [Soar](https://github.com/pkgforge/soar) for package management with automatic upgrades every 3 hours.

### TO BE DONE

- Auto convert files when changing the extension from anywhere
- Ghost commands (if the program is not installed run it will run it in an ephemereal container)
- First class integration with syncthing
- Small QOL changes for developers (like detecting when you are trying to run something on an used port and launching a cosmic notification telling you which program is using it and more)
- (very in the future) an android app that merges the functionality of syncthing and kdeconnect.

## Apps

- **Browser**: Helium
- **Editor**: Zed
- **Media**: VLC, Loupe
- **AppImage Manager**: Gear Lever
- **Theme**: Gato Dark (COSMIC)

## Building

```bash
bluebuild build --build-driver podman --inspect-driver podman --run-driver podman recipe.yml
```

## License

MIT
