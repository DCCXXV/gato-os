extends Control

var folders: Array = []
var current_folder: String = ""
var user_presets: Dictionary = {}

# Colors
const MAIN_BG = Color("051315")
const SIDEBAR_BG = Color("0a1d20f5")
const ACCENT = Color("FF7A7A")
const ACCENT_DIM = Color("FF7A7A3D")
const ACCENT_HOVER = Color("FF7A7A1E")
const TEXT = Color("c8d2d2")
const DIM = Color("647376")
const BORDER = Color("1a3a3e")
const SURFACE = Color("0c1e22")
const HOVER = Color("122a2f")

const PRESET_ORDER = ["Compress", "WebP", "MP4", "MP3", "Resize 50%", "Resize 25%", "PNG", "JPG", "Optimize", "GIF"]
const DEFAULT_PRESETS = {
	"Compress": "convert {} -strip -quality 75 {}",
	"WebP": "convert {} -quality 90 {dir}/{name}.webp && rm {}",
	"PNG": "convert {} {dir}/{name}.png && rm {}",
	"JPG": "convert {} -quality 85 {dir}/{name}.jpg && rm {}",
	"Optimize": "pngquant --force --quality=65-80 --output {} {}",
	"Resize 50%": "convert {} -resize 50% {}",
	"Resize 25%": "convert {} -resize 25% {}",
	"MP4": "ffmpeg -i {} -c:v libx264 -c:a aac -y {dir}/{name}.mp4 && rm {}",
	"MP3": "ffmpeg -i {} -c:a libmp3lame -q:a 2 -y {dir}/{name}.mp3 && rm {}",
	"GIF": "ffmpeg -i {} -vf 'fps=10,scale=480:-1' -y {dir}/{name}.gif && rm {}",
}

# Nodes
@onready var hsplit = $HSplit
@onready var folder_list = $HSplit/SidebarPanel/SidebarMargin/SidebarVBox/FolderScroll/FolderList
@onready var folders_label = $HSplit/SidebarPanel/SidebarMargin/SidebarVBox/SidebarHeader/FoldersLabel
@onready var add_folder_btn = $HSplit/SidebarPanel/SidebarMargin/SidebarVBox/SidebarBottom/AddFolderBtn

@onready var welcome_panel = $HSplit/ContentPanel/ContentMargin/ContentVBox/WelcomePanel
@onready var welcome_title = $HSplit/ContentPanel/ContentMargin/ContentVBox/WelcomePanel/WelcomeTitle
@onready var welcome_subtitle = $HSplit/ContentPanel/ContentMargin/ContentVBox/WelcomePanel/WelcomeSubtitle
@onready var welcome_hint = $HSplit/ContentPanel/ContentMargin/ContentVBox/WelcomePanel/WelcomeHint

@onready var config_panel = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel
@onready var folder_name_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/Header/HeaderRow/FolderName
@onready var folder_path_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/Header/FolderPath
@onready var open_folder_btn = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/Header/HeaderRow/OpenFolderBtn
@onready var commands_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/CommandsSection/CommandsLabel
@onready var commands_list = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/CommandsSection/CommandsScroll/CommandsList
@onready var add_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/AddLabel
@onready var hint_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/Hint
@onready var command_input = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/CommandInput
@onready var presets_label = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/PresetsLabel
@onready var presets_flow = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/PresetsFlow
@onready var save_preset_btn = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/ButtonRow/SavePresetBtn
@onready var manage_presets_btn = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/ButtonRow/ManagePresetsBtn
@onready var add_btn = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/AddSection/ButtonRow/AddBtn
@onready var remove_folder_btn = $HSplit/ContentPanel/ContentMargin/ContentVBox/ConfigPanel/Footer/RemoveFolderBtn

@onready var file_dialog = $FileDialog
@onready var save_preset_dialog = $SavePresetDialog
@onready var preset_name_input = $SavePresetDialog/VBox/PresetNameInput
@onready var presets_dialog = $PresetsDialog
@onready var presets_list = $PresetsDialog/Margin/Scroll/PresetsList
@onready var confirm_dialog = $ConfirmDialog

