#!/bin/bash
# Gato COSMIC Setup - Runs once on first login

MARKER="$HOME/.config/gato-cosmic-setup-done"

# Checks again and exits if already run
if [ -f "$MARKER" ]; then
    exit 0
fi

# Set up COSMIC theme
mkdir -p ~/.local/share/cosmic/themes
if [ -f /usr/share/cosmic/themes/gato-dark.ron ]; then
    cp /usr/share/cosmic/themes/gato-dark.ron ~/.local/share/cosmic/themes/
fi

mkdir -p ~/.config/cosmic/com.system76.CosmicTheme/v1/
echo '(active: "gato-dark", mode: Dark)' > ~/.config/cosmic/com.system76.CosmicTheme/v1/theme

# Set up COSMIC wallpaper
mkdir -p ~/.config/cosmic/com.system76.CosmicBackground/v1/
cat > ~/.config/cosmic/com.system76.CosmicBackground/v1/all <<'EOF'
(
    output: "all",
    source: Path("/usr/share/backgrounds/gato/gato-wp2.png"),
    filter_by_theme: true,
    rotation_frequency: 300,
    filter_method: Lanczos,
    scaling_mode: Zoom,
    sampling_method: Alphanumeric,
)
EOF

# Mark as complete
touch "$MARKER"
