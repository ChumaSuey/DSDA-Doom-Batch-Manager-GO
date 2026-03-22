package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var cfg *Config

func main() {
	cfg = LoadConfig()

	myApp := app.New()
	myWindow := myApp.NewWindow("DSDA Batch Manager")
	myWindow.Resize(fyne.NewSize(600, 620))

	tabs := container.NewAppTabs(
		container.NewTabItem("Record Demo", buildRecordTab(myWindow)),
		container.NewTabItem("Play Demo", buildPlayTab(myWindow)),
		container.NewTabItem("Settings", buildSettingsTab(myWindow)),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

// ──────────────────────────────────────────────
//  Tab 1 — Record Demo
// ──────────────────────────────────────────────

func buildRecordTab(win fyne.Window) fyne.CanvasObject {
	outputDir := cfg.DefaultOutputDir
	outputLabel := widget.NewLabel(labelOrPlaceholder(outputDir, "No folder selected"))
	outputLabel.Wrapping = fyne.TextTruncate

	// IWAD
	iwadSelect := widget.NewSelect(cfg.IWADOptions(), nil)
	iwadSelect.SetSelected(firstOrDefault(cfg.IWADOptions(), "doom2.wad"))

	// Skill
	skillSelect := widget.NewSelect([]string{
		"1 - I'm Too Young To Die",
		"2 - Hey, Not Too Rough",
		"3 - Hurt Me Plenty",
		"4 - Ultra-Violence",
		"5 - Nightmare!",
	}, nil)
	skillSelect.SetSelected("4 - Ultra-Violence")

	// Warp
	warpEntry := widget.NewEntry()
	warpEntry.SetPlaceHolder("e.g. 01 (blank = none)")

	// PWAD
	pwadEntry := widget.NewEntry()
	pwadEntry.SetPlaceHolder("e.g. sunlust.wad (blank = none)")

	// DEH
	dehEntry := widget.NewEntry()
	dehEntry.SetPlaceHolder("e.g. cyber110.deh (blank = none)")

	// Complevel
	complevelSelect := widget.NewSelect([]string{
		"2  - Doom 2 / Vanilla",
		"3  - Ultimate Doom",
		"4  - TNT / Plutonia",
		"9  - Boom",
		"11 - MBF",
		"21 - MBF21",
	}, nil)
	complevelSelect.SetSelected("2  - Doom 2 / Vanilla")

	// Demo name
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. lv01-xxx")

	// Additional params
	additionalEntry := widget.NewEntry()
	additionalEntry.SetPlaceHolder("e.g. -nomusic -nomonsters")

	// Folder picker
	folderBtn := widget.NewButton("Browse…", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			outputDir = dir.Path()
			outputLabel.SetText(outputDir)
		}, win)
	})

	// Status
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrap(fyne.TextWrapWord)

	// Shared batch generation logic
	generateBatFile := func() (string, error) {
		if nameEntry.Text == "" {
			return "", fmt.Errorf("please enter a demo name")
		}
		if outputDir == "" {
			return "", fmt.Errorf("please choose an output folder")
		}

		skillVal := string(skillSelect.Selected[0])
		complevelVal := strings.TrimSpace(strings.Split(complevelSelect.Selected, "-")[0])

		pwadVal := ""
		if pwadEntry.Text != "" {
			pwadVal = "-file " + pwadEntry.Text
		}

		bat := composeBatch(
			iwadSelect.Selected,
			skillVal,
			warpEntry.Text,
			pwadVal,
			dehEntry.Text,
			complevelVal,
			nameEntry.Text,
			additionalEntry.Text,
		)

		outPath := filepath.Join(outputDir, nameEntry.Text+".bat")
		if err := os.WriteFile(outPath, []byte(bat), 0644); err != nil {
			return "", err
		}
		return outPath, nil
	}

	// Generate button
	generateBtn := widget.NewButton("Generate .bat", func() {
		outPath, err := generateBatFile()
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("⚠ %v", err))
			return
		}
		statusLabel.SetText(fmt.Sprintf("✅ Saved to %s", outPath))
	})

	// Generate and Run button
	generateRunBtn := widget.NewButton("Generate .bat and Run", func() {
		outPath, err := generateBatFile()
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("⚠ %v", err))
			return
		}
		statusLabel.SetText(fmt.Sprintf("✅ Saved — launching %s …", filepath.Base(outPath)))

		cmd := exec.Command("cmd", "/C", outPath)
		cmd.Dir = filepath.Dir(outPath)
		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			msg := string(output)
			if msg == "" {
				msg = runErr.Error()
			}
			statusLabel.SetText(fmt.Sprintf("❌ Run error: %s", msg))
			return
		}
		statusLabel.SetText(fmt.Sprintf("✅ %s finished.", filepath.Base(outPath)))
	})

	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("IWAD"), iwadSelect,
		widget.NewLabel("Skill"), skillSelect,
		widget.NewLabel("Warp"), warpEntry,
		widget.NewLabel("PWAD(s)"), pwadEntry,
		widget.NewLabel("DEH"), dehEntry,
		widget.NewLabel("Complevel"), complevelSelect,
		widget.NewLabel("Demo Name"), nameEntry,
		widget.NewLabel("Additional"), additionalEntry,
		widget.NewLabel("Output Folder"), container.NewBorder(nil, nil, nil, folderBtn, outputLabel),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Record Demo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		generateBtn,
		widget.NewSeparator(),
		generateRunBtn,
		statusLabel,
	)
}