var pending_confirm_action: Callable

func _ready():
	_apply_theme()
	_connect_signals()
	_load_user_presets()
	_setup_presets()

	var args = OS.get_cmdline_args()
	for arg in args:
		if not arg.begins_with("-") and not arg.ends_with(".tscn") and not arg.ends_with(".pck"):
			current_folder = arg
			break

	load_folders()

	if current_folder != "":
		show_folder_config(current_folder)
	else:
		show_welcome()

func _apply_theme():
	# HSplit grabber styling
	var grabber_style = StyleBoxFlat.new()
	grabber_style.bg_color = BORDER
	hsplit.add_theme_stylebox_override("grabber", grabber_style)
	hsplit.add_theme_icon_override("grabber", ImageTexture.new())

	# Labels
	folders_label.add_theme_color_override("font_color", DIM)
	welcome_title.add_theme_color_override("font_color", TEXT)
	welcome_subtitle.add_theme_color_override("font_color", DIM)
	welcome_hint.add_theme_color_override("font_color", DIM)
	folder_name_label.add_theme_color_override("font_color", TEXT)
	folder_path_label.add_theme_color_override("font_color", DIM)
	commands_label.add_theme_color_override("font_color", DIM)
	add_label.add_theme_color_override("font_color", DIM)
	hint_label.add_theme_color_override("font_color", DIM)
	presets_label.add_theme_color_override("font_color", DIM)

	# Buttons
	_style_button(add_folder_btn, "secondary")
	_style_button(open_folder_btn, "ghost")
	_style_button(save_preset_btn, "ghost")
	_style_button(manage_presets_btn, "ghost")
	_style_button(add_btn, "primary")
	_style_button(remove_folder_btn, "danger")

	# Input
	_style_input(command_input)
	_style_input(preset_name_input)

