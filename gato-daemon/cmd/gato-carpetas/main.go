package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/veinticinco/gato-daemon/internal/folders"
)

var (
	bg      = color.NRGBA{R: 5, G: 19, B: 21, A: 255}
	surface = color.NRGBA{R: 12, G: 30, B: 34, A: 255}
	accent  = color.NRGBA{R: 255, G: 122, B: 122, A: 255}
	text    = color.NRGBA{R: 185, G: 190, B: 190, A: 255}
	dim     = color.NRGBA{R: 80, G: 90, B: 92, A: 255}
	border  = color.NRGBA{R: 25, G: 45, B: 50, A: 255}
)

// Preset commands
var presetOrder = []string{"Compress", "WebP", "MP4", "MP3", "Resize 50%", "Resize 25%", "PNG", "JPG", "Optimize PNG", "GIF"}
var presets = map[string]string{
	"Compress":     "convert {} -strip -quality 75 {}",
	"WebP":         "convert {} {dir}/{name}.webp && rm {}",
	"PNG":          "convert {} {dir}/{name}.png && rm {}",
	"JPG":          "convert {} -quality 85 {dir}/{name}.jpg && rm {}",
	"Optimize PNG": "pngquant --force --quality=65-80 --output {} {}",
	"Resize 50%":   "convert {} -resize 50% {}",
	"Resize 25%":   "convert {} -resize 25% {}",
	"MP4":          "ffmpeg -i {} -c:v libx264 -c:a aac -y {dir}/{name}.mp4 && rm {}",
	"MP3":          "ffmpeg -i {} -c:a libmp3lame -q:a 2 -y {dir}/{name}.mp3 && rm {}",
	"GIF":          "ffmpeg -i {} -vf 'fps=10,scale=480:-1' -y {dir}/{name}.gif && rm {}",
}

type gatoTheme struct{}

func (g *gatoTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return bg
	case theme.ColorNameButton:
		return surface
	case theme.ColorNameDisabled:
		return dim
	case theme.ColorNameDisabledButton:
		return surface
	case theme.ColorNameFocus:
		return accent
	case theme.ColorNameForeground:
		return text
	case theme.ColorNameForegroundOnPrimary:
		return bg
	case theme.ColorNameHover:
		return color.NRGBA{R: 20, G: 42, B: 48, A: 255}
	case theme.ColorNameInputBackground:
		return surface
	case theme.ColorNameInputBorder:
		return border
	case theme.ColorNamePlaceHolder:
		return dim
	case theme.ColorNamePressed:
		return accent
	case theme.ColorNamePrimary:
		return accent
	case theme.ColorNameScrollBar:
		return border
	case theme.ColorNameSelection:
		return color.NRGBA{R: 255, G: 122, B: 122, A: 30}
	case theme.ColorNameSeparator:
		return border
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 30}
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return surface
	default:
		return theme.DefaultTheme().Color(n, v)
	}
}

