package bot

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/storage"
)

type UpdatesHandler struct {
	api     telegramBotAPI
	conf    config.Config
	storage GifkoskladMetaStorage
	alert   alerter
	// animationsNewCaptions список сообщений для отправки, по fileID
	animationsNewCaptions map[string]*storage.SentAnimation
	sentAnimations        map[string]*storage.SentAnimation
	allowedUsers          map[string]bool
	tagsAliases           map[string]string
	// uniqueTags уникальные теги, сюда будут добавляться новые
	uniqueTags map[string]bool
	// hasTagsListChanges были ли добавлены новые теги в uniqueTags
	hasTagsListChanges bool
}

func NewUpdatesHandler(
	conf config.Config,
	store GifkoskladMetaStorage,
	alert alerter,
	tgAPI telegramBotAPI,
) *UpdatesHandler {
	aliases := store.GetTagsAliases()
	if aliases == nil {
		aliases = make(map[string]string)
	}

	allowedUsers := make(map[string]bool)
	for _, username := range conf.AllowedUsers {
		allowedUsers[username] = true
	}

	sentAnimations := store.GetSentAnimations()
	if sentAnimations == nil {
		sentAnimations = make(map[string]*storage.SentAnimation)
	}

	uniqueTags := make(map[string]bool)
	for _, tag := range store.GetTags() {
		uniqueTags[tag] = true
	}

	return &UpdatesHandler{
		api:                   tgAPI,
		conf:                  conf,
		storage:               store,
		alert:                 alert,
		animationsNewCaptions: make(map[string]*storage.SentAnimation),
		tagsAliases:           aliases,
		allowedUsers:          allowedUsers,
		sentAnimations:        sentAnimations,
		uniqueTags:            uniqueTags,
	}
}

func (u *UpdatesHandler) HandleUpdates(updates []tgbotapi.Update) error {
	handlers := []updateHandler{
		u.handleAnimationCaption,
	}

	if len(updates) == 0 {
		log.Println("Нет обновлений")

		return nil
	}

	for _, update := range updates {
		for _, handler := range handlers {
			ok, err := handler(update)
			if err != nil {
				u.sendMeError(err)

				break // дальше не запускаем обработчики
			}

			if ok {
				break
			}
		}
	}

	u.PublishAnimations()
	return u.UpdateTagsList()
}

// handleAnimationCaption вся суть бота. Предполагаю что ему будут отправляться гифки и реплей на них с подписью
// дальше он сам все разрулит
// вернет false если сообщение не обрабатывается этим методом
func (u *UpdatesHandler) handleAnimationCaption(update tgbotapi.Update) (bool, error) {
	message := update.Message

	if update.Message == nil {
		if update.EditedMessage == nil {
			return false, nil
		}
		message = update.EditedMessage
	}

	if message.From == nil || !u.allowedUsers[message.From.UserName] {
		// возможно это не тот тип сообщения, отправим дальше на обработку
		return false, nil
	}

	if message.ReplyToMessage == nil {
		// возможно это не тот тип сообщения, отправим дальше на обработку
		return false, nil
	}

	animation := message.ReplyToMessage.Animation
	// ждем тут гифку
	if animation == nil {
		// возможно это не тот тип сообщения, отправим дальше на обработку
		return false, nil
	}

	text := message.Text
	if text == "" {
		return false, nil
	}

	fmt.Println(animation.FileID, text)

	tags := u.parseTags(strings.ToLower(text))
	if u.AddAnimationWithTags(animation.FileID, tags) {
		log.Printf("%s => %v\n", text, tags)
	}

	return true, nil
}

func (u *UpdatesHandler) AddAnimationWithTags(fileID string, tags []string) bool {
	id := 0
	sentMsg := u.sentAnimations[fileID]
	if sentMsg != nil {
		if u.captionsIsEqual(sentMsg.Tags, tags) {
			// если было предыдущее сообщение с другими тегами, а потом было отредактировано, но в этот виде
			// было сохранено в базе, то почистим все что сюда попало
			// была такая бага
			delete(u.animationsNewCaptions, fileID)

			log.Printf("Нет изменений '%s' (fileID: %s)\n", strings.Join(tags, " "), fileID)
			// к этому файлу уже было отправлены теги и не изменились
			return false
		}

		log.Printf(
			"Обновлены теги '%s' => '%s' (fileID: %s)\n",
			strings.Join(sentMsg.Tags, " "),
			strings.Join(tags, " "),
			fileID,
		)

		id = sentMsg.MessageID
	}

	u.animationsNewCaptions[fileID] = &storage.SentAnimation{
		FileID:    fileID,
		Tags:      tags,
		MessageID: id,
	}

	return true
}

func (u *UpdatesHandler) PublishAnimations() {
	if len(u.animationsNewCaptions) == 0 {
		return
	}

	var wg sync.WaitGroup

	for _, msg := range u.animationsNewCaptions {
		wg.Add(1)
		go u.sendAnimation(msg, &wg) // TODO: check it was sent without error
	}

	wg.Wait()

	u.storage.AddSentAnimations(u.animationsNewCaptions)
	// добавим в уже отправленные, а список новых сбросим
	for k, v := range u.animationsNewCaptions {
		u.sentAnimations[k] = v
		u.addTagsToList(v.Tags)
	}
	u.animationsNewCaptions = make(map[string]*storage.SentAnimation)
}

