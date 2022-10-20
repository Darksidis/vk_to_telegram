package main

import (
	"context"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/anaskhan96/soup"
	"github.com/corpix/uarand"
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

	fmt.Println("app start")

	VKtoken := goDotEnvVariable("VK_TOKEN") // your group token
	standaloneToken := goDotEnvVariable("STANDALONE_TOKEN")

	vk := api.NewVK(VKtoken)

	botToken := goDotEnvVariable("BOT_TOKEN") // your bot token
	bot := initBot(botToken)
	urlTgGroup := goDotEnvVariable("URL_TG_GROUP")

	idChannelStr := goDotEnvVariable("ID_CHANNEL") // id your channel in telegram
	idChannel, err := strconv.ParseInt(idChannelStr, 10, 64)
	check(err)

	// get information about the group
	group, err := vk.GroupsGetByID(nil)
	idGroup := group[0].ID
	
	lp, err := longpoll.NewLongPoll(vk, idGroup)
	if err != nil {
		panic(err)
	}

	lp.WallPostNew(func(ctx context.Context, obj events.WallPostNewObject) {
		idAuthorPost := obj.FromID
		text := obj.Text
		attacments := obj.Attachments
		postID := obj.ID

		if idAuthorPost == idGroup*-1 {
			//if idAuthorPost does not match with idGroup, then it means that the post is from the offer, skip

			if len(attacments) != 0 {
				//if not empty, then the post contains media content

				listAttchForEditMethod, tgPostID := sendPostWithMedia(obj, idChannel, bot)

				if len(standaloneToken) != 0 && tgPostID != "" {
					vkStandalone := api.NewVK(standaloneToken)

					editPost(vkStandalone, idGroup, postID, tgPostID, text, urlTgGroup, listAttchForEditMethod)
				}

			} else {
				//if there is no media content in the post, we send only the text
				var emptyListAttach []string
				tgPostID := sendPostWithTextOnly(idChannel, bot, text)

				if len(standaloneToken) != 0 {
					vkStandalone := api.NewVK(standaloneToken)

					editPost(vkStandalone, idGroup, postID, tgPostID, text, urlTgGroup, emptyListAttach)
				}
			}

			CleaningFiles()
		}
	})

	lp.Run()

}

func check(err error) {
	// Проверка на ошибки
	if err != nil {
		panic(err)
	}
}

