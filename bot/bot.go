package bot

import (
	"encoding/json"
	"log"

	"github.com/HirbodBehnam/RedditDownloaderBot/cache"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// RunBot runs the bot with the specified token
func RunBot(token string, allowedUsers AllowedUsers) {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Cannot initialize the bot: ", err.Error())
	}
	log.Println("Bot authorized on account", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.CallbackQuery != nil {
			go handleCallback(update.CallbackQuery.Data, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			continue
		}
		if update.Message == nil {
			continue
		}
		if !allowedUsers.IsAllowed(update.Message.From.ID) {
			// You might want to implement a logic here
			continue
		}
		// Only text messages are allowed
		if update.Message.Text == "" {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please send the reddit post link to bot"))
			continue
		}
		// Check if the message is command
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Hello and welcome!\nJust send me the link of the post to download it for you."))
			case "about":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Reddit Downloader Bot v"+config.Version+"\nBy Hirbod Behnam\nSource: https://github.com/HirbodBehnam/RedditDownloaderBot"))
			case "help":
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Just send me the link of the reddit post or comment. If it's text, I will send the text of the post. If it's a photo or video, I will send the it with the title as caption."))
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry this command is not recognized; Try /help"))
			}
			continue
		}
		go fetchPostDetailsAndSend(update.Message.Text, update.Message.Chat.ID, update.Message.MessageID)
	}
}

// fetchPostDetailsAndSend gets the basic info about the post being sent to us
func fetchPostDetailsAndSend(text string, chatID int64, messageID int) {
	result, fetchErr := RedditOauth.StartFetch(text)
	if fetchErr != nil {
		msg := tgbotapi.NewMessage(chatID, fetchErr.BotError)
		msg.ReplyToMessageID = messageID
		bot.Send(msg)
		if fetchErr.NormalError != "" {
			log.Println("cannot get post ", text, ":", fetchErr.NormalError)
		}
		return
	}
	// Check the result type
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ReplyToMessageID = messageID
	msg.ParseMode = Markdown
	switch data := result.(type) {
	case reddit.FetchResultText:
		msg.Text = data.Title + "\n" + data.Text
	case reddit.FetchResultComment:
		msg.Text = data.Text
	case reddit.FetchResultMedia:
		if len(data.Medias) == 0 {
			msg.Text = "No media found!"
			break
		}
		// If there is one media quality, download it
		// Also allow the user to choose between photo or document in image
		if len(data.Medias) >= 1 {
			switch data.Type {
			case reddit.FetchResultMediaTypePhoto:
				handlePhotoUpload(data.Medias[0].Link, data.Title, data.ThumbnailLink, chatID, true)
				return
			case reddit.FetchResultMediaTypeGif:
				handleGifUpload(data.Medias[0].Link, data.Title, data.ThumbnailLink, chatID)
				return
			case reddit.FetchResultMediaTypeVideo:
				// If the video does have an audio, ask user if they want the audio
				if _, hasAudio := data.HasAudio(); !hasAudio {
					// Otherwise, just download the video
					handleVideoUpload(data.Medias[0].Link, data.Title, data.ThumbnailLink, data.Duration, chatID)
					return
				}
			}
		}
		// Allow the user to select quality
		msg.Text = "Please select the quality"
		idString := util.UUIDToBase64(uuid.New())
		audioIndex, _ := data.HasAudio()
		switch data.Type {
		case reddit.FetchResultMediaTypePhoto:
			msg.ReplyMarkup = createPhotoInlineKeyboard(idString, data)
		case reddit.FetchResultMediaTypeGif:
			msg.ReplyMarkup = createGifInlineKeyboard(idString, data)
		case reddit.FetchResultMediaTypeVideo:
			msg.ReplyMarkup = createVideoInlineKeyboard(idString, data)
		}
		// Insert the id in cache
		err := CallbackCache.SetMediaCache(idString, cache.CallbackDataCached(CallbackDataCached{
			Links:         data.Medias.ToLinkMap(),
			Title:         data.Title,
			ThumbnailLink: data.ThumbnailLink,
			Type:          data.Type,
			Duration:      data.Duration,
			AudioIndex:    audioIndex,
		}))
		if err != nil {
			log.Println("Cannot set the media cache in database:", err)
		}
	case reddit.FetchResultAlbum:
		idString := util.UUIDToBase64(uuid.New())
		err := CallbackCache.SetAlbumCache(idString, data)
		if err != nil {
			log.Println("Cannot set the album cache in database:", err)
		}
		msg.Text = "Download album as media or files?"
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{
				tgbotapi.NewInlineKeyboardButtonData("Media", CallbackButtonData{
					ID:   idString,
					Mode: CallbackButtonDataModePhoto,
				}.String()),
				tgbotapi.NewInlineKeyboardButtonData("File", CallbackButtonData{
					ID:   idString,
					Mode: CallbackButtonDataModeFile,
				}.String()),
			}},
		}
	default:
		log.Printf("unknown type: %T\n", result)
		msg.Text = "unknown type (report please)"
	}
	_, err := bot.Send(msg)
	if err != nil {
		msg.ParseMode = "" // fall back and don't format message
		bot.Send(msg)
	}
}

// handleCallback handles the callback query of selecting a quality for any media type
func handleCallback(dataString string, chatID int64, msgId int) {
	// Don't crash!
	defer func() {
		if r := recover(); r != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Cannot get data. (panic)"))
			log.Println("recovering from panic:", r)
		}
	}()
	// Delete the message
	bot.Send(tgbotapi.NewDeleteMessage(chatID, msgId))
	// Parse the data
	var data CallbackButtonData
	err := json.Unmarshal(util.StringToByte(dataString), &data)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Broken callback data"))
		return
	}
	// Get the cache from database
	cachedData, err := CallbackCache.GetAndDeleteMediaCache(data.ID)
	if err == cache.NotFoundErr {
		// Check albums
		var album reddit.FetchResultAlbum
		album, err = CallbackCache.GetAndDeleteAlbumCache(data.ID)
		if err == nil {
			handleAlbumUpload(album, chatID, data.Mode == CallbackButtonDataModeFile)
			return
		} else if err == cache.NotFoundErr {
			// It does not exists...
			bot.Send(tgbotapi.NewMessage(chatID, "Please resend the link to bot"))
			return
		}
		// Fall to report internal error
	}
	// Check other errors
	if err != nil {
		log.Println("Cannot get callback id from database:", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Internal error"))
		return
	}
	// Check the link
	link, exists := cachedData.Links[data.LinkKey]
	if !exists {
		bot.Send(tgbotapi.NewMessage(chatID, "Please resend the link to bot"))
		return
	}
	// Check the media type
	switch cachedData.Type {
	case reddit.FetchResultMediaTypeGif:
		handleGifUpload(link, cachedData.Title, cachedData.ThumbnailLink, chatID)
	case reddit.FetchResultMediaTypePhoto:
		handlePhotoUpload(link, cachedData.Title, cachedData.ThumbnailLink, chatID, data.Mode == CallbackButtonDataModePhoto)
	case reddit.FetchResultMediaTypeVideo:
		if data.LinkKey == cachedData.AudioIndex {
			handleAudioUpload(link, cachedData.Title, cachedData.Duration, chatID)
		} else {
			handleVideoUpload(link, cachedData.Title, cachedData.ThumbnailLink, cachedData.Duration, chatID)
		}
	}
}
