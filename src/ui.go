package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func renderText(fontSize *ttf.Font, x, y int32, text string, color sdl.Color) error {
	surf, err := fontSize.RenderUTF8Solid(text, color)
	if err != nil {
		return err
	}
	defer surf.Free()
	tex, err := APPCtx.Renderer.CreateTextureFromSurface(surf)
	if err != nil {
		return err
	}
	defer tex.Destroy()

	_, _, w, h, _ := tex.Query()
	return APPCtx.Renderer.Copy(tex, nil, &sdl.Rect{X: x, Y: y, W: w, H: h})
}

func showErrorDialog(message string) {
	bg, err := APPCtx.Renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET,
		Cfg.Data.Width, Cfg.Data.Height)
	if err != nil {
		log.Printf("CreateTexture failed: %v", err)
		return
	}
	APPCtx.Renderer.SetRenderTarget(bg)
	render()
	APPCtx.Renderer.SetRenderTarget(nil)
	showing := true
	for showing {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					showing = false
				}
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					showing = false
				}
			}
		}
		APPCtx.Renderer.SetDrawColor(0, 0, 0, 160)
		APPCtx.Renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: Cfg.Data.Width, H: Cfg.Data.Height})
		if bg != nil {
			APPCtx.Renderer.Copy(bg, nil, nil)
		}
		w := int32(400)
		h := int32(100)
		x := (Cfg.Data.Width - w) / 2
		y := (Cfg.Data.Height - h) / 2
		APPCtx.Renderer.SetDrawColor(50, 50, 50, 240)
		APPCtx.Renderer.FillRect(&sdl.Rect{X: x, Y: y, W: w, H: h})
		renderText(APPCtx.FontStatus, x+20, y+20, message, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		renderText(APPCtx.FontStatus, x+20, y+60, "Any key to close", sdl.Color{R: 180, G: 180, B: 180, A: 255})
		APPCtx.Renderer.Present()
		sdl.Delay(16)
	}
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d Bytes", size)
	}
	kb := float64(size) / 1024.0
	mb := kb / 1024.0
	gb := mb / 1024.0
	switch {
	case gb >= 1:
		return fmt.Sprintf("%.2f GB", gb)
	case mb >= 1:
		return fmt.Sprintf("%.2f MB", mb)
	default:
		return fmt.Sprintf("%.2f KB", kb)
	}
}

func renderTextScrolled(font *ttf.Font, x, y int32, text string, scrollOffset int32, color sdl.Color) {
	surface, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		log.Printf("render text failed: %v", err)
		return
	}
	defer surface.Free()
	texture, err := APPCtx.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		log.Printf("create texture failed: %v", err)
		return
	}
	defer texture.Destroy()
	APPCtx.Renderer.SetClipRect(&sdl.Rect{
		X: x,
		Y: y,
		W: Cfg.Data.Width - 10,
		H: int32(surface.H),
	})
	if int32(surface.W) > int32(Cfg.Data.Width-10) {
		APPCtx.Renderer.Copy(texture, nil, &sdl.Rect{
			X: x - scrollOffset,
			Y: y,
			W: surface.W,
			H: surface.H,
		})
	} else {
		APPCtx.Renderer.Copy(texture, nil, &sdl.Rect{
			X: x,
			Y: y,
			W: surface.W,
			H: surface.H,
		})
	}
	APPCtx.Renderer.SetClipRect(nil)
}

func render() {
	// render background
	APPCtx.Renderer.Clear()
	APPCtx.Renderer.SetDrawColor(0, 0, 0, 255)
	if APPCtx.BkTexture != nil {
		APPCtx.BkTexture.SetBlendMode(sdl.BLENDMODE_MOD)
		APPCtx.BkTexture.SetAlphaMod(Cfg.Resources.Transparency)
		APPCtx.Renderer.Copy(APPCtx.BkTexture, nil, &sdl.Rect{X: 0, Y: 0, W: Cfg.Data.Width, H: Cfg.Data.Height})
	}
	// render file list
	for i := APPCtx.WindowStart; i < min(APPCtx.WindowStart+APPCtx.WindowSize-1, len(APPCtx.FileEntries)); i++ {
		e := APPCtx.FileEntries[i]
		y := int32((i - APPCtx.WindowStart) * Cfg.Data.LineHeight)
		if i == APPCtx.Selected {
			APPCtx.Renderer.SetDrawColor(20, 30, 80, 255)
			APPCtx.Renderer.FillRect(&sdl.Rect{X: 0, Y: y, W: Cfg.Data.Width, H: int32(Cfg.Data.LineHeight)})
		}
		var displayName string
		if e.IsDir {
			displayName = "[DIR]  " + e.Name
		} else {
			displayName = "[FILE] " + e.Name
		}
		renderText(APPCtx.FontFileList, 5, y, displayName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}
	// render status bar
	statusBarHeight := int32(Cfg.Data.LineHeight)
	statusBarY := int32(Cfg.Data.Height - statusBarHeight)
	APPCtx.Renderer.SetDrawColor(50, 50, 50, 255)
	APPCtx.Renderer.FillRect(&sdl.Rect{X: 0, Y: statusBarY, W: Cfg.Data.Width, H: statusBarHeight})
	if len(APPCtx.FileEntries) > 0 && APPCtx.Selected >= 0 && APPCtx.Selected < len(APPCtx.FileEntries) {
		renderTextScrolled(
			APPCtx.FontStatus,
			5,
			statusBarY+4,
			func() string {
				info, err := os.Stat(filepath.Join(APPCtx.CurrentDir, APPCtx.FileEntries[APPCtx.Selected].Name))
				if err != nil {
					log.Println("get information failed: %v", err)
				}
				statusText := ""
				if info.IsDir() {
					statusText = fmt.Sprintf("%s  Permission: %s          ", APPCtx.CurrentDir, info.Mode().Perm().String())
				} else {
					statusText = fmt.Sprintf("%s  Permission: %s  Size: %s", APPCtx.CurrentDir, info.Mode().Perm().String(), formatFileSize(info.Size()))
				}
				return statusText
			}(),
			func() int32 {
				now := time.Now().UnixMilli()
				if now-APPCtx.LastScrollTime > 30 {
					APPCtx.StatusScrollOffset += 2
					APPCtx.LastScrollTime = now
				}
				if APPCtx.StatusScrollOffset > Cfg.Data.Width {
					APPCtx.StatusScrollOffset = 0
				}
				return APPCtx.StatusScrollOffset
			}(),
			sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}
	APPCtx.Renderer.Present()
}