func (u *UpdatesHandler) sendAnimation(msg *storage.SentAnimation, wg *sync.WaitGroup) {
	defer wg.Done()

	var err error
	caption := strings.Join(msg.Tags, " ")
	tryToSendNew := false

	if msg.MessageID != 0 {
		// если ранее отправляли, то отредактируем сообщение
		log.Printf("Теги отредактированы '%s' (fileID: %s)\n", caption, msg.FileID)

		err = u.api.EditMessage(u.conf.ChannelID, msg.MessageID, caption)
		if err != nil && strings.Contains(err.Error(), "message to edit not found") {
			// сообщение из канала было удалено
			err = nil
			tryToSendNew = true
		}
	} else {
		tryToSendNew = true
	}

	if tryToSendNew {
		log.Printf("Новая гифка '%s' (fileID: %s)\n", caption, msg.FileID)

		msg.MessageID, err = u.api.SendAnimation(u.conf.ChannelID, msg.FileID, caption)
	}

	if err != nil {
		u.sendMeError(fmt.Errorf("отправка гифки '%s': %w", caption, err))
	}
}

func (u *UpdatesHandler) captionsIsEqual(tagsA, tagsB []string) bool {
	if len(tagsA) != len(tagsB) {
		return false
	}

	for _, tagA := range tagsA {
		found := false
		for _, tagB := range tagsB {
			if tagA == tagB {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (u *UpdatesHandler) UpdateTagsList() error {
	if !u.hasTagsListChanges {
		return nil
	}

	text := u.createTagsList()

	msgID, err := u.api.GetChatPinnedMessageID(u.conf.ChannelID)
	if err != nil {
		return fmt.Errorf("получение закрепленного сообщения чата: %w", err)
	}

	if msgID != 0 {
		err := u.api.EditMessage(u.conf.ChannelID, msgID, text)
		if err != nil {
			return fmt.Errorf("редактирование списка тегов: %w", err)
		}
	} else {
		// нет запиненного сообщения, создадим новое и запиним
		newID, err := u.api.SendMessage(u.conf.ChannelID, text)
		if err != nil {
			return fmt.Errorf("отправка списка тегов: %w", err)
		}
		if err := u.api.PinMessage(u.conf.ChannelID, newID); err != nil {
			return fmt.Errorf("пин сообщения #%d: %w", newID, err)
		}
	}

	u.hasTagsListChanges = false

	log.Printf("Обновил список тегов:\n%s\n", text)

	return nil
}

func (u *UpdatesHandler) createTagsList() string {
	list := make([]string, 0, len(u.uniqueTags))
	for tag := range u.uniqueTags {
		list = append(list, tag)
	}
	// отсортируем
	sort.Strings(list)

	u.storage.SetTags(list)

	return strings.Join(list, "\n")
}

// parseTags спарист теги из текста
func (u *UpdatesHandler) parseTags(text string) []string {
	words := strings.Split(text, " ")
	var tags []string
	var tag string
	var sep string
	waitEnd := false

	for _, word := range words {
		if word == "" {
			continue
		}

		if waitEnd {
			if len(word) < 2 {
				tag += sep + word
				continue
			}
			last2 := word[len(word)-2:]
			if last2 == "00" || last2 == "11" {
				word = word[:len(word)-2]
				if word != "" {
					tag += sep + word
				}
				tags = append(tags, tag)
				tag = ""
				waitEnd = false
			} else {
				tag += sep + word
			}

			continue
		}

		if len(word) < 2 {
			tags = append(tags, "#"+word)
			continue
		}

		first2 := word[:2]
		if first2 == "00" {
			tag = word[2:]
			sep = " "
			waitEnd = true
		} else if first2 == "11" {
			tag = "#" + word[2:]
			sep = "_"
			waitEnd = true
		} else {
			tags = append(tags, "#"+word)

			continue
		}

		// однако если конец этого же слова такой же, то просто обрежем и запишем
		// например случай не тега в одно слово 00описание00
		if waitEnd && first2 == tag[len(tag)-2:] {
			tags = append(tags, tag[:len(tag)-2])
			waitEnd = false
			tag = ""
		}
	}

	if tag != "" {
		tags = append(tags, tag)
	}

	// заменим алиасы
	for i, tag := range tags {
		if alias, ok := u.tagsAliases[tag]; ok {
			tags[i] = alias
		}
		// Убрать отюда нахер		u.addTagToList(tags[i])
	}

	return tags
}

func (u *UpdatesHandler) addTagsToList(tags []string) {
	for _, tag := range tags {
		if strings.Contains(tag, "#") {
			if !u.uniqueTags[tag] {
				u.uniqueTags[tag] = true
				u.hasTagsListChanges = true
			}
		}
	}
}

func (u *UpdatesHandler) sendMeError(err error) {
	if err == nil {
		return
	}

	if alertErr := u.alert.Send(err); alertErr != nil {
		log.Println("send alert error:", err)
	}
}

type updateHandler func(update tgbotapi.Update) (bool, error)
