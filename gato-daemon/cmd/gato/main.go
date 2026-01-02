package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veinticinco/gato-daemon/internal/folders"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "folder", "f":
		handleFolder(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Println("gato v0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func handleFolder(args []string) {
	if len(args) == 0 {
		// Default to list
		args = []string{"ls"}
	}

	mgr := folders.New()
	if err := mgr.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "add", "a":
		cmdAdd(mgr, args[1:])
	case "remove", "rm":
		cmdRemove(mgr, args[1:])
	case "list", "ls":
		cmdList(mgr, args[1:])
	case "-h", "--help", "help":
		printFolderHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
		printFolderHelp()
		os.Exit(1)
	}
}

// Presets for common operations
var presets = map[string]string{
	"compress":   "convert {} -strip -quality 75 {}",
	"webp":       "convert {} -quality 90 {dir}/{name}.webp && rm {}",
	"png":        "convert {} {dir}/{name}.png && rm {}",
	"jpg":        "convert {} -quality 85 {dir}/{name}.jpg && rm {}",
	"optimize":   "pngquant --force --quality=65-80 --output {} {}",
	"resize-50":  "convert {} -resize 50% {}",
	"resize-25":  "convert {} -resize 25% {}",
	"mp4":        "ffmpeg -i {} -c:v libx264 -c:a aac -y {dir}/{name}.mp4 && rm {}",
	"mp3":        "ffmpeg -i {} -c:a libmp3lame -q:a 2 -y {dir}/{name}.mp3 && rm {}",
	"gif":        "ffmpeg -i {} -vf 'fps=10,scale=480:-1' -y {dir}/{name}.gif && rm {}",
}

func cmdAdd(mgr *folders.Manager, args []string) {
	// Parse flags manually for flexibility
	var path, command, preset string
	var extensions []string
	var keepOriginal bool

	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "-p" || arg == "--preset":
			if i+1 < len(args) {
				preset = args[i+1]
				i += 2
			} else {
				fmt.Fprintln(os.Stderr, "Error: -p requires a preset name")
				os.Exit(1)
			}
		case arg == "-e" || arg == "--ext":
			if i+1 < len(args) {
				extensions = strings.Split(args[i+1], ",")
				i += 2
			} else {
				fmt.Fprintln(os.Stderr, "Error: -e requires extensions")
				os.Exit(1)
			}
		case arg == "-k" || arg == "--keep":
			keepOriginal = true
			i++
		case arg == "-h" || arg == "--help":
			printAddHelp()
			os.Exit(0)
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
			os.Exit(1)
		default:
			// Positional args: first is path, second (if any) is command
			if path == "" {
				path = arg
			} else if command == "" {
				command = arg
			}
			i++
		}
	}

	if path == "" {
		printAddHelp()
		os.Exit(1)
	}

	// Expand path
	path = expandPath(path)

	// Handle preset
	if preset != "" {
		if cmd, ok := presets[preset]; ok {
			command = cmd
		} else {
			fmt.Fprintf(os.Stderr, "Unknown preset: %s\n\n", preset)
			fmt.Println("Available presets:")
			for name := range presets {
				fmt.Printf("  %s\n", name)
			}
			os.Exit(1)
		}
	}

	if err := mgr.AddFolder(path, "", command, extensions, keepOriginal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	name := filepath.Base(path)
	if command != "" {
		display := command
		if len(display) > 50 {
			display = display[:47] + "..."
		}
		fmt.Printf("Added command to %s:\n  %s\n", name, display)
	} else {
		fmt.Printf("Added folder: %s\n", name)
	}
}

func cmdRemove(mgr *folders.Manager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("  gato f rm <path>            Remove folder entirely")
		fmt.Println("  gato f rm <path> <command>  Remove specific command")
		os.Exit(1)
	}

	path := expandPath(args[0])
	name := filepath.Base(path)

	// Check if removing specific command or entire folder
	if len(args) >= 2 {
		command := args[1]
		if err := mgr.RemoveAction(path, "", command); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		display := command
		if len(display) > 50 {
			display = display[:47] + "..."
		}
		fmt.Printf("Removed command from %s:\n  %s\n", name, display)
	} else {
		if err := mgr.RemoveFolder(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed folder: %s\n", name)
	}
}

