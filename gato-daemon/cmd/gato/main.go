package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
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
		printFolderHelp()
		return
	}

	mgr := folders.New()
	if err := mgr.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "add", "a":
		cmdAdd(mgr, args[1:])
	case "remove", "rm", "r":
		cmdRemove(mgr, args[1:])
	case "list", "ls", "l":
		cmdList(mgr)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
		printFolderHelp()
		os.Exit(1)
	}
}

func cmdAdd(mgr *folders.Manager, args []string) {
	fs := pflag.NewFlagSet("add", pflag.ExitOnError)

	action := fs.StringP("action", "a", "", "Predefined action (compress, convert-webp, convert-mp4, convert-mp3, resize-50, resize-25)")
	cmd := fs.StringP("cmd", "c", "", "Custom command ({} = filepath, {name}, {ext}, {dir})")
	ext := fs.StringP("ext", "e", "", "Only process these extensions (comma-separated)")
	keep := fs.BoolP("keep-original", "k", false, "Keep originals in .originals/")

	fs.Usage = func() {
		fmt.Println("Usage: gato folder add <path> [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Actions:")
		fmt.Println("  compress      Compress images (PNG, JPG, WebP)")
		fmt.Println("  convert-webp  Convert images to WebP")
		fmt.Println("  convert-mp4   Convert videos to MP4")
		fmt.Println("  convert-mp3   Convert audio to MP3")
		fmt.Println("  resize-50     Resize images to 50%")
		fmt.Println("  resize-25     Resize images to 25%")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gato f a ~/Photos -a compress -k")
		fmt.Println("  gato folder add ~/Videos --action convert-mp4")
		fmt.Println("  gato f a ~/Custom -c 'convert {} -resize 800x600 {}'")
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	path := fs.Arg(0)

	// Default to compress if no action specified
	if *action == "" && *cmd == "" {
		*action = "compress"
	}

	var extensions []string
	if *ext != "" {
		extensions = strings.Split(*ext, ",")
	}

	if err := mgr.AddFolder(path, *action, *cmd, extensions, *keep); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added: %s\n", path)
	if *cmd != "" {
		fmt.Printf("  Command: %s\n", *cmd)
	} else {
		fmt.Printf("  Action: %s\n", *action)
	}
	if len(extensions) > 0 {
		fmt.Printf("  Extensions: %s\n", strings.Join(extensions, ", "))
	}
	if *keep {
		fmt.Printf("  Keep originals: yes\n")
	}
}

func cmdRemove(mgr *folders.Manager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gato folder remove <path>")
		os.Exit(1)
	}

	if err := mgr.RemoveFolder(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Removed: %s\n", args[0])
}

func cmdList(mgr *folders.Manager) {
	list := mgr.ListFolders()
	if len(list) == 0 {
		fmt.Println("No intelligent folders configured.")
		fmt.Println()
		fmt.Println("Add one:")
		fmt.Println("  gato folder add ~/Photos --action compress")
		return
	}

	fmt.Println("Intelligent folders:")
	fmt.Println()
	for _, f := range list {
		fmt.Printf("  %s\n", f.Path)
		if f.Command != "" {
			fmt.Printf("    cmd: %s\n", f.Command)
		} else {
			fmt.Printf("    action: %s\n", f.Action)
		}
		if len(f.Extensions) > 0 {
			fmt.Printf("    ext: %s\n", strings.Join(f.Extensions, ","))
		}
		if f.KeepOriginal {
			fmt.Printf("    keep-original: yes\n")
		}
	}
}

func printHelp() {
	fmt.Println("gato - Gato OS intelligent utilities")
	fmt.Println()
	fmt.Println("Usage: gato <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  folder, f    Manage intelligent folders")
	fmt.Println("  help         Show this help")
	fmt.Println("  version      Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gato f a ~/Photos -a compress -k")
	fmt.Println("  gato folder list")
}

func printFolderHelp() {
	fmt.Println("gato folder - Manage intelligent folders")
	fmt.Println()
	fmt.Println("Usage: gato folder <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add, a       Add intelligent folder")
	fmt.Println("  remove, rm   Remove intelligent folder")
	fmt.Println("  list, ls     List all folders")
	fmt.Println()
	fmt.Println("Run 'gato folder add --help' for options.")
}
