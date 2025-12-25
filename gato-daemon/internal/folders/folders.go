package folders

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

// FolderAction defines what happens when a file is added to a folder
type FolderAction struct {
	Path         string   `toml:"path"`
	Action       string   `toml:"action"`     // predefined: compress, convert-mp4, convert-webp, etc.
	Command      string   `toml:"command"`    // custom command, {} = filename
	Extensions   []string `toml:"extensions"` // only process these extensions (empty = all)
	Notify       bool     `toml:"notify"`
	KeepOriginal bool     `toml:"keep_original"`
}

// Config holds all folder configurations
type Config struct {
	Folders []FolderAction `toml:"folders"`
}

// Manager handles intelligent folders
type Manager struct {
	configPath string
	config     Config
	watchers   map[string]*fsnotify.Watcher
}

// New creates a new folder manager
func New() *Manager {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "gato", "folders.toml")

	return &Manager{
		configPath: configPath,
		watchers:   make(map[string]*fsnotify.Watcher),
	}
}

// LoadConfig reads the configuration file
func (m *Manager) LoadConfig() error {
	// Create config dir if needed
	os.MkdirAll(filepath.Dir(m.configPath), 0755)

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			m.config = Config{Folders: []FolderAction{}}
			return m.SaveConfig()
		}
		return err
	}

	return toml.Unmarshal(data, &m.config)
}

// SaveConfig writes the configuration file
func (m *Manager) SaveConfig() error {
	data, err := toml.Marshal(m.config)
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

// AddFolder adds a new intelligent folder
func (m *Manager) AddFolder(path, action, command string, extensions []string, keepOriginal bool) error {
	// Expand path
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}

	// Create folder if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// Check if already exists
	for i, f := range m.config.Folders {
		if f.Path == path {
			// Update existing
			m.config.Folders[i] = FolderAction{
				Path:         path,
				Action:       action,
				Command:      command,
				Extensions:   extensions,
				Notify:       true,
				KeepOriginal: keepOriginal,
			}
			return m.SaveConfig()
		}
	}

	// Add new
	m.config.Folders = append(m.config.Folders, FolderAction{
		Path:         path,
		Action:       action,
		Command:      command,
		Extensions:   extensions,
		Notify:       true,
		KeepOriginal: keepOriginal,
	})

	return m.SaveConfig()
}

// RemoveFolder removes an intelligent folder
func (m *Manager) RemoveFolder(path string) error {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}

	for i, f := range m.config.Folders {
		if f.Path == path {
			m.config.Folders = append(m.config.Folders[:i], m.config.Folders[i+1:]...)
			return m.SaveConfig()
		}
	}
	return fmt.Errorf("folder not found: %s", path)
}

// ListFolders returns all configured folders
func (m *Manager) ListFolders() []FolderAction {
	return m.config.Folders
}

// Start begins watching all configured folders
func (m *Manager) Start(ctx context.Context) error {
	if err := m.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(m.config.Folders) == 0 {
		log.Println("No intelligent folders configured. Use 'gato folder add' to add one.")
		<-ctx.Done()
		return nil
	}

	for _, folder := range m.config.Folders {
		if err := m.watchFolder(ctx, folder); err != nil {
			log.Printf("Warning: failed to watch %s: %v", folder.Path, err)
			continue
		}
		log.Printf("Watching: %s -> %s", folder.Path, m.describeAction(folder))
	}

	<-ctx.Done()

	// Cleanup
	for _, w := range m.watchers {
		w.Close()
	}
	return nil
}

func (m *Manager) describeAction(f FolderAction) string {
	if f.Command != "" {
		return fmt.Sprintf("custom: %s", f.Command)
	}
	return f.Action
}

func (m *Manager) watchFolder(ctx context.Context, folder FolderAction) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(folder.Path); err != nil {
		watcher.Close()
		return err
	}

	m.watchers[folder.Path] = watcher

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Wait for file to be fully written
					time.Sleep(500 * time.Millisecond)
					m.processFile(event.Name, folder)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watch error: %v", err)
			}
		}
	}()

	return nil
}

