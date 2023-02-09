/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-30
* Time: 12:27
 */

package websocket

import (
	"errors"
	"fmt"
	"gowebsocket/lib/cache"
	"gowebsocket/models"
	"gowebsocket/servers/grpcclient"
	"time"

	"github.com/go-redis/redis"
)

// 查询所有用户
// 个人疑问，从redis获取的结果不能用于判断所有用户吗？
func UserList(appId uint32) (userList []string) {

	userList = make([]string, 0)
	currentTime := uint64(time.Now().Unix())
	// 获取redis中注册，且活跃状态中的服务，活跃状态由心跳机制判断
	servers, err := cache.GetServerAll(currentTime)
	if err != nil {
		fmt.Println("给全体用户发消息", err)

		return
	}
	// 因用户上线时，会与不同的服务建立链接，所以循环遍历所有服务，获取机器信息，本地机器直接获取用户列表，非本地机器rpc获取用户列表
	for _, server := range servers {
		var (
			list []string
		)
		if IsLocal(server) {
			list = GetUserList(appId)
		} else {
			list, _ = grpcclient.GetUserList(server, appId)
		}
		userList = append(userList, list...)
	}

	return
}

// 查询用户是否在线
func CheckUserOnline(appId uint32, userId string) (online bool) {
	// 全平台查询
	if appId == 0 {
		for _, appId := range GetAppIds() {
			online, _ = checkUserOnline(appId, userId)
			if online == true {
				break
			}
		}
	} else {
		online, _ = checkUserOnline(appId, userId)
	}

	return
}

// 查询用户 是否在线
func checkUserOnline(appId uint32, userId string) (online bool, err error) {
	key := GetUserKey(appId, userId)
	userOnline, err := cache.GetUserOnlineInfo(key)
	if err != nil {
		if err == redis.Nil {
			fmt.Println("GetUserOnlineInfo", appId, userId, err)

			return false, nil
		}

		fmt.Println("GetUserOnlineInfo", appId, userId, err)

		return
	}

	online = userOnline.IsOnline()

	return
}

// 给用户发送消息
func SendUserMessage(appId uint32, userId string, msgId, message string) (sendResults bool, err error) {
	// 封装发生数据格式
	data := models.GetTextMsgData(userId, msgId, message)
	// 获取与用户建立的socket client，如果不为空，则是当前机器，否则需要通过redis查找对应的服务，并通过rpc发生消息
	client := GetUserClient(appId, userId)

	if client != nil {
		// 在本机发送
		sendResults, err = SendUserMessageLocal(appId, userId, data)
		if err != nil {
			fmt.Println("给用户发送消息", appId, userId, err)
		}

		return
	}

	key := GetUserKey(appId, userId)
	info, err := cache.GetUserOnlineInfo(key)
	if err != nil {
		fmt.Println("给用户发送消息失败", key, err)

		return false, nil
	}
	if !info.IsOnline() {
		fmt.Println("用户不在线", key)
		return false, nil
	}
	server := models.NewServer(info.AccIp, info.AccPort)
	msg, err := grpcclient.SendMsg(server, msgId, appId, userId, models.MessageCmdMsg, models.MessageCmdMsg, message)
	if err != nil {
		fmt.Println("给用户发送消息失败", key, err)

		return false, err
	}
	fmt.Println("给用户发送消息成功-rpc", msg)
	sendResults = true

	return
}

// 给本机用户发送消息
func SendUserMessageLocal(appId uint32, userId string, data string) (sendResults bool, err error) {

	client := GetUserClient(appId, userId)
	if client == nil {
		err = errors.New("用户不在线")

		return
	}

	// 发送消息
	client.SendMsg([]byte(data))
	sendResults = true

	return
}

// 给全体用户发消息, 发生消息给所有的服务上的所有用户
func SendUserMessageAll(appId uint32, userId string, msgId, cmd, message string) (sendResults bool, err error) {
	sendResults = true

	currentTime := uint64(time.Now().Unix())
	servers, err := cache.GetServerAll(currentTime)
	if err != nil {
		fmt.Println("给全体用户发消息", err)

		return
	}

	for _, server := range servers {
		if IsLocal(server) {
			data := models.GetMsgData(userId, msgId, cmd, message)
			AllSendMessages(appId, userId, data)
		} else {
			grpcclient.SendMsgAll(server, msgId, appId, userId, cmd, message)
		}
	}

	return
}
