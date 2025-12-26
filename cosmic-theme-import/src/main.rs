use cosmic_config::CosmicConfigEntry;
use cosmic_theme::{Theme, ThemeBuilder};
use std::env;
use std::fs;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();

    if args.len() != 2 {
        eprintln!("Usage: cosmic-theme-import <theme.ron>");
        std::process::exit(1);
    }

    let theme_path = &args[1];

    // Read the .ron file
    let theme_content = fs::read_to_string(theme_path)?;

    // Parse into ThemeBuilder
    let theme_builder: ThemeBuilder = ron::de::from_str(&theme_content)?;

    // Get the dark theme configs (both builder and theme)
    let builder_config = ThemeBuilder::dark_config()?;
    let theme_config = Theme::dark_config()?;

    // Write the ThemeBuilder (same as cosmic-settings apply_builder)
    theme_builder.write_entry(&builder_config)?;

    // Build and write the actual Theme (same as cosmic-settings apply_theme)
    let theme = theme_builder.build();
    theme.write_entry(&theme_config)?;

    println!("Successfully imported dark theme from {}", theme_path);

    Ok(())
}
