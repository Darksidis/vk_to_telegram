package main

import (
	"context"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/anaskhan96/soup"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	VKtoken := goDotEnvVariable("VK_TOKEN") // your group token
	vk := api.NewVK(VKtoken)
	botToken := goDotEnvVariable("BOT_TOKEN") // your bot token
	bot := initBot(botToken)

	idChannelStr := goDotEnvVariable("ID_CHANNEL") // id your channel in telegram
	idChannel, err := strconv.ParseInt(idChannelStr, 10, 64)
	check(err)
	// get information about the group
	group, err := vk.GroupsGetByID(nil)

	lp, err := longpoll.NewLongPoll(vk, group[0].ID)
	if err != nil {
		panic(err)
	}

	lp.WallPostNew(func(ctx context.Context, obj events.WallPostNewObject) {
		text := obj.Text
		attacments := obj.Attachments

		if len(attacments) != 0 {
			var inputFiles []interface{}


			mediaGroup := obj.Attachments
			count := 0
			for _, media := range mediaGroup {
			
				photo := media.Photo
				nameFile := strconv.Itoa(photo.ID) + ".jpg"

				if photo.ID != 0 {
					sizes := photo.Sizes
					photoUrl := sizes[len(sizes)-1].URL
					response, err := sendingRequest(photoUrl)
					path := filepath.Join("inputMedia", "photos", nameFile)
					err = downloadFile(path, response)
					check(err)

					file := createInputFile(path, nameFile)
					input := tgbotapi.NewInputMediaPhoto(file)
					if count == 0 {
						input.Caption = text
						count += 1
					}
					inputFiles = append(inputFiles, input)

				}

				fileURL := media.Doc.URL
				fileName := media.Doc.Title
				if len(fileURL) != 0 {
					response, err := sendingRequest(fileURL)

					nameFile := fmt.Sprintf("%s.gif", fileName)
					path := filepath.Join("inputMedia", "files", nameFile)
					err = downloadFile(path, response)
					check(err)

					file := createInputFile(path, nameFile)
					input := tgbotapi.NewInputMediaDocument(file)
					if count == 0 {
						input.Caption = text
						count += 1
					}
					inputFiles = append(inputFiles, input)

					// ... скачиваем гифку и отправляем
				}

				videoOwnerID := media.Video.OwnerID
				videoID := media.Video.ID
				videoName := media.Video.Title
				if videoID != 0 {

					videoURL := getUrlVideo(videoOwnerID, videoID)
					if len(videoURL) != 0 {
						// add to failed status in post
						response, err := sendingRequest(videoURL)
						nameFile := fmt.Sprintf("%s.mp4", videoName)
						path := filepath.Join("inputMedia", "videos", nameFile)
						err = downloadFile(path, response)
						check(err)

						file := createInputFile(path, nameFile)

						input := tgbotapi.NewInputMediaVideo(file)
						if count == 0 {
							input.Caption = text
							count += 1
						}
						inputFiles = append(inputFiles, input)
					}
					// .. download video
				}

				}

			if len(inputFiles) != 0 {
				// If we got a message

				msg := tgbotapi.NewMediaGroup(idChannel, inputFiles)

				bot.Send(msg)

			
			}

		}  else {

				msg := tgbotapi.NewMessage(idChannel, text)
				bot.Send(msg)

		}

		CleaningFiles ()
	})

	lp.Run()

}

func check(err error) {
	// Проверка на ошибки
	if err != nil {
		panic(err)
	}
}
//func AddCapture (T comparable, ) {
//
//}
func initBot (token string) *tgbotapi.BotAPI{
	// Create bot from environment value.
	bot, err := tgbotapi.NewBotAPI(token)
	check(err)

	return bot

}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}


func CleaningFiles () {
	dirs, err := os.ReadDir("inputMedia")
	for _, dir := range dirs {
		path := filepath.Join("inputMedia", dir.Name())
		dirRead, err := os.ReadDir(path)
		check(err)
		for _, file := range dirRead {
			path := filepath.Join("inputMedia",  dir.Name(), file.Name())
			err = os.Remove(path)
		}


	}
	check(err)
}

func createInputFile (path string, nameFile string) tgbotapi.FileReader{
	inputFile, err := os.Open(path)
	check(err)
	file := tgbotapi.FileReader{
		Name:   nameFile,
		Reader: inputFile,
	}
	//err = os.Remove(path)
	//check(err)
	return file
}

func downloadFile(filepath string, resp *http.Response) (err error) {
	// In this function, we are downloading a file from a link
	// Create file
	out, err := os.Create(filepath)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Getting a server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writing a file
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		return err
	}

	return nil
}

func getUrlVideo (ownerId int, videoId int) string{
	strOwnerId := strconv.Itoa(ownerId)
	strVideoId := strconv.Itoa(videoId)
	urlVideo := "https://vk.com/video" + strOwnerId + "_" + strVideoId

	request, err := sendingRequest(urlVideo)

	if err != nil {
		log.Fatal("parsing html failed: ", err)
	}

	bodyHtml, err := getHtmlPage(request)

	doc := soup.HTMLParse(bodyHtml)

	embedUrl := ""
	tagList := []string{"link", "itemprop", "embedUrl"}
	embedUrl, err = findElements(doc, tagList)
	check(err)

	request, err = sendingRequest(embedUrl)

	if err != nil {
		log.Fatal("parsing html failed: ", err)
	}

	bodyHtml, err = getHtmlPage(request)

	embedUrl = findWithRegexp(bodyHtml)

	return embedUrl

}

func sendingRequest (url string) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func getHtmlPage(resp *http.Response) (string, error) {
	// Get page content and read

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {

		return "", err
	}

	defer resp.Body.Close()

	return string(body), nil

}

func findElements (doc soup.Root, tagList []string) (string, error) {
	// We parse the markup and find the necessary elements

	elements:= doc.Find(tagList ...)

	if elements.Error != nil {
		return "", elements.Error
	}
	href := elements.Attrs()["href"]

	return href, nil
}

func findWithRegexp (data string) string{
	data = strings.Replace(data,"\n", "", -1)

	listSizes := [5]string{"720", "480", "360", "240", "144"}
	embedUrl := ""
	for _, size := range listSizes {
		textRegexp := fmt.Sprintf("\"url%s\":(.*?)", size)
		findUrlReg, err :=  regexp.Compile(textRegexp  + "\"(.*?)" + "\"(.*?)")
		check(err)
		embedUrl = findUrlReg.FindString(data)

		if len(embedUrl) != 0 {
			sizeReplace := fmt.Sprintf("\"url%s\":", size)
			embedUrl = strings.Replace(embedUrl, sizeReplace, "", -1)
			embedUrl = strings.Replace(embedUrl, "\"", "", -1)
			embedUrl = strings.Replace(embedUrl, "\\/", "//", 1)
			embedUrl = strings.Replace(embedUrl, "\\/", "", 1)
			embedUrl = strings.Replace(embedUrl, "\\/", "//", 1)

			break
		}
	}

	return embedUrl

}