// composeBatch builds the .bat file content.
func composeBatch(iwad, skill, warp, pwad, deh, complevel, name, additional string) string {
	var b strings.Builder
	b.WriteString("@echo off\r\n")
	b.WriteString(fmt.Sprintf("set iwad=%s\r\n", iwad))
	b.WriteString(fmt.Sprintf("set skill=%s\r\n", skill))
	b.WriteString(fmt.Sprintf("set warp=%s\r\n", warp))
	b.WriteString(fmt.Sprintf("set pwad=%s\r\n", pwad))
	b.WriteString(fmt.Sprintf("set deh=%s\r\n", deh))
	b.WriteString(fmt.Sprintf("set complevel=%s\r\n", complevel))
	b.WriteString(fmt.Sprintf("set name=%s\r\n", name))
	b.WriteString(fmt.Sprintf("set additional=%s\r\n", additional))
	b.WriteString("\r\n")
	b.WriteString("set path=.\\%%name%%\\\r\n")
	b.WriteString("if not exist \"%%path%%\" mkdir %%path%%\r\n")
	b.WriteString("set record=%%path%%%%name%%.lmp\r\n")
	b.WriteString("\r\n")
	b.WriteString("dsda-doom.exe -iwad %%iwad%% -skill %%skill%% -deh %%deh%% -warp %%warp%% -file %%pwad%% -complevel %%complevel%% -record %%record%% %%additional%%\r\n")
	return b.String()
}

// ──────────────────────────────────────────────
//  Tab 2 — Play Demo
// ──────────────────────────────────────────────

type lmpFile struct {
	Name    string
	Size    int64
	ModTime time.Time
	Path    string
}