func cmdList(mgr *folders.Manager, args []string) {
	// Check for --raw flag
	raw := false
	var specificPath string

	for _, arg := range args {
		if arg == "--raw" {
			raw = true
		} else if !strings.HasPrefix(arg, "-") {
			specificPath = expandPath(arg)
		}
	}

	paths := mgr.ListUniqueFolders()

	// If specific path provided, show only that folder's commands
	if specificPath != "" {
		actions := mgr.GetFolderActions(specificPath)
		if len(actions) == 0 {
			// Check if folder exists but has no commands
			found := false
			for _, p := range paths {
				if p == specificPath {
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "Folder not found: %s\n", specificPath)
				os.Exit(1)
			}
			fmt.Printf("%s: no commands\n", filepath.Base(specificPath))
			return
		}

		if raw {
			for _, a := range actions {
				cmd := a.Command
				if cmd == "" {
					cmd = a.Action
				}
				fmt.Printf("%s\t%s\n", specificPath, cmd)
			}
		} else {
			fmt.Printf("%s (%d commands):\n", filepath.Base(specificPath), len(actions))
			for _, a := range actions {
				cmd := a.Command
				if cmd == "" {
					cmd = a.Action
				}
				fmt.Printf("  %s\n", cmd)
			}
		}
		return
	}

	// List all folders
	if raw {
		for _, path := range paths {
			actions := mgr.GetFolderActions(path)
			if len(actions) == 0 {
				fmt.Printf("%s\t\n", path)
			}
			for _, a := range actions {
				cmd := a.Command
				if cmd == "" {
					cmd = a.Action
				}
				fmt.Printf("%s\t%s\n", path, cmd)
			}
		}
		return
	}

	if len(paths) == 0 {
		fmt.Println("No intelligent folders configured.")
		fmt.Println()
		fmt.Println("Add one:")
		fmt.Println("  gato f add ~/Photos -p compress")
		fmt.Println("  gato f add ~/Downloads \"convert {} -resize 50% {}\"")
		return
	}

	for _, path := range paths {
		actions := mgr.GetFolderActions(path)
		fmt.Printf("%s (%d)\n", filepath.Base(path), len(actions))
		for _, a := range actions {
			cmd := a.Command
			if cmd == "" {
				cmd = a.Action
			}
			display := cmd
			if len(display) > 60 {
				display = display[:57] + "..."
			}
			fmt.Printf("  %s\n", display)
		}
	}
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return path
}

func printHelp() {
	fmt.Println("gato - Intelligent utilities for Gato OS")
	fmt.Println()
	fmt.Println("Usage: gato <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  folder, f    Manage intelligent folders")
	fmt.Println("  help         Show this help")
	fmt.Println("  version      Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gato f ls                              List all folders")
	fmt.Println("  gato f add ~/Photos -p compress        Add with preset")
	fmt.Println("  gato f add ~/Downloads \"convert ...\"   Add custom command")
	fmt.Println("  gato f rm ~/Photos                     Remove folder")
}

func printFolderHelp() {
	fmt.Println("gato folder - Manage intelligent folders")
	fmt.Println()
	fmt.Println("Usage: gato f <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ls, list     List folders (default)")
	fmt.Println("  add, a       Add folder or command")
	fmt.Println("  rm, remove   Remove folder or command")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gato f ls                              List all folders")
	fmt.Println("  gato f ls ~/Photos                     Show folder's commands")
	fmt.Println("  gato f add ~/Photos                    Add empty folder")
	fmt.Println("  gato f add ~/Photos -p compress        Add folder with preset")
	fmt.Println("  gato f add ~/Photos \"convert {} ...\"   Add custom command")
	fmt.Println("  gato f rm ~/Photos                     Remove folder entirely")
	fmt.Println("  gato f rm ~/Photos \"convert {} ...\"    Remove specific command")
}

func printAddHelp() {
	fmt.Println("gato folder add - Add folder or command")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gato f add <path>                      Add empty folder")
	fmt.Println("  gato f add <path> <command>            Add command to folder")
	fmt.Println("  gato f add <path> -p <preset>          Add preset command")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --preset <name>   Use preset command")
	fmt.Println("  -e, --ext <list>      Only process these extensions (comma-separated)")
	fmt.Println("  -k, --keep            Keep originals in .originals/")
	fmt.Println()
	fmt.Println("Presets:")
	fmt.Println("  compress    Compress images (quality 75)")
	fmt.Println("  webp        Convert to WebP")
	fmt.Println("  png         Convert to PNG")
	fmt.Println("  jpg         Convert to JPG (quality 85)")
	fmt.Println("  optimize    Optimize PNG with pngquant")
	fmt.Println("  resize-50   Resize to 50%")
	fmt.Println("  resize-25   Resize to 25%")
	fmt.Println("  mp4         Convert video to MP4")
	fmt.Println("  mp3         Convert audio to MP3")
	fmt.Println("  gif         Convert video to GIF")
	fmt.Println()
	fmt.Println("Command placeholders:")
	fmt.Println("  {}          Full file path")
	fmt.Println("  {name}      Filename without extension")
	fmt.Println("  {ext}       File extension")
	fmt.Println("  {dir}       Directory path")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gato f add ~/Photos -p compress -k")
	fmt.Println("  gato f add ~/Screenshots -p webp -e png,jpg")
	fmt.Println("  gato f add ~/Videos \"ffmpeg -i {} -crf 28 {dir}/{name}_small.mp4\"")
}
