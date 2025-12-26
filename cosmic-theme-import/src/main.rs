use cosmic_config::CosmicConfigEntry;
use cosmic_theme::ThemeBuilder;
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

    // Get the dark theme config (gato-dark is always dark)
    let config = ThemeBuilder::dark_config()?;

    // Write the theme using the same method cosmic-settings uses
    theme_builder.write_entry(&config)?;

    println!("Successfully imported dark theme from {}", theme_path);

    Ok(())
}
