package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

// return bundled font resource
func (*myTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.LightTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.LightTheme().Font(s)
		}
		return resourceFZHTJWTTF
	}
	if s.Italic {
		return theme.LightTheme().Font(s)
	}
	return resourceFZHTJWTTF
}

func (*myTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.LightTheme().Color(n, v)
}

func (*myTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.LightTheme().Icon(n)
}

func (*myTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.LightTheme().Size(n)
}
