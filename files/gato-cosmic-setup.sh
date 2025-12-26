#!/bin/bash
# Gato COSMIC Setup - Runs once on first login

MARKER="$HOME/.config/gato-cosmic-setup-done"

# Checks again and exits if already run
if [ -f "$MARKER" ]; then
    exit 0
fi

# Import Gato Dark theme
if [ -f /usr/share/cosmic/themes/gato-dark.ron ] && [ -x /usr/bin/cosmic-theme-import ]; then
    /usr/bin/cosmic-theme-import /usr/share/cosmic/themes/gato-dark.ron
fi

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
