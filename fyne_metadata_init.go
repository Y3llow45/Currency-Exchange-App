package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func init() {
	app.SetMetadata(fyne.AppMetadata{
		ID: "",
		Name: "currency-exchange-app.exe",
		Version: "0.0.1",
		Build: 1,
		Icon: &fyne.StaticResource{
	StaticName: "Icon.png",
	StaticContent: []byte{
		}},
		
		Release: false,
		Custom: map[string]string{
			
		},
		
	})
}

