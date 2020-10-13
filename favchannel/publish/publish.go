package publish

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/Arman92/go-tdlib"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/favchannel/tdlibclient"
	fileStorage "github.com/cyhalothrin/gifkoskladbot/storage"
)

const (
	CommandCollect = "collect"
	CommandPublish = "publish"
	CommandDelete  = "delete"
)

func PublishGifWithTags(command string) error {
	conf, err := config.ReadConfig()
	if err != nil {
		return err
	}

	store, err := fileStorage.NewFileMetaStorage(conf.StoragePath)
	if err != nil {
		return err
	}
	defer store.Close()

	client, err := tdlibclient.NewClient(conf.TDLib)
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Println("destroy telegram client:", err)
		}
	}()

	gifPub, err := NewGifTagsPublisher(conf, client)
	if err != nil {
		return err
	}

	switch command {
	case CommandCollect:
		return gifPub.collect()
	case CommandPublish:
		return gifPub.publishMessages(store)
	}

	return nil
}

func TestPublish() {
	conf, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := tdlibclient.NewClient(conf.TDLib)
	if err != nil {
		log.Fatal("client start", err)
	}
	defer client.Close()

	chatID := int64(-1001481153966)
	//fileID := "CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE"
	fileID := "CgACAgIAAxkBAAEDBU9fXjb76uuhZkONrEXHA3BVxb66xwAC6AIAAg0IUEuo9FFl_K-mRxgE"
	caption := "test"
	messageData := tdlib.NewInputMessageAnimation(
		tdlib.NewInputFileRemote(fileID),
		nil,
		1,
		0,
		0,
		tdlib.NewFormattedText(caption, nil),
	)
	//sendData := tdlib.UpdateData{
	//	"@type":                 "sendMessage",
	//	"chat_id":               chatID,
	//	"disable_notification":  true,
	//	"from_background":       false,
	//	"input_message_content": messageData,
	//}

	//donech := make(chan struct{})
	//go func() {
	//	ch := client.GetRawUpdatesChannel(0)
	//	for upd := range ch {
	//		fmt.Println("update", spew.Sdump(upd))
	//	}
	//}()

	//eventID := randString()

	// set @extra field
	//sendData["@extra"] = eventID
	waitMessageType := tdlib.NewUpdateMessageSendSucceeded(nil, 0)
	filterFunc := func(msg *tdlib.TdMessage) bool {
		spew.Dump(msg)
		// I have no idea why msg is ref to interface
		return true
	}
	updCh := client.AddEventReceiver(waitMessageType, filterFunc, 100)
	msg, err := client.SendMessage(
		chatID,
		0,
		true,
		false,
		nil,
		messageData,
	)
	if err != nil {
		log.Fatal("send message", err)
	}

	//fmt.Println("SendingState", msg.SendingState)
	fmt.Println("Message", msg.ID)

	for {
		select {
		case <-time.After(30 * time.Second):
			return
		case upd, ok := <-updCh.Chan:
			if !ok {
				fmt.Println("channel closed")
				return
			}
			//spew.Dump(upd)
			if msgSent, ok := upd.(*tdlib.UpdateMessageSendSucceeded); ok {
				if msgSent.OldMessageID == msg.ID {
					fmt.Println("new message id", msg.ID)

					//_, err := client.DeleteMessages(chatID, []int64{msgSent.Message.ID}, false)
					//if err != nil {
					//	log.Fatal(err)
					//}
				}
			}
		}
	}

	//filterFunc := func(msg *tdlib.TdMessage) bool {
	//	spew.Dump(msg)
	//	//fmt.Println(msg.MessageType())
	//
	//	return false
	//}
	//recv := client.AddEventReceiver(messageData, filterFunc, 1)
	//
	//for in := range recv.Chan {
	//	spew.Dump(in)
	//}

	// create waiter chan and save it in Waiters
	//waiter := make(chan tdlib.UpdateMsg, 1)
	//client.waiters.Store(randomString, waiter)

	//// send it through already implemented method
	//client.Send(update)
	//
	//select {
	//// wait response from main loop in NewClient()
	//case response := <-waiter:
	//	return response, nil
	//	// or timeout
	//case <-time.After(10 * time.Second):
	//	client.waiters.Delete(randomString)
	//	return UpdateMsg{}, errors.New("timeout")
	//}
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//if result.Data["@type"].(string) == "error" {
	//	return nil, fmt.Errorf("error! code: %d msg: %s", result.Data["code"], result.Data["message"])
	//}
	//
	//var messageDummy Message
	//err = json.Unmarshal(result.Raw, &messageDummy)
	//return &messageDummy, err

	//msg, err := client.SendMessage(
	//	chatID,
	//	0,
	//	true,
	//	false,
	//	nil,
	//	tdlib.NewInputMessageAnimation(
	//		tdlib.NewInputFileRemote(fileID),
	//		nil,
	//		1,
	//		0,
	//		0,
	//		tdlib.NewFormattedText(caption, nil),
	//	),
	//)

	//if err != nil {
	//	log.Fatalf("message not sent: %s", err)
	//}
	//
	//spew.Dump(msg)
}

func randString() string {
	// letters for generating random string
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// generate random string for @extra field
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