func (m *Manager) processFile(filePath string, folder FolderAction) {
	// Skip directories
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return
	}

	// Skip hidden files and .originals
	base := filepath.Base(filePath)
	if strings.HasPrefix(base, ".") {
		return
	}

	// Check extension filter
	if len(folder.Extensions) > 0 {
		ext := strings.ToLower(filepath.Ext(filePath))
		matched := false
		for _, e := range folder.Extensions {
			if ext == "."+strings.TrimPrefix(e, ".") {
				matched = true
				break
			}
		}
		if !matched {
			return
		}
	}

	log.Printf("Processing: %s", filePath)

	// Backup original if requested
	if folder.KeepOriginal {
		backupDir := filepath.Join(folder.Path, ".originals")
		os.MkdirAll(backupDir, 0755)
		backupPath := filepath.Join(backupDir, base)
		copyFile(filePath, backupPath)
	}

	// Execute action
	var cmdErr error
	if folder.Command != "" {
		cmdErr = m.runCustomCommand(filePath, folder.Command)
	} else {
		cmdErr = m.runPredefinedAction(filePath, folder.Action)
	}

	if cmdErr != nil {
		log.Printf("Action failed for %s: %v", filePath, cmdErr)
		if folder.Notify {
			notify("Gato", fmt.Sprintf("Failed: %s", base))
		}
		return
	}

	log.Printf("Processed: %s", filePath)
	if folder.Notify {
		notify("Gato", fmt.Sprintf("Processed: %s", base))
	}
}

func (m *Manager) runCustomCommand(filePath, command string) error {
	// Replace {} with the file path
	cmd := strings.ReplaceAll(command, "{}", fmt.Sprintf("%q", filePath))

	// Also support {name}, {ext}, {dir}
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(base, ext)

	cmd = strings.ReplaceAll(cmd, "{name}", name)
	cmd = strings.ReplaceAll(cmd, "{ext}", ext)
	cmd = strings.ReplaceAll(cmd, "{dir}", dir)

	return exec.Command("bash", "-c", cmd).Run()
}

func (m *Manager) runPredefinedAction(filePath, action string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch action {
	case "compress":
		return m.compressFile(filePath, ext)
	case "convert-webp":
		return m.convertToWebP(filePath)
	case "convert-mp4":
		return m.convertToMP4(filePath)
	case "convert-mp3":
		return m.convertToMP3(filePath)
	case "resize-50":
		return m.resizeImage(filePath, "50%")
	case "resize-25":
		return m.resizeImage(filePath, "25%")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (m *Manager) compressFile(filePath, ext string) error {
	switch ext {
	case ".png":
		if _, err := exec.LookPath("pngquant"); err == nil {
			return exec.Command("pngquant", "--force", "--quality=65-80", "--output", filePath, filePath).Run()
		}
		return exec.Command("convert", filePath, "-strip", "-colors", "256", filePath).Run()
	case ".jpg", ".jpeg":
		return exec.Command("convert", filePath, "-strip", "-quality", "75", filePath).Run()
	case ".webp":
		return exec.Command("convert", filePath, "-strip", "-quality", "75", filePath).Run()
	default:
		return nil // Skip unsupported formats
	}
}

func (m *Manager) convertToWebP(filePath string) error {
	output := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".webp"
	err := exec.Command("convert", filePath, "-quality", "80", output).Run()
	if err == nil {
		os.Remove(filePath) // Remove original
	}
	return err
}

func (m *Manager) convertToMP4(filePath string) error {
	output := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp4"
	err := exec.Command("ffmpeg", "-i", filePath, "-c:v", "libx264", "-c:a", "aac", "-y", output).Run()
	if err == nil {
		os.Remove(filePath)
	}
	return err
}

func (m *Manager) convertToMP3(filePath string) error {
	output := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp3"
	err := exec.Command("ffmpeg", "-i", filePath, "-c:a", "libmp3lame", "-q:a", "2", "-y", output).Run()
	if err == nil {
		os.Remove(filePath)
	}
	return err
}

func (m *Manager) resizeImage(filePath, size string) error {
	return exec.Command("convert", filePath, "-resize", size, filePath).Run()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func notify(title, message string) {
	exec.Command("notify-send", title, message).Run()
}