func _style_button(btn: Button, style_type: String):
	var normal = StyleBoxFlat.new()
	var hover = StyleBoxFlat.new()
	var pressed = StyleBoxFlat.new()

	for s in [normal, hover, pressed]:
		s.content_margin_left = 12
		s.content_margin_right = 12
		s.content_margin_top = 6
		s.content_margin_bottom = 6

	match style_type:
		"primary":
			normal.bg_color = ACCENT
			hover.bg_color = ACCENT.lightened(0.1)
			pressed.bg_color = ACCENT.darkened(0.1)
			btn.add_theme_color_override("font_color", MAIN_BG)
			btn.add_theme_color_override("font_hover_color", MAIN_BG)
			btn.add_theme_color_override("font_pressed_color", MAIN_BG)
		"secondary":
			normal.bg_color = SURFACE
			normal.border_width_left = 1
			normal.border_width_right = 1
			normal.border_width_top = 1
			normal.border_width_bottom = 1
			normal.border_color = BORDER
			hover.bg_color = HOVER
			hover.border_width_left = 1
			hover.border_width_right = 1
			hover.border_width_top = 1
			hover.border_width_bottom = 1
			hover.border_color = BORDER
			pressed.bg_color = SURFACE
			pressed.border_width_left = 1
			pressed.border_width_right = 1
			pressed.border_width_top = 1
			pressed.border_width_bottom = 1
			pressed.border_color = ACCENT
			btn.add_theme_color_override("font_color", TEXT)
			btn.add_theme_color_override("font_hover_color", TEXT)
			btn.add_theme_color_override("font_pressed_color", ACCENT)
		"ghost":
			normal.bg_color = Color.TRANSPARENT
			hover.bg_color = HOVER
			pressed.bg_color = SURFACE
			btn.add_theme_color_override("font_color", DIM)
			btn.add_theme_color_override("font_hover_color", TEXT)
			btn.add_theme_color_override("font_pressed_color", ACCENT)
		"danger":
			normal.bg_color = Color.TRANSPARENT
			normal.border_width_left = 1
			normal.border_width_right = 1
			normal.border_width_top = 1
			normal.border_width_bottom = 1
			normal.border_color = BORDER
			hover.bg_color = Color("3a1a1e")
			hover.border_width_left = 1
			hover.border_width_right = 1
			hover.border_width_top = 1
			hover.border_width_bottom = 1
			hover.border_color = ACCENT
			pressed.bg_color = ACCENT.darkened(0.5)
			btn.add_theme_color_override("font_color", DIM)
			btn.add_theme_color_override("font_hover_color", ACCENT)
			btn.add_theme_color_override("font_pressed_color", TEXT)
		"preset":
			normal.bg_color = SURFACE
			normal.border_width_left = 1
			normal.border_width_right = 1
			normal.border_width_top = 1
			normal.border_width_bottom = 1
			normal.border_color = BORDER
			hover.bg_color = HOVER
			hover.border_width_left = 1
			hover.border_width_right = 1
			hover.border_width_top = 1
			hover.border_width_bottom = 1
			hover.border_color = DIM
			pressed.bg_color = ACCENT_DIM
			pressed.border_width_left = 1
			pressed.border_width_right = 1
			pressed.border_width_top = 1
			pressed.border_width_bottom = 1
			pressed.border_color = ACCENT
			btn.add_theme_color_override("font_color", DIM)
			btn.add_theme_color_override("font_hover_color", TEXT)
			btn.add_theme_color_override("font_pressed_color", ACCENT)
		"user_preset":
			normal.bg_color = ACCENT_DIM
			normal.border_width_left = 1
			normal.border_width_right = 1
			normal.border_width_top = 1
			normal.border_width_bottom = 1
			normal.border_color = ACCENT.darkened(0.3)
			hover.bg_color = ACCENT_DIM.lightened(0.1)
			hover.border_width_left = 1
			hover.border_width_right = 1
			hover.border_width_top = 1
			hover.border_width_bottom = 1
			hover.border_color = ACCENT
			pressed.bg_color = ACCENT.darkened(0.3)
			btn.add_theme_color_override("font_color", TEXT)
			btn.add_theme_color_override("font_hover_color", TEXT)
			btn.add_theme_color_override("font_pressed_color", TEXT)

	btn.add_theme_stylebox_override("normal", normal)
	btn.add_theme_stylebox_override("hover", hover)
	btn.add_theme_stylebox_override("pressed", pressed)
	btn.add_theme_stylebox_override("focus", StyleBoxEmpty.new())

func _style_input(input: LineEdit):
	var normal = StyleBoxFlat.new()
	normal.bg_color = SURFACE
	normal.border_width_left = 1
	normal.border_width_right = 1
	normal.border_width_top = 1
	normal.border_width_bottom = 1
	normal.border_color = BORDER
	normal.content_margin_left = 10
	normal.content_margin_right = 10
	normal.content_margin_top = 8
	normal.content_margin_bottom = 8

	var focus = normal.duplicate()
	focus.border_color = ACCENT

	input.add_theme_stylebox_override("normal", normal)
	input.add_theme_stylebox_override("focus", focus)
	input.add_theme_color_override("font_color", TEXT)
	input.add_theme_color_override("font_placeholder_color", DIM)
	input.add_theme_color_override("caret_color", ACCENT)
	input.add_theme_color_override("selection_color", ACCENT_DIM)

func _connect_signals():
	add_folder_btn.pressed.connect(_on_add_folder_pressed)
	open_folder_btn.pressed.connect(_on_open_folder_pressed)
	add_btn.pressed.connect(_on_add_command)
	remove_folder_btn.pressed.connect(_on_remove_folder_pressed)
	save_preset_btn.pressed.connect(_on_save_preset_pressed)
	manage_presets_btn.pressed.connect(_on_manage_presets_pressed)
	save_preset_dialog.confirmed.connect(_on_save_preset_confirmed)
	confirm_dialog.confirmed.connect(_on_confirm_action)
	command_input.text_submitted.connect(func(_t): _on_add_command())

