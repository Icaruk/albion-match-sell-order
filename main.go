package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/draw"

	"github.com/atotto/clipboard"
	"github.com/kbinani/screenshot"
)

func main() {

	// Read json file
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error: invalid config.json")
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	type Config struct {
		Apikey      string `json:"apikey"`
		StartX      int    `json:"startX"`
		StartY      int    `json:"startY"`
		SizeX       int    `json:"sizeX"`
		SizeY       int    `json:"sizeY"`
		DeleteImage bool   `json:"deleteImage"`
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Error: invalid config.json")
	}

	const monitorIdx = 0

	apikey := config.Apikey
	startX := config.StartX
	startY := config.StartY
	sizeX := config.SizeX
	sizeY := config.SizeY
	deleteImage := config.DeleteImage

	if len(apikey) == 0 {
		fmt.Println("Error: apikey is empty")
		return
	}

	fileName := "img.png"
	resizedFilename := "img_resized.png"

	if deleteImage {
		defer os.Remove(fileName)
		defer os.Remove(resizedFilename)
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
	output, _ := os.Create(resizedFilename)
	defer output.Close()
	png.Encode(output, rgba)

	fmt.Println("Capturing screen...")
	fmt.Printf("    #screen %d : %v \"%s\"\n", monitorIdx, bounds, fileName)

	// ---------------------------------
	// OCR
	// ---------------------------------

	fmt.Println("Reading image...")

	const apiurl string = "https://api.ocr.space/parse/image"

	// curl -H "apikey:helloworld" --form "base64Image=data:image/jpeg;base64,/9j/AAQSk [Long string here ]" --form "language=eng" --form "isOverlayRequired=false" https://api.ocr.space/parse/image
	var requestBody bytes.Buffer
	requestWriter := multipart.NewWriter(&requestBody)

	// add image file to request body
	imageFile, err := os.Open(resizedFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer imageFile.Close()

	imageContents, err := ioutil.ReadAll(imageFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	imageField, err := requestWriter.CreateFormFile("base64Image", "image.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	imageField.Write(imageContents)

	// add form data fields to request body
	requestWriter.WriteField("apikey", apikey)
	requestWriter.WriteField("language", "eng")
	requestWriter.WriteField("isOverlayRequired", "false")
	requestWriter.WriteField("isCreateSearchablePdf", "false")

	requestWriter.Close()

	// create HTTP request with the request body
	req, err := http.NewRequest("POST", apiurl, &requestBody)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", requestWriter.FormDataContentType())

	// send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// ---------------------------------
	// Read OCR response
	// ---------------------------------

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
		{
			"ParsedResults": [
				{
					"TextOverlay": {
						"Lines": [],
						"HasOverlay": false,
						"Message": "Text overlay is not provided as it is not requested"
					},
					"TextOrientation": "0",
					"FileParseExitCode": 1,
					"ParsedText": "1 39.999\r\n",
					"ErrorMessage": "",
					"ErrorDetails": ""
				}
			],
			"OCRExitCode": 1,
			"IsErroredOnProcessing": false,
			"ProcessingTimeInMilliseconds": "937",
			"SearchablePDFURL": "Searchable PDF not generated as it was not requested."
		}
	*/

	var jsonData map[string]interface{}
	json.Unmarshal(responseBody, &jsonData)

	// Get ParsedResults[0].ParsedText
	parsedText := jsonData["ParsedResults"].([]interface{})[0].(map[string]interface{})["ParsedText"].(string)
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
		// Ask to press "Y/n key"
		fmt.Println("Execute again? y/n")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(response)

		allowedResponses := map[string]string{
			"y": "y",
			"n": "n",
		}

		if allowedResponses[response] == "" {
			fmt.Println("❌ Invalid response")
		}

		if response == "n" {
			break
		}

		if response == "y" {
			main()
			break
		}
	}

}