func buildPlayTab(win fyne.Window) fyne.CanvasObject {
	demosDir := cfg.DefaultDemosDir
	var lmpFiles []lmpFile
	var selectedLmp *lmpFile

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrap(fyne.TextWrapWord)

	// Demo list
	demoList := widget.NewList(
		func() int { return len(lmpFiles) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("filename.lmp"),
				layout.NewSpacer(),
				widget.NewLabel("0 KB"),
				widget.NewLabel("2025-01-01"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			f := lmpFiles[id]
			box.Objects[0].(*widget.Label).SetText(f.Name)
			box.Objects[2].(*widget.Label).SetText(formatSize(f.Size))
			box.Objects[3].(*widget.Label).SetText(f.ModTime.Format("2006-01-02 15:04"))
		},
	)

	demoList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(lmpFiles) {
			selectedLmp = &lmpFiles[id]
			statusLabel.SetText(fmt.Sprintf("Selected: %s", selectedLmp.Name))
		}
	}

	// Scan function
	scanDir := func(dir string) {
		lmpFiles = nil
		selectedLmp = nil
		entries, err := os.ReadDir(dir)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("❌ Could not read folder: %v", err))
			demoList.Refresh()
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.EqualFold(filepath.Ext(e.Name()), ".lmp") {
				info, _ := e.Info()
				lmpFiles = append(lmpFiles, lmpFile{
					Name:    e.Name(),
					Size:    info.Size(),
					ModTime: info.ModTime(),
					Path:    filepath.Join(dir, e.Name()),
				})
			}
		}
		demoList.Refresh()
		if len(lmpFiles) == 0 {
			statusLabel.SetText("No .lmp files found in this folder.")
		} else {
			statusLabel.SetText(fmt.Sprintf("Found %d demo(s).", len(lmpFiles)))
		}
	}

	// If a default demos dir is set, scan it on load
	if demosDir != "" {
		scanDir(demosDir)
	}

	// Folder label
	folderLabel := widget.NewLabel(labelOrPlaceholder(demosDir, "No folder selected"))
	folderLabel.Wrapping = fyne.TextTruncate

	browseBtn := widget.NewButton("Browse…", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			demosDir = dir.Path()
			folderLabel.SetText(demosDir)
			scanDir(demosDir)
		}, win)
	})

	// IWAD for playback
	iwadSelect := widget.NewSelect(cfg.IWADOptions(), nil)
	iwadSelect.SetSelected(firstOrDefault(cfg.IWADOptions(), "doom2.wad"))

	// Optional PWAD for playback
	pwadEntry := widget.NewEntry()
	pwadEntry.SetPlaceHolder("Optional PWAD (blank = none)")

	// Export text checkbox
	exportText := widget.NewCheck("Export text file (-export_text_file)", nil)
	exportTextDesc := widget.NewLabel("Auto-generates a .txt with run time, category, etc.")
	exportTextDesc.Wrapping = fyne.TextWrap(fyne.TextWrapWord)

	// Play button
	playBtn := widget.NewButton("▶ Play Selected Demo", func() {
		if selectedLmp == nil {
			statusLabel.SetText("⚠ Select a demo first.")
			return
		}
		if cfg.DSDADoomPath == "" {
			statusLabel.SetText("⚠ Set DSDA-Doom path in Settings first.")
			return
		}

		args := []string{
			"-iwad", iwadSelect.Selected,
			"-playdemo", selectedLmp.Path,
		}
		if pwadEntry.Text != "" {
			args = append(args, "-file", pwadEntry.Text)
		}
		if exportText.Checked {
			args = append(args, "-export_text_file")
		}

		cmd := exec.Command(cfg.DSDADoomPath, args...)
		cmd.Dir = filepath.Dir(cfg.DSDADoomPath)
		if err := cmd.Start(); err != nil {
			statusLabel.SetText(fmt.Sprintf("❌ Launch error: %v", err))
			return
		}
		statusLabel.SetText(fmt.Sprintf("🎮 Launched: %s", selectedLmp.Name))
	})

	// Layout
	topRow := container.NewBorder(nil, nil, widget.NewLabel("Demos Folder"), browseBtn, folderLabel)
	options := container.New(layout.NewFormLayout(),
		widget.NewLabel("IWAD"), iwadSelect,
		widget.NewLabel("PWAD"), pwadEntry,
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Play Demo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		topRow,
		widget.NewSeparator(),
		container.New(layout.NewMaxLayout(), container.NewVScroll(demoList)),
		widget.NewSeparator(),
		options,
		exportText,
		exportTextDesc,
		playBtn,
		statusLabel,
	)
}

// ──────────────────────────────────────────────
//  Tab 3 — Settings
// ──────────────────────────────────────────────