func _get_presets_path() -> String:
	return OS.get_environment("HOME") + "/.config/gato/presets.json"

func _load_user_presets():
	user_presets.clear()
	var file = FileAccess.open(_get_presets_path(), FileAccess.READ)
	if file:
		var json = JSON.new()
		if json.parse(file.get_as_text()) == OK:
			var data = json.get_data()
			if data is Dictionary and data.has("presets"):
				user_presets = data["presets"]
		file.close()

func _save_user_presets():
	var dir = _get_presets_path().get_base_dir()
	DirAccess.make_dir_recursive_absolute(dir)
	var file = FileAccess.open(_get_presets_path(), FileAccess.WRITE)
	if file:
		file.store_string(JSON.stringify({"presets": user_presets}, "  "))
		file.close()

func _setup_presets():
	for child in presets_flow.get_children():
		child.queue_free()

	for preset_name in PRESET_ORDER:
		var btn = Button.new()
		btn.text = preset_name
		btn.add_theme_font_size_override("font_size", 10)
		_style_button(btn, "preset")
		btn.pressed.connect(_on_preset_pressed.bind(preset_name, false))
		presets_flow.add_child(btn)

	for preset_name in user_presets.keys():
		var btn = Button.new()
		btn.text = preset_name
		btn.add_theme_font_size_override("font_size", 10)
		_style_button(btn, "user_preset")
		btn.pressed.connect(_on_preset_pressed.bind(preset_name, true))
		presets_flow.add_child(btn)

func load_folders():
	folders.clear()
	var output = []
	var exit_code = OS.execute("gato", ["f", "ls", "--raw"], output, true)

	if exit_code != 0 or output.size() == 0:
		return

	var folder_map = {}
	for line in output[0].split("\n"):
		if line.strip_edges() == "":
			continue
		var parts = line.split("\t")
		var path = parts[0]
		var cmd = parts[1] if parts.size() > 1 else ""

		if not folder_map.has(path):
			folder_map[path] = []
		if cmd != "":
			folder_map[path].append({"command": cmd})

	for path in folder_map:
		folders.append({"path": path, "actions": folder_map[path]})

func show_welcome():
	current_folder = ""
	welcome_panel.visible = true
	config_panel.visible = false
	_refresh_folder_list()

func show_folder_config(path: String):
	current_folder = path
	welcome_panel.visible = false
	config_panel.visible = true

	folder_name_label.text = path.get_file()
	folder_path_label.text = path

	_refresh_folder_list()
	_refresh_commands_list()

func _refresh_folder_list():
	for child in folder_list.get_children():
		child.queue_free()

	for folder in folders:
		var path: String = folder.get("path", "")
		var actions: Array = folder.get("actions", [])
		var is_selected = path == current_folder
		folder_list.add_child(_create_folder_item(path, actions.size(), is_selected))

func _create_folder_item(path: String, cmd_count: int, is_selected: bool) -> Control:
	var panel = PanelContainer.new()
	panel.size_flags_horizontal = Control.SIZE_EXPAND_FILL

	# Style the panel
	var style = StyleBoxFlat.new()
	style.bg_color = ACCENT_HOVER if is_selected else Color.TRANSPARENT
	style.border_width_left = 3
	style.border_color = ACCENT if is_selected else Color.TRANSPARENT
	style.content_margin_left = 12
	style.content_margin_right = 12
	style.content_margin_top = 10
	style.content_margin_bottom = 10
	panel.add_theme_stylebox_override("panel", style)

	# Content
	var vbox = VBoxContainer.new()
	vbox.mouse_filter = Control.MOUSE_FILTER_IGNORE

	var name_label = Label.new()
	name_label.text = path.get_file()
	name_label.add_theme_color_override("font_color", ACCENT if is_selected else TEXT)
	name_label.add_theme_font_size_override("font_size", 13)
	name_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	vbox.add_child(name_label)

	var count_label = Label.new()
	count_label.text = "%d command%s" % [cmd_count, "s" if cmd_count != 1 else ""]
	count_label.add_theme_color_override("font_color", DIM)
	count_label.add_theme_font_size_override("font_size", 10)
	count_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	vbox.add_child(count_label)

	panel.add_child(vbox)

	# Make clickable
	panel.gui_input.connect(func(event):
		if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
			show_folder_config(path)
	)
	panel.mouse_default_cursor_shape = Control.CURSOR_POINTING_HAND

	return panel

