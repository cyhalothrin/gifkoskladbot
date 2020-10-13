package tdlibclient

import (
	"fmt"
	"log"
	"time"

	"github.com/cyhalothrin/gifkoskladbot/config"

	"github.com/Arman92/go-tdlib"
)

func authorize(client *TdLibClient, conf config.TDLibClient) error {
	phoneNumber := conf.Phone

	for {
		currentState, _ := client.Authorize()
		authState := currentState.GetAuthorizationStateEnum()
		switch authState {
		case tdlib.AuthorizationStateWaitPhoneNumberType:
			fmt.Printf("Sending phone %s...\n", phoneNumber)
			_, err := client.SendPhoneNumber(phoneNumber)
			if err != nil {
				return fmt.Errorf("sending phone number: %w", err)
			}
		case tdlib.AuthorizationStateWaitCodeType:
			fmt.Print("Enter code: ")
			var code string
			fmt.Scanln(&code)
			_, err := client.SendAuthCode(code)
			if err != nil {
				return fmt.Errorf("sending auth code : %w", err)
			}
		case tdlib.AuthorizationStateReadyType:
			fmt.Println("Authorization Ready! Let's rock")

			return nil
		default:
			log.Printf("Unknown auth state: %s\n", authState)
			time.Sleep(500 * time.Millisecond)
		}
	}
}
