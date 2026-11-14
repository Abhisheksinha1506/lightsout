package output

import "github.com/wayneashleyberry/truecolor/pkg/color"

// Title prints the LIGHTS OUT banner to standard output, shaded from red to
// orange.
func Title() {
	title := []string{
		"  ██╗     ██╗ ██████╗ ██╗  ██╗████████╗███████╗    ██████╗ ██╗   ██╗████████╗",
		"  ██║     ██║██╔════╝ ██║  ██║╚══██╔══╝██╔════╝   ██╔═══██╗██║   ██║╚══██╔══╝",
		"  ██║     ██║██║  ███╗███████║   ██║   ███████╗   ██║   ██║██║   ██║   ██║",
		"  ██║     ██║██║   ██║██╔══██║   ██║   ╚════██║   ██║   ██║██║   ██║   ██║",
		"  ███████╗██║╚██████╔╝██║  ██║   ██║   ███████║   ╚██████╔╝╚██████╔╝   ██║",
		"  ╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝    ╚═════╝  ╚═════╝    ╚═╝",
	}
	for i, value := range title {
		color.Color(216, uint8(i*30), 0).Println(value)
	}
}