func _refresh_commands_list():
	for child in commands_list.get_children():
		child.queue_free()

	var folder_data = null
	for f in folders:
		if f.get("path", "") == current_folder:
			folder_data = f
			break

	if folder_data == null:
		return

	var actions: Array = folder_data.get("actions", [])

	if actions.size() == 0:
		var empty = Label.new()
		empty.text = "No commands configured"
		empty.add_theme_color_override("font_color", DIM)
		empty.add_theme_font_size_override("font_size", 12)
		commands_list.add_child(empty)
		return

	for action in actions:
		var cmd: String = action.get("command", action.get("action", ""))
		commands_list.add_child(_create_command_item(cmd))

func _create_command_item(cmd: String) -> Control:
	var panel = PanelContainer.new()
	var style = StyleBoxFlat.new()
	style.bg_color = SURFACE
	style.border_width_left = 1
	style.border_width_right = 1
	style.border_width_top = 1
	style.border_width_bottom = 1
	style.border_color = BORDER
	style.content_margin_left = 12
	style.content_margin_right = 8
	style.content_margin_top = 8
	style.content_margin_bottom = 8
	panel.add_theme_stylebox_override("panel", style)

	var hbox = HBoxContainer.new()
	panel.add_child(hbox)

	var label = Label.new()
	label.text = cmd if cmd.length() <= 50 else cmd.substr(0, 47) + "..."
	label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	label.add_theme_color_override("font_color", TEXT)
	label.add_theme_font_size_override("font_size", 11)
	hbox.add_child(label)

	var remove_btn = Button.new()
	remove_btn.text = "x"
	remove_btn.add_theme_font_size_override("font_size", 12)
	_style_button(remove_btn, "ghost")
	remove_btn.pressed.connect(_on_remove_command.bind(cmd))
	hbox.add_child(remove_btn)

	return panel

func _on_add_folder_pressed():
	# Use native file picker via XDG Desktop Portal
	DisplayServer.file_dialog_show(
		"Select Folder",
		OS.get_environment("HOME"),
		"",
		false,  # show_hidden
		DisplayServer.FILE_DIALOG_MODE_OPEN_DIR,
		PackedStringArray(),
		_on_native_folder_selected
	)

func _on_native_folder_selected(status: bool, selected: PackedStringArray, _idx: int):
	if status and selected.size() > 0:
		_on_folder_selected(selected[0])

func _on_folder_selected(dir: String):
	OS.execute("gato", ["f", "add", dir])
	load_folders()
	show_folder_config(dir)

func _on_open_folder_pressed():
	if current_folder != "":
		OS.shell_open(current_folder)

func _on_preset_pressed(preset_name: String, is_user_preset: bool):
	if is_user_preset:
		command_input.text = user_presets.get(preset_name, "")
	else:
		command_input.text = DEFAULT_PRESETS.get(preset_name, "")

func _on_add_command():
	var cmd = command_input.text.strip_edges()
	if cmd == "":
		return
	OS.execute("gato", ["f", "add", current_folder, cmd])
	command_input.text = ""
	load_folders()
	_refresh_commands_list()
	_refresh_folder_list()

func _on_remove_command(cmd: String):
	OS.execute("gato", ["f", "rm", current_folder, cmd])
	load_folders()
	_refresh_commands_list()
	_refresh_folder_list()

