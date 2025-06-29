package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var (
	Cfg    *Config
	APPCtx *APPContext
)

type APPContext struct {
	// resource
	LogFile      *os.File
	Controller   *sdl.GameController
	Window       *sdl.Window
	Renderer     *sdl.Renderer
	BkTexture    *sdl.Texture
	FontFileList *ttf.Font
	FontStatus   *ttf.Font
	// state
	LastScrollTime     int64
	StatusScrollOffset int32
	CurrentDir         string
	WindowStart        int
	WindowSize         int
	Selected           int
	Btn_Select_Pressed bool
	Btn_Back_Pressed   bool
	ExecJustNow        bool
	// file entries
	FileEntries []*FileEntry
}

type FileEntry struct {
	Name  string
	IsDir bool
}

type Config struct {
	Data struct {
		Title      string `json:"title"`
		Log        string `json:"log"`
		Width      int32  `json:"width"`
		Height     int32  `json:"height"`
		LineHeight int    `json:"line_height"`
		FontSize   int    `json:"font_size"`
	} `json:"data"`
	Resources struct {
		Font         string `json:"font"`
		Bk           string `json:"bk"`
		Transparency uint8  `json:"transparency"`
	} `json:"resources"`
	Args struct {
		ShowDotFiles bool
		ConfigPath   string
		StartDir     string
		Cmds         string
		Filters      string
		FiltersMap   map[string]struct{}
	} `json:"args"`
}

func loadArgs() {
	Cfg = &Config{}
	var (
		startDirFlag, cmdsFlag, filtersFlag, cfgFlag string
		showDotFiles                                 bool
	)
	flag.StringVar(&startDirFlag, "start", "", "Start directory (default is current working directory)")
	flag.StringVar(&cmdsFlag, "cmds", "", "Command to execute with selected file (e.g. '/opt/app/mpv $__FILE__')")
	flag.StringVar(&filtersFlag, "filters", "", "Comma-separated list of allowed file extensions (e.g. 'mp4,avi')")
	flag.StringVar(&cfgFlag, "conf", "", "Config file path (default is '/etc/xxx/config.json')")
	flag.BoolVar(&showDotFiles, "show-dotfiles", false, "Show dotfiles (files beginning with '.')")
	flag.Parse()
	if strings.TrimSpace(startDirFlag) != "" {
		Cfg.Args.StartDir = startDirFlag
	} else {
		Cfg.Args.StartDir, _ = os.Getwd()
	}
	if strings.TrimSpace(cmdsFlag) != "" {
		Cfg.Args.Cmds = cmdsFlag
	}
	if strings.TrimSpace(cfgFlag) != "" {
		Cfg.Args.ConfigPath = cfgFlag
	} else {
		Cfg.Args.ConfigPath = "./config.json"
	}
	Cfg.Args.ShowDotFiles = showDotFiles
	if strings.TrimSpace(filtersFlag) != "" {
		Cfg.Args.Filters = filtersFlag
		Cfg.Args.FiltersMap = make(map[string]struct{})
		filterArr := strings.Split(filtersFlag, ",")
		for _, f := range filterArr {
			f = strings.TrimSpace(strings.ToLower(f))
			if f != "" {
				Cfg.Args.FiltersMap[f] = struct{}{}
			}
		}
	}
}
func loadCfg() {
	jsonBytes, err := os.ReadFile(Cfg.Args.ConfigPath)
	if err != nil {
		log.Println("Read config failed: ", err)
	}
	if err := json.Unmarshal(jsonBytes, &Cfg); err != nil {
		log.Println("Unmarshal json failed: ", err)
	}
	if Cfg.Data.Title == "" {
		Cfg.Data.Title = "350 File Manager"
	}
	if Cfg.Resources.Font == "" {
		Cfg.Resources.Font = "/usr/share/fonts/truetype/CharisSILB.ttf"
	}
	if Cfg.Data.Width == 0 {
		Cfg.Data.Width = 680
	}
	if Cfg.Data.Height == 0 {
		Cfg.Data.Height = 480
	}
	if Cfg.Data.LineHeight == 0 {
		Cfg.Data.LineHeight = 24
	}
	if Cfg.Data.FontSize == 0 {
		Cfg.Data.FontSize = 16
	}
}

func destroy() {
	log.Println("Destroying resources...")
	APPCtx.LogFile.Close()
	APPCtx.Window.Destroy()
	APPCtx.Renderer.Destroy()
	APPCtx.FontFileList.Close()
	APPCtx.FontStatus.Close()
	APPCtx.BkTexture.Destroy()
	APPCtx.Controller.Close()
	sdl.Quit()
	ttf.Quit()
}

func loadSysResource() error {
	var err error
	APPCtx = &APPContext{}
	if Cfg.Data.Log != "" {
		APPCtx.LogFile, err = os.OpenFile(Cfg.Data.Log, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("Open log file failed: ", err)
		}
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Set log output=", Cfg.Data.Log)
		log.SetOutput(APPCtx.LogFile)
	}
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK); err != nil {
		return err
	}
	if err = ttf.Init(); err != nil {
		return err
	}
	APPCtx.Window, err = sdl.CreateWindow(Cfg.Data.Title,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		Cfg.Data.Width, Cfg.Data.Height, sdl.WINDOWEVENT_DISPLAY_CHANGED)
	if err != nil {
		return err
	}
	APPCtx.Renderer, err = sdl.CreateRenderer(APPCtx.Window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	APPCtx.FontFileList, err = ttf.OpenFont(Cfg.Resources.Font, Cfg.Data.FontSize)
	if err != nil {
		return err
	}
	APPCtx.FontStatus, err = ttf.OpenFont(Cfg.Resources.Font, int(float64(Cfg.Data.FontSize)*0.8))
	if err != nil {
		return err
	}
	img, err := img.Load(Cfg.Resources.Bk)
	if err != nil {
		log.Printf("Failed to load background: %v", err)
		return nil
	}
	APPCtx.BkTexture, err = APPCtx.Renderer.CreateTextureFromSurface(img)
	if err != nil {
		log.Printf("Failed to create texture from background: %v", err)
		return nil
	}
	if sdl.NumJoysticks() > 0 {
		if sdl.IsGameController(0) {
			log.Println("Joystick 0 is a GameController")
			APPCtx.Controller = sdl.GameControllerOpen(0)
			if APPCtx.Controller == nil {
				return fmt.Errorf("failed to open game controller: %v", sdl.GetError())
			}
		}
	}
	sdl.GameControllerEventState(sdl.ENABLE)
	return nil
}

func load() error {
	log.Println("Load cmd args...")
	loadArgs()
	log.Println("Load config file...")
	loadCfg()
	log.Println("Load system resource...")
	return loadSysResource()
}
