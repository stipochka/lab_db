package main

import window "lab_db/internal/window"

func main() {
	win := window.New()
	win.SetWidgets()
	win.StartWindow("file db GUI")
}
