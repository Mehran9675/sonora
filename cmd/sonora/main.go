// Command sonora is a desktop studio for text-to-speech models.
package main

import (
	"log"

	"github.com/mehran9675/sonora/ui/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
