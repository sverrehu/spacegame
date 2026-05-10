package utils

import (
	"log"
	"os"

	"github.com/gogpu/gg/text"
)

func LoadFontSource() *text.FontSource {
	fontPath := findSystemFont()
	if fontPath == "" {
		log.Fatal("No system font found.")
		return nil
	}
	source, err := text.NewFontSourceFromFile(fontPath)
	if err != nil {
		log.Fatalf("Failed to load font %s: %v", fontPath, err)
		return nil
	}
	log.Printf("Loaded font: %s", source.Name())
	return source
}

func findSystemFont() string {
	candidates := []string{
		// Linux
		"/usr/share/fonts/truetype/dejavu/VeraMono.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
		"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		// macOS
		"/System/Library/Fonts/Supplemental/Andale Mono.ttf",
		"/System/Library/Fonts/Supplemental/Courier New.ttf",
		"/Library/Fonts/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Monaco.ttf",
		// Windows
		"C:\\Windows\\Fonts\\lucon.ttf",   // Lucida Console
		"C:\\Windows\\Fonts\\l_10646.ttf", // Lucida Sans Unicode
		"C:\\Windows\\Fonts\\cour.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