func (g *gatoTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (g *gatoTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

func (g *gatoTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameScrollBar:
		return 3
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 0
	case theme.SizeNameSelectionRadius:
		return 0
	default:
		return theme.DefaultTheme().Size(n)
	}
}

func main() {
	a := app.New()
	a.Settings().SetTheme(&gatoTheme{})
	w := a.NewWindow("Gato Carpetas")

	mgr := folders.New()
	mgr.LoadConfig()

	// Get folder path from args
	folderPath := ""
	if len(os.Args) > 1 {
		folderPath = os.Args[1]
		if strings.HasPrefix(folderPath, "~/") {
			home, _ := os.UserHomeDir()
			folderPath = filepath.Join(home, folderPath[2:])
		}
		// Always convert to absolute path
		if abs, err := filepath.Abs(folderPath); err == nil {
			folderPath = abs
		}
	}

	// Main content holder
	mainContent := container.NewStack()

	var showFolderList func()
	var showFolderConfig func(path string)

	// === FOLDER LIST VIEW ===
	showFolderList = func() {
		paths := mgr.ListUniqueFolders()

		if len(paths) == 0 {
			empty := canvas.NewText("No folders configured", dim)
			empty.TextSize = 13
			mainContent.Objects = []fyne.CanvasObject{container.NewCenter(empty)}
			mainContent.Refresh()
			return
		}

		var items []fyne.CanvasObject
		for _, path := range paths {
			path := path
			actions := mgr.GetFolderActions(path)

			name := filepath.Base(path)
			nameText := canvas.NewText(name, text)
			nameText.TextSize = 13
			nameText.TextStyle = fyne.TextStyle{Bold: true}

			pathText := canvas.NewText(path, dim)
			pathText.TextSize = 10

			countText := canvas.NewText(pluralize(len(actions), "command", "commands"), dim)
			countText.TextSize = 10

			openBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
				exec.Command("xdg-open", path).Start()
			})
			openBtn.Importance = widget.LowImportance

			row := widget.NewButton("", func() {
				showFolderConfig(path)
			})
			row.Importance = widget.LowImportance

			item := container.NewStack(
				row,
				container.NewBorder(nil, nil, nil, openBtn,
					container.NewPadded(container.NewVBox(nameText, pathText, countText)),
				),
			)
			items = append(items, item)
		}

		list := container.NewVBox(items...)
		scroll := container.NewVScroll(list)

		mainContent.Objects = []fyne.CanvasObject{scroll}
		mainContent.Refresh()
	}

	// === FOLDER CONFIG VIEW ===
	showFolderConfig = func(path string) {
		actions := mgr.GetFolderActions(path)

		// Header
		folderName := filepath.Base(path)
		nameText := canvas.NewText(folderName, text)
		nameText.TextSize = 15
		nameText.TextStyle = fyne.TextStyle{Bold: true}

		pathText := canvas.NewText(path, dim)
		pathText.TextSize = 11

		backBtn := widget.NewButton(" ← ", func() {
			showFolderList()
		})
		backBtn.Importance = widget.LowImportance

		header := container.NewBorder(nil, nil, backBtn, nil,
			container.NewVBox(nameText, pathText),
		)

		// Current commands section
		var commandItems []fyne.CanvasObject
		if len(actions) == 0 {
			empty := canvas.NewText("No commands", dim)
			empty.TextSize = 12
			commandItems = append(commandItems, empty)
		} else {
			for _, action := range actions {
				action := action
				cmdText := action.Command
				if cmdText == "" {
					cmdText = action.Action // fallback for old presets
				}
				// Truncate if too long
				display := cmdText
				if len(display) > 70 {
					display = display[:67] + "..."
				}

				label := canvas.NewText(display, text)
				label.TextSize = 11
				label.TextStyle = fyne.TextStyle{Monospace: true}

				removeBtn := widget.NewButton(" × ", func() {
					mgr.RemoveAction(path, action.Action, action.Command)
					showFolderConfig(path)
				})

				row := container.NewBorder(nil, nil, nil, removeBtn, label)
				commandItems = append(commandItems, row)
			}
		}

		commandsBox := container.NewVBox(commandItems...)

		// Add command section
		addLabel := canvas.NewText("Add command", dim)
		addLabel.TextSize = 11

		hint := canvas.NewText("{} = file, {dir} = folder, {name} = name, {ext} = ext", dim)
		hint.TextSize = 10

		cmdEntry := widget.NewEntry()
		cmdEntry.SetPlaceHolder("convert {} -resize 50% {}")

		// Presets as small buttons in a grid
		presetsLabel := canvas.NewText("Presets", dim)
		presetsLabel.TextSize = 10

		var presetBtns []fyne.CanvasObject
		for _, name := range presetOrder {
			name := name
			cmd := presets[name]
			btn := widget.NewButton(name, func() {
				cmdEntry.SetText(cmd)
			})
			btn.Importance = widget.LowImportance
			presetBtns = append(presetBtns, btn)
		}
		presetsGrid := container.NewGridWithColumns(5, presetBtns...)

		// Bottom buttons
		addBtn := widget.NewButton("Add", func() {
			cmd := strings.TrimSpace(cmdEntry.Text)
			if cmd == "" {
				return
			}
			mgr.AddFolder(path, "", cmd, nil, false)
			cmdEntry.SetText("")
			showFolderConfig(path)
		})
		addBtn.Importance = widget.HighImportance

		removeBtn := widget.NewButton("Remove folder", func() {
			mgr.RemoveFolder(path)
			showFolderList()
		})

		// Layout
		sep1 := canvas.NewLine(border)
		sep1.StrokeWidth = 1
		sep2 := canvas.NewLine(border)
		sep2.StrokeWidth = 1

		bottomRow := container.NewGridWithColumns(2, removeBtn, addBtn)

		content := container.NewBorder(
			container.NewVBox(header, sep1),
			container.NewVBox(sep2, bottomRow),
			nil, nil,
			container.NewVBox(
				commandsBox,
				layout.NewSpacer(),
				addLabel,
				hint,
				cmdEntry,
				presetsLabel,
				presetsGrid,
			),
		)

		mainContent.Objects = []fyne.CanvasObject{content}
		mainContent.Refresh()
	}

	// Initial view
	if folderPath != "" {
		showFolderConfig(folderPath)
	} else {
		showFolderList()
	}

	w.SetContent(container.NewPadded(mainContent))
	w.Resize(fyne.NewSize(360, 340))
	w.ShowAndRun()
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", n, plural)
}
