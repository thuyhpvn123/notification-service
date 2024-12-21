package network

import (
	"context"
	"fmt"
	"time"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	e_common "github.com/ethereum/go-ethereum/common"
	"github.com/meta-node-blockchain/meta-node/cmd/client"
	"github.com/meta-node-blockchain/meta-node/pkg/logger"
	"github.com/meta-node-blockchain/meta-node/types"
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"github.com/meta-node-blockchain/noti-contract/internal/usecase"
	"github.com/meta-node-blockchain/noti-contract/internal/utils"
	"github.com/meta-node-blockchain/noti-contract/pkg/apns"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"github.com/meta-node-blockchain/noti-contract/pkg/fcm"
)

type NotiHandler struct {
	config                   *config.AppConfig
	chainClient              *client.Client
	notiSmartContractAddress e_common.Address
	notiABI                  *abi.ABI
	// notiUsecase              usecase.NotificationUsecase
	deviceTokenUsecase usecase.DeviceTokenUsecase
	privateKeyPem      string
}

func NewNotiEventHandler(
	config *config.AppConfig,
	chainClient *client.Client,
	address e_common.Address,
	abi *abi.ABI,
	// notiUsecase usecase.NotificationUsecase,
	deviceTokenUsecase usecase.DeviceTokenUsecase,
	privateKeyPem string,
) *NotiHandler {
	return &NotiHandler{
		config:                   config,
		chainClient:              chainClient,
		notiSmartContractAddress: address,
		notiABI:                  abi,
		// notiUsecase:              notiUsecase,
		deviceTokenUsecase: deviceTokenUsecase,
		privateKeyPem:      privateKeyPem,
	}
}

func (h *NotiHandler) HandleConnectSmartContract(events types.EventLogs) {
	for _, event := range events.EventLogList() {
		switch event.Topics()[0] {
		case h.notiABI.Events["DeviceTokenRegistered"].ID.String()[2:]:
			h.handleDeviceTokenRegistered(event.Data())
		case h.notiABI.Events["NotificationSent"].ID.String()[2:]:
			h.handlePublishNotification(event.Data())
		}
	}
}
func (h *NotiHandler) handleDeviceTokenRegistered(data string) {
	fmt.Println("data:",data)
	result := make(map[string]interface{})
	err := h.notiABI.UnpackIntoMap(result, "DeviceTokenRegistered", e_common.FromHex(data))
	if err != nil {
		logger.Error("can't unpack to map", err)
		return
	}
	fmt.Println("result:", result)
	deviceToken := model.DeviceToken{
		DAppAddress:    result["dapp"].(common.Address).Hex(),
		UserAddress:    result["user"].(common.Address).Hex(),
		EncryptedToken: result["encryptedToken"].(string),
		Platform:       result["platform"].(uint8),
		CreatedAt:      uint64(time.Now().Unix()),
	}
	histories, err := h.deviceTokenUsecase.GetEncryptedTokensByDappAndUser(result["dapp"].(common.Address).Hex(), result["user"].(common.Address).Hex())
	if err != nil {
		logger.Error("fail in get Encrypted Token by dapp and user:", err)
		return
	}
	if len(histories) == 0 {
		err = h.deviceTokenUsecase.Insert(deviceToken)
		if err != nil {
			msg := fmt.Sprintf("Unable to store device token from user %s, %v", result["user"].(string), err)
			logger.Error(msg)
		}
	}else{
		var count uint
		for _,v := range histories{
			if v.Platform == result["platform"].(uint8) {
				count ++
				if v.EncryptedToken != result["encryptedToken"].(string){					
					deviceToken.ID = v.ID
					err = h.deviceTokenUsecase.Update(deviceToken)
					if err != nil {
						msg := fmt.Sprintf("Unable to update device token from user %s, %v", result["user"].(string), err)
						logger.Error(msg)
					}

				}else{
					logger.Error("same token device existed")
				}
				break				
			}
		}
		if count == 0 {
			err = h.deviceTokenUsecase.Insert(deviceToken)
			if err != nil {
				msg := fmt.Sprintf("Unable to insert device token from user %s, %v", result["user"].(string), err)
				logger.Error(msg)
			}
		}
	}

}
func (h *NotiHandler) handlePublishNotification(data string) {
	var result model.NotiEvent
	var histories []*model.DeviceToken
	err := h.notiABI.UnpackIntoInterface(&result, "NotificationSent", e_common.FromHex(data))
	if err != nil {
		logger.Error("can't unpack to interface handlePublishNotification", err)
		return
	}
	fmt.Println("result:", result)
	if !result.SystemApp {
		histories, err = h.deviceTokenUsecase.GetEncryptedTokensByDappAndUser(result.Dapp.Hex(), result.User.Hex())
		if err != nil {
			logger.Error("fail in get Encrypted Token by dapp and user:", err)
			return
		}
	}else{
		histories, err = h.deviceTokenUsecase.GetEncryptedTokensByUser(result.User.Hex())
		if err != nil {
			logger.Error("fail in get Encrypted Token by dapp and user:", err)
			return
		}
	}
	for _,v := range histories{
		if v.EncryptedToken != "" {
			deviceToken, err := utils.DecryptToken(v.EncryptedToken, h.privateKeyPem)
			if err != nil {
				logger.Error("fail in decrypt token:", err)
				return
			}
			h.handlePublish(result,deviceToken,v.Platform)
		}else{
			logger.Error(fmt.Printf("can not find encryptedToken in db with user %v \n",result.User.Hex()))
		}

	}
}
func(h *NotiHandler) handlePublish (result model.NotiEvent,deviceToken string,platform uint8){
switch platform {
	case uint8(ANDROID):
		fcm.PushAndroidNotification(&result, deviceToken)
	case uint8(IOS):
		apns.PushIosNotification(context.Background(), h.config, &apns.PushNotification{
			// ID:      result.EventId.String(),
			Topic:   h.config.APNSTopic,
			Tokens:  []string{deviceToken},
			Message: result.Body,
			Title:   result.Title,
			// Data: map[string]interface{}{
			// 	"repo": result.Dapp,
			// },
		})
	case uint8(WEB):
		fcm.PushWebNotification(&result, deviceToken)
	default:
		logger.Error("Invalid device type")
	}
}