package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := &App{}

	err := wails.Run(&options.App{
		Title:                    "Moyan",
		Width:                    1120,
		Height:                   760,
		MinWidth:                 760,
		MinHeight:                520,
		BackgroundColour:         &options.RGBA{R: 245, G: 244, B: 237, A: 255},
		AssetServer:              &assetserver.Options{Assets: assets},
		OnStartup:                app.startup,
		OnShutdown:               app.shutdown,
		Bind:                     []interface{}{app},
		EnableDefaultContextMenu: false,
	})
	if err != nil {
		log.Fatal(err)
	}
}
