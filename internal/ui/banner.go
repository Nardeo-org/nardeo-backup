package ui

import (
	"fmt"
	"io"
	"os"
)

const banner = `
 _   _    _    ____  ____  _____ ___
| \ | |  / \  |  _ \|  _ \| ____/ _ \
|  \| | / _ \ | |_) | | | |  _|| | | |
| |\  |/ ___ \|  _ <| |_| | |__| |_| |
|_| \_/_/   \_\_| \_\____/|_____\___/
 ____    _    ____ _  ___   _ ____
| __ )  / \  / ___| |/ / | | |  _ \
|  _ \ / _ \| |   | ' /| | | | |_) |
| |_) / ___ \ |___| . \| |_| |  __/
|____/_/   \_\____|_|\_\\___/|_|

              NARDEO BACKUP

`

func PrintBanner(out io.Writer) {
	if file, ok := out.(*os.File); ok && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" {
		if info, err := file.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			fmt.Fprint(out, "\033[36m", banner, "\033[0m")
			return
		}
	}
	fmt.Fprint(out, banner)
}
