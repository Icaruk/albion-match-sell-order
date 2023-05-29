package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/draw"

	"github.com/atotto/clipboard"
	"github.com/kbinani/screenshot"
)

func main() {

	// Read json file
	jsonFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error: invalid config.json")
	}

	type Config struct {
		StartX      int  `json:"startX"`
		StartY      int  `json:"startY"`
		SizeX       int  `json:"sizeX"`
		SizeY       int  `json:"sizeY"`
		DeleteImage bool `json:"deleteImage"`
	}

	var config Config
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		fmt.Println("Error: invalid config.json")
	}

	const monitorIdx = 0

	startX := config.StartX
	startY := config.StartY
	sizeX := config.SizeX
	sizeY := config.SizeY
	deleteImage := config.DeleteImage

	fileName := "img.png"
	// fileName := "img_debug.png" //? Debug

	if deleteImage {
		defer os.Remove(fileName)
	}

	// Screenshot
	bounds := image.Rect(startX, startY, startX+sizeX, startY+sizeY)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}

	// ---------------------------------
	// Scale up image
	// ---------------------------------

	rgba := image.NewRGBA(image.Rect(
		0,
		0,
		img.Bounds().Max.X*3,
		img.Bounds().Max.Y*3,
	))

	// Resize:
	draw.CatmullRom.Scale(rgba, rgba.Rect, img, img.Bounds(), draw.Over, nil)

	// Encode to `output`:
	output, _ := os.Create(fileName)
	defer output.Close()
	png.Encode(output, rgba)

	fmt.Println("Capturing screen...")
	fmt.Printf("    #screen %d : %v \"%s\"\n", monitorIdx, bounds, fileName)

	// ---------------------------------
	// OCR
	// ---------------------------------

	fmt.Println("Reading image...")

	// tesseract -l eng .\img_debug.png stdout
	cmd := exec.Command("tesseract", "-l", "eng", fileName, "stdout")
	cmdOutput, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	parsedText := string(cmdOutput)
	fmt.Println("    ParsedText ->", parsedText)

	// Replace regex
	regex := regexp.MustCompile(`[^0-9]`)
	parsedText = regex.ReplaceAllString(parsedText, "")

	// Cast to int
	parsedInt, _ := strconv.Atoi(parsedText)
	parsedInt--

	// Cast to string
	parsedText = strconv.Itoa(parsedInt)

	if parsedInt == -1 {
		fmt.Println("    ❌ Error, could not parse text")
	} else {
		clipboard.WriteAll(parsedText)
		fmt.Println("    ✅ Copied to clipboard:", parsedText)
	}

	fmt.Println("_______________________________")

	for {
		fmt.Println("Press ENTER to execute again or CTRL+C to quit")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(response)

		main()

	}

}