func buildSettingsTab(win fyne.Window) fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrap(fyne.TextWrapWord)

	// ── DSDA-Doom path ──
	dsdaLabel := widget.NewLabel(labelOrPlaceholder(cfg.DSDADoomPath, "Not set"))
	dsdaLabel.Wrapping = fyne.TextTruncate

	dsdaBrowse := widget.NewButton("Browse…", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			cfg.DSDADoomPath = f.URI().Path()
			dsdaLabel.SetText(cfg.DSDADoomPath)
			f.Close()
		}, win)
	})

	// ── IWAD Folder (auto-scan) ──
	iwadFolderLabel := widget.NewLabel(labelOrPlaceholder(cfg.IWADFolder, "Not set"))
	iwadFolderLabel.Wrapping = fyne.TextTruncate

	// IWAD list (shared between folder scan and manual add)
	iwadListWidget := widget.NewList(
		func() int { return len(cfg.IWADPaths) },
		func() fyne.CanvasObject {
			return widget.NewLabel("path/to/iwad.wad")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(cfg.IWADPaths[id])
		},
	)

	scanIwadBtn := widget.NewButton("Scan Folder", func() {
		if cfg.IWADFolder == "" {
			statusLabel.SetText("⚠ Set an IWAD folder first.")
			return
		}
		count, err := cfg.ScanIWADFolder()
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("❌ Scan error: %v", err))
			return
		}
		iwadListWidget.Refresh()
		statusLabel.SetText(fmt.Sprintf("🔍 Found %d IWAD(s) in folder.", count))
	})

	iwadFolderBrowse := widget.NewButton("Browse…", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			cfg.IWADFolder = dir.Path()
			iwadFolderLabel.SetText(cfg.IWADFolder)
			// Auto-scan immediately after selecting folder
			count, scanErr := cfg.ScanIWADFolder()
			if scanErr != nil {
				statusLabel.SetText(fmt.Sprintf("❌ Scan error: %v", scanErr))
				return
			}
			iwadListWidget.Refresh()
			statusLabel.SetText(fmt.Sprintf("🔍 Found %d IWAD(s) in folder.", count))
		}, win)
	})

	// Manual add/remove (fallback)
	addIwadBtn := widget.NewButton("Add IWAD…", func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, err error) {
			if err != nil || f == nil {
				return
			}
			cfg.IWADPaths = append(cfg.IWADPaths, f.URI().Path())
			iwadListWidget.Refresh()
			f.Close()
		}, win)
	})

	removeIwadBtn := widget.NewButton("Remove Selected", func() {
		if len(cfg.IWADPaths) > 0 {
			cfg.IWADPaths = cfg.IWADPaths[:len(cfg.IWADPaths)-1]
			iwadListWidget.Refresh()
		}
	})

	// ── Default Demos Folder ──
	demosLabel := widget.NewLabel(labelOrPlaceholder(cfg.DefaultDemosDir, "Not set"))
	demosLabel.Wrapping = fyne.TextTruncate

	demosBrowse := widget.NewButton("Browse…", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			cfg.DefaultDemosDir = dir.Path()
			demosLabel.SetText(cfg.DefaultDemosDir)
		}, win)
	})

	// ── Default Output Folder ──
	outputLabel := widget.NewLabel(labelOrPlaceholder(cfg.DefaultOutputDir, "Not set"))
	outputLabel.Wrapping = fyne.TextTruncate

	outputBrowse := widget.NewButton("Browse…", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			cfg.DefaultOutputDir = dir.Path()
			outputLabel.SetText(cfg.DefaultOutputDir)
		}, win)
	})

	// ── Save button ──
	saveBtn := widget.NewButton("Save Settings", func() {
		savedPath, err := SaveConfig(cfg)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("❌ Error saving: %v", err))
			return
		}
		statusLabel.SetText(fmt.Sprintf("✅ Settings saved to %s", savedPath))
	})

	// Layout
	form := container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("DSDA-Doom Executable", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, dsdaBrowse, dsdaLabel),

		widget.NewSeparator(),
		widget.NewLabelWithStyle("IWAD Folder (Auto-Scan)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Point to your IWADs folder — known IWADs are detected automatically."),
		container.NewBorder(nil, nil, nil, iwadFolderBrowse, iwadFolderLabel),
		scanIwadBtn,

		widget.NewSeparator(),
		widget.NewLabelWithStyle("Detected IWADs", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.New(layout.NewMaxLayout(), container.NewVScroll(iwadListWidget)),
		container.NewHBox(addIwadBtn, removeIwadBtn),

		widget.NewSeparator(),
		widget.NewLabelWithStyle("Default Folders", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Demos Folder"), container.NewBorder(nil, nil, nil, demosBrowse, demosLabel),
			widget.NewLabel("Output Folder"), container.NewBorder(nil, nil, nil, outputBrowse, outputLabel),
		),

		widget.NewSeparator(),
		saveBtn,
		widget.NewLabelWithStyle("Note: When changing default folders, please reboot the app for them to take effect across all tabs.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		statusLabel,
	)

	return form
}

// ──────────────────────────────────────────────
//  Helpers
// ──────────────────────────────────────────────

func labelOrPlaceholder(val, placeholder string) string {
	if val == "" {
		return placeholder
	}
	return val
}

func firstOrDefault(opts []string, fallback string) string {
	for _, o := range opts {
		if strings.Contains(strings.ToLower(o), strings.TrimSuffix(fallback, ".wad")) {
			return o
		}
	}
	if len(opts) > 0 {
		return opts[0]
	}
	return fallback
}

func formatSize(bytes int64) string {
	const kb = 1024
	const mb = kb * 1024
	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