func _on_remove_folder_pressed():
	confirm_dialog.dialog_text = "Remove '%s' from Gato Carpetas?\nThe folder itself will not be deleted." % current_folder.get_file()
	pending_confirm_action = func():
		OS.execute("gato", ["f", "rm", current_folder])  # No second arg = remove entire folder
		load_folders()
		show_welcome()
	confirm_dialog.popup_centered()

func _on_confirm_action():
	if pending_confirm_action:
		pending_confirm_action.call()
		pending_confirm_action = Callable()

func _on_save_preset_pressed():
	if command_input.text.strip_edges() == "":
		return
	preset_name_input.text = ""
	save_preset_dialog.popup_centered()

func _on_save_preset_confirmed():
	var pname = preset_name_input.text.strip_edges()
	var cmd = command_input.text.strip_edges()
	if pname == "" or cmd == "":
		return
	user_presets[pname] = cmd
	_save_user_presets()
	_setup_presets()

func _on_manage_presets_pressed():
	_refresh_presets_dialog()
	presets_dialog.popup_centered()

func _refresh_presets_dialog():
	for child in presets_list.get_children():
		child.queue_free()

	var user_label = Label.new()
	user_label.text = "CUSTOM PRESETS"
	user_label.add_theme_color_override("font_color", DIM)
	user_label.add_theme_font_size_override("font_size", 10)
	presets_list.add_child(user_label)

	if user_presets.size() == 0:
		var empty = Label.new()
		empty.text = "No custom presets saved"
		empty.add_theme_color_override("font_color", DIM)
		empty.add_theme_font_size_override("font_size", 12)
		presets_list.add_child(empty)
	else:
		for preset_name in user_presets.keys():
			presets_list.add_child(_create_preset_dialog_item(preset_name, user_presets[preset_name], true))

	var spacer = Control.new()
	spacer.custom_minimum_size = Vector2(0, 16)
	presets_list.add_child(spacer)

	var default_label = Label.new()
	default_label.text = "DEFAULT PRESETS"
	default_label.add_theme_color_override("font_color", DIM)
	default_label.add_theme_font_size_override("font_size", 10)
	presets_list.add_child(default_label)

	for preset_name in PRESET_ORDER:
		presets_list.add_child(_create_preset_dialog_item(preset_name, DEFAULT_PRESETS[preset_name], false))

func _create_preset_dialog_item(preset_name: String, cmd: String, deletable: bool) -> Control:
	var panel = PanelContainer.new()
	var style = StyleBoxFlat.new()
	style.bg_color = SURFACE
	style.border_width_left = 1
	style.border_width_right = 1
	style.border_width_top = 1
	style.border_width_bottom = 1
	style.border_color = BORDER
	style.content_margin_left = 12
	style.content_margin_right = 12
	style.content_margin_top = 10
	style.content_margin_bottom = 10
	panel.add_theme_stylebox_override("panel", style)

	var hbox = HBoxContainer.new()
	panel.add_child(hbox)

	var vbox = VBoxContainer.new()
	vbox.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	hbox.add_child(vbox)

	var name_label = Label.new()
	name_label.text = preset_name
	name_label.add_theme_color_override("font_color", TEXT if deletable else DIM)
	name_label.add_theme_font_size_override("font_size", 13)
	vbox.add_child(name_label)

	var cmd_label = Label.new()
	cmd_label.text = cmd if cmd.length() <= 40 else cmd.substr(0, 37) + "..."
	cmd_label.add_theme_color_override("font_color", DIM)
	cmd_label.add_theme_font_size_override("font_size", 10)
	vbox.add_child(cmd_label)

	if deletable:
		var delete_btn = Button.new()
		delete_btn.text = "x"
		delete_btn.add_theme_font_size_override("font_size", 12)
		_style_button(delete_btn, "ghost")
		delete_btn.pressed.connect(_on_delete_preset.bind(preset_name))
		hbox.add_child(delete_btn)

	return panel

func _on_delete_preset(preset_name: String):
	user_presets.erase(preset_name)
	_save_user_presets()
	_refresh_presets_dialog()
	_setup_presets()