// func AddCapture (T comparable, ) {
//
// }
func initBot(token string) *tgbotapi.BotAPI {
	// Create bot from environment value.
	bot, err := tgbotapi.NewBotAPI(token)
	check(err)

	return bot

}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load("local.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func sendPostWithMedia(obj events.WallPostNewObject, idChannel int64, bot *tgbotapi.BotAPI) ([]string, string) {

	text := obj.Text
	var listAttchForEditMethod []string

	var inputFiles []interface{}

	mediaGroup := obj.Attachments
	count := 0
	for _, media := range mediaGroup {

		photo := media.Photo
		nameFile := strconv.Itoa(photo.ID) + ".jpg"

		if photo.ID != 0 {
			nameForEditMethod := addToListAttacments("photo", photo.OwnerID, photo.ID)
			listAttchForEditMethod = append(listAttchForEditMethod, nameForEditMethod)

			//sizes := photo.Sizes

			photoUrl := photo.MaxSize().URL
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
		file := media.Doc
		fileURL := file.URL
		fileName := file.Title
		if len(fileURL) != 0 {
			response, err := sendingRequest(fileURL)

			nameForEditMethod := addToListAttacments("doc", file.OwnerID, file.ID)
			listAttchForEditMethod = append(listAttchForEditMethod, nameForEditMethod)

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
		video := media.Video
		videoName := video.Title
		if video.ID != 0 {
			nameForEditMethod := addToListAttacments("video", video.OwnerID, video.ID)
			listAttchForEditMethod = append(listAttchForEditMethod, nameForEditMethod)

			videoURL := getUrlVideo(video.OwnerID, video.ID)
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
		postChannel, err := bot.SendMediaGroup(msg)
		check(err)

		tgPostID := strconv.Itoa(postChannel[0].MessageID)

		return listAttchForEditMethod, tgPostID

	}

	return listAttchForEditMethod, ""

}

func sendPostWithTextOnly(idChannel int64, bot *tgbotapi.BotAPI, text string) string {

	msg := tgbotapi.NewMessage(idChannel, text)
	postChannel, err := bot.Send(msg)
	check(err)
	tgPostID := strconv.Itoa(postChannel.MessageID)

	return tgPostID

}

func addToListAttacments(typeMedia string, ownerID int, mediaID int) string {

	name := typeMedia + strconv.Itoa(ownerID) + "_" + strconv.Itoa(mediaID)

	return name
}

func editPost(vk *api.VK, idGroup int, postID int, tgPostID string, text string, urlTgGroup string, attachments []string) {

	urlTgPost := urlTgGroup + "/" + tgPostID
	attachments = append(attachments, urlTgPost)
	if len(text) == 0 {
		text += "."
	}

	par := make(map[string]interface{})
	par["owner_id"] = idGroup * -1
	par["post_id"] = postID
	par["message"] = text
	par["attachments"] = attachments

	_, err := vk.WallEdit(par)
	if err != nil {
		log.Fatal(err)
	}
}

func CleaningFiles() {
	dirs, err := os.ReadDir("inputMedia")
	for _, dir := range dirs {
		path := filepath.Join("inputMedia", dir.Name())
		dirRead, err := os.ReadDir(path)
		check(err)
		for _, file := range dirRead {
			path := filepath.Join("inputMedia", dir.Name(), file.Name())
			if file.Name() != "readme.md" {
				err = os.Remove(path)

			}
		}

	}
	check(err)
}

func createInputFile(path string, nameFile string) tgbotapi.FileReader {
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
	if err != nil {
		return err
	}
	defer out.Close()

	// Getting a server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writing a file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getUrlVideo(ownerId int, videoId int) string {
	strOwnerId := strconv.Itoa(ownerId)
	strVideoId := strconv.Itoa(videoId)
	urlVideo := "https://vk.com/video" + strOwnerId + "_" + strVideoId

	request, err := sendingRequest(urlVideo)

	if err != nil {
		log.Fatal("parsing html failed: ", err)
	}
	embedUrl := ""
	tagList := []string{"link", "itemprop", "embedUrl"}
	embedUrl, err = getElementsFromHtmlPage(request, tagList)
	check(err)

	request, err = sendingRequest(embedUrl)

	if err != nil {
		log.Fatal("parsing html failed: ", err)
	}

	bodyHtml, err := getHtmlPage(request)
	check(err)

	embedUrl = findWithRegexp(bodyHtml)

	return embedUrl

}

func GetUserAgent() string {
	userAgent := uarand.GetRandom()

	return userAgent

}
func getElementsFromHtmlPage(request *http.Response, tagList []string) (string, error) {

	bodyHtml, err := getHtmlPage(request)

	if err != nil {
		return "", err
	}

	doc := soup.HTMLParse(bodyHtml)

	result, err := findElements(doc, tagList)
	if err != nil {
		return "", err
	}
	return result, nil
}

func sendingRequest(url string) (*http.Response, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", "Mozilla/8.0 (Windows NT 6.1) AppleWebKit/537.3 Chrome/49.0.2623.112 Safari/537.36")

	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("bad status: %s", resp.Status)
	}
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

func findElements(doc soup.Root, tagList []string) (string, error) {
	// We parse the markup and find the necessary elements

	elements := doc.Find(tagList...)

	if elements.Error != nil {
		return "", elements.Error
	}
	href := elements.Attrs()["href"]

	return href, nil
}

func findWithRegexp(data string) string {
	data = strings.Replace(data, "\n", "", -1)

	listSizes := [5]string{"720", "480", "360", "240", "144"}
	embedUrl := ""
	for _, size := range listSizes {
		textRegexp := fmt.Sprintf("\"url%s\":(.*?)", size)
		findUrlReg, err := regexp.Compile(textRegexp + "\"(.*?)" + "\"(.*?)")
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
