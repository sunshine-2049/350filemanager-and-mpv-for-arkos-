package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

func handlePageMove(direction int) {
	if direction == -1 {
		pageSize := APPCtx.WindowSize - 1
		if APPCtx.WindowStart >= pageSize {
			APPCtx.WindowStart -= pageSize
		} else {
			APPCtx.WindowStart = 0
		}
		APPCtx.Selected = APPCtx.WindowStart
	} else if direction == 1 {
		pageSize := APPCtx.WindowSize - 1
		newStart := APPCtx.WindowStart + pageSize
		if newStart < len(APPCtx.FileEntries) {
			APPCtx.WindowStart = newStart
			APPCtx.Selected = APPCtx.WindowStart
		}
	}
	log.Println("selected=,windowstart=%d,pagesize=%d", APPCtx.Selected, APPCtx.WindowStart, APPCtx.WindowSize-1)
}

func handleSelect(delta int) {
	APPCtx.Selected += delta
	if APPCtx.Selected < 0 {
		APPCtx.Selected = len(APPCtx.FileEntries) - 1
	}
	if APPCtx.Selected >= len(APPCtx.FileEntries) {
		APPCtx.Selected = 0
	}
	/*
		file list size: window size-1
		status bar size: 1
	*/
	if APPCtx.Selected < APPCtx.WindowStart {
		APPCtx.WindowStart = APPCtx.Selected
	} else if APPCtx.Selected >= APPCtx.WindowStart+APPCtx.WindowSize-1 {
		APPCtx.WindowStart = APPCtx.Selected - (APPCtx.WindowSize - 1) + 1
		if APPCtx.WindowStart < 0 {
			APPCtx.WindowStart = 0
		}
	}
}

func listDir() error {
	files, err := os.ReadDir(APPCtx.CurrentDir)
	if err != nil {
		return err
	}
	APPCtx.FileEntries = make([]*FileEntry, 0)
	// add ../
	APPCtx.FileEntries = append(APPCtx.FileEntries, &FileEntry{Name: "..", IsDir: true})
	for _, file := range files {
		name := file.Name()
		if !Cfg.Args.ShowDotFiles && strings.HasPrefix(name, ".") {
			continue
		}
		if file.IsDir() {
			APPCtx.FileEntries = append(APPCtx.FileEntries, &FileEntry{Name: name, IsDir: true})
		} else {
			if len(Cfg.Args.FiltersMap) > 0 {
				ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
				if _, ok := Cfg.Args.FiltersMap[ext]; !ok {
					continue
				}
			}
			APPCtx.FileEntries = append(APPCtx.FileEntries, &FileEntry{Name: name, IsDir: false})
		}
	}
	return nil
}

func executeCmd(file string) error {
	if len(Cfg.Args.Cmds) <= 0 {
		return nil
	}
	cmd := exec.Command("sh", "-c", Cfg.Args.Cmds)
	cmd.Env = append(os.Environ(), fmt.Sprintf("__FILE__=%s", file))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func initManager() error {
	APPCtx.CurrentDir = Cfg.Args.StartDir
	APPCtx.WindowSize = int(Cfg.Data.Height) / Cfg.Data.LineHeight
	APPCtx.Selected = 0
	return listDir()
}

func main() {
	defer destroy()
	log.Println("Load cfg and resource...")
	err := load()
	if err != nil {
		log.Printf("Failed to preload resources: %v", err)
	}
	log.Println("Init file manager...")
	err = initManager()
	if err != nil {
		log.Printf("Failed to initialize manager: %v", err)
		return
	}
	log.Println("Start Msg loop...")
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					if e.Type == sdl.CONTROLLERBUTTONDOWN {
						switch e.Button {
						case sdl.CONTROLLER_BUTTON_B:
							handleClick()
						case sdl.CONTROLLER_BUTTON_A:
							handleReset()
						case sdl.CONTROLLER_BUTTON_DPAD_UP:
							handleSelect(-1)
						case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
							handleSelect(1)
						case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
							handlePageMove(-1)
						case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
							handlePageMove(1)
						case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
							handlePageMove(-1)
						case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
							handlePageMove(1)
						case sdl.CONTROLLER_BUTTON_BACK:
							APPCtx.Btn_Back_Pressed = true
						case sdl.CONTROLLER_BUTTON_START:
							APPCtx.Btn_Select_Pressed = true
						}
						if APPCtx.Btn_Back_Pressed && APPCtx.Btn_Select_Pressed {
							if APPCtx.ExecJustNow {
								APPCtx.ExecJustNow = false
							} else {
								running = false
							}
						}
					}
				} else if e.Type == sdl.CONTROLLERBUTTONUP {
					switch e.Button {
					case sdl.CONTROLLER_BUTTON_BACK:
						APPCtx.Btn_Back_Pressed = false
					case sdl.CONTROLLER_BUTTON_START:
						APPCtx.Btn_Select_Pressed = false
					}
				}
			}
		}
		render()
		sdl.Delay(16)
	}
}

func handleReset() {
	APPCtx.CurrentDir = Cfg.Args.StartDir
	err := listDir()
	if err != nil {
		log.Println("Read path filaed:", err)
		return
	}
	APPCtx.Selected = 0
	APPCtx.WindowStart = 0
}

func handleClick() {
	if len(APPCtx.FileEntries) == 0 {
		return
	}
	targetPath := filepath.Join(APPCtx.CurrentDir, APPCtx.FileEntries[APPCtx.Selected].Name)
	file, err := os.Stat(targetPath)
	if err != nil {
		log.Println("Get faile info failed: ", err)
		return
	}
	if file.IsDir() {
		APPCtx.CurrentDir = targetPath
		err := listDir()
		if err != nil {
			log.Println("Read path failed: ", err)
			return
		}
		APPCtx.Selected = 0
		APPCtx.WindowStart = 0
	} else {
		err := executeCmd(targetPath)
		if err != nil {
			log.Println("Execute command failed: ", err)
			showErrorDialog(err.Error())
		}
		APPCtx.ExecJustNow = true
	}
}
