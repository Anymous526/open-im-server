package main

import (
	_ "Open_IM/cmd/open_im_api/docs"
	apiAuth "Open_IM/internal/api/auth"
	"Open_IM/internal/api/conversation"
	"Open_IM/internal/api/friend"
	"Open_IM/internal/api/group"
	"Open_IM/internal/api/manage"
	apiChat "Open_IM/internal/api/msg"
	apiThird "Open_IM/internal/api/third"
	"Open_IM/internal/api/user"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/middleware"
	"Open_IM/pkg/common/tokenverify"
	"Open_IM/pkg/utils"
	"flag"
	"fmt"

	"io"
	"os"
	"strconv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	//"syscall"
	"Open_IM/pkg/common/constant"
	promePkg "Open_IM/pkg/common/prometheus"
)

// @title open-IM-Server API
// @version 1.0
// @description  open-IM-Server 的API服务器文档, 文档中所有请求都有一个operationID字段用于链路追踪

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
func main() {
	log.NewPrivateLog(constant.LogFileName)
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create("../logs/api.log")
	gin.DefaultWriter = io.MultiWriter(f)
	//	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.GinParseOperationID)
	log.Info("load config: ", config.Config)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if config.Config.Prometheus.Enable {
		promePkg.NewApiRequestCounter()
		promePkg.NewApiRequestFailedCounter()
		promePkg.NewApiRequestSuccessCounter()
		r.Use(promePkg.PromeTheusMiddleware)
		r.GET("/metrics", promePkg.PrometheusHandler())
	}
	// user routing group, which handles user registration and login services
	userRouterGroup := r.Group("/user")
	{
		userRouterGroup.POST("/update_user_info", user.UpdateUserInfo) //1
		userRouterGroup.POST("/set_global_msg_recv_opt", user.SetGlobalRecvMessageOpt)
		userRouterGroup.POST("/get_users_info", user.GetUsersPublicInfo)            //1
		userRouterGroup.POST("/get_self_user_info", user.GetSelfUserInfo)           //1
		userRouterGroup.POST("/get_users_online_status", user.GetUsersOnlineStatus) //1
		userRouterGroup.POST("/get_users_info_from_cache", user.GetUsersInfoFromCache)
		userRouterGroup.POST("/get_user_friend_from_cache", user.GetFriendIDListFromCache)
		userRouterGroup.POST("/get_black_list_from_cache", user.GetBlackIDListFromCache)
		userRouterGroup.POST("/get_all_users_uid", manage.GetAllUsersUid) //1
		userRouterGroup.POST("/account_check", manage.AccountCheck)       //1
		//	userRouterGroup.POST("/get_users_online_status", manage.GetUsersOnlineStatus) //1
		userRouterGroup.POST("/get_users", user.GetUsers)
	}
	////friend routing group
	friendRouterGroup := r.Group("/friend")
	{
		//	friendRouterGroup.POST("/get_friends_info", friend.GetFriendsInfo)
		friendRouterGroup.POST("/add_friend", friend.AddFriend)                        //1
		friendRouterGroup.POST("/delete_friend", friend.DeleteFriend)                  //1
		friendRouterGroup.POST("/get_friend_apply_list", friend.GetFriendApplyList)    //1
		friendRouterGroup.POST("/get_self_friend_apply_list", friend.GetSelfApplyList) //1
		friendRouterGroup.POST("/get_friend_list", friend.GetFriendList)               //1
		friendRouterGroup.POST("/add_friend_response", friend.AddFriendResponse)       //1
		friendRouterGroup.POST("/set_friend_remark", friend.SetFriendRemark)           //1

		friendRouterGroup.POST("/add_black", friend.AddBlack)           //1
		friendRouterGroup.POST("/get_black_list", friend.GetBlacklist)  //1
		friendRouterGroup.POST("/remove_black", friend.RemoveBlacklist) //1

		friendRouterGroup.POST("/import_friend", friend.ImportFriend) //1
		friendRouterGroup.POST("/is_friend", friend.IsFriend)         //1
	}
	//group related routing group
	groupRouterGroup := r.Group("/group")
	groupRouterGroup.Use(func(c *gin.Context) {
		userID, err := tokenverify.ParseUserIDFromToken(c.GetHeader("token"), c.GetString("operationID"))
		if err != nil {
			c.String(400, err.Error())
			c.Abort()
			return
		}
		c.Set("opUserID", userID)
		c.Next()
	})
	{
		groupRouterGroup.POST("/create_group", group.NewCreateGroup)                                //1
		groupRouterGroup.POST("/set_group_info", group.NewSetGroupInfo)                             //1
		groupRouterGroup.POST("/join_group", group.JoinGroup)                                       //1
		groupRouterGroup.POST("/quit_group", group.QuitGroup)                                       //1
		groupRouterGroup.POST("/group_application_response", group.ApplicationGroupResponse)        //1
		groupRouterGroup.POST("/transfer_group", group.TransferGroupOwner)                          //1
		groupRouterGroup.POST("/get_recv_group_applicationList", group.GetRecvGroupApplicationList) //1
		groupRouterGroup.POST("/get_user_req_group_applicationList", group.GetUserReqGroupApplicationList)
		groupRouterGroup.POST("/get_groups_info", group.GetGroupsInfo) //1
		groupRouterGroup.POST("/kick_group", group.KickGroupMember)    //1
		//	groupRouterGroup.POST("/get_group_member_list", group.FindGroupMemberAll)        //no use
		groupRouterGroup.POST("/get_group_all_member_list", group.GetGroupAllMemberList) //1
		groupRouterGroup.POST("/get_group_members_info", group.GetGroupMembersInfo)      //1
		groupRouterGroup.POST("/invite_user_to_group", group.InviteUserToGroup)          //1
		groupRouterGroup.POST("/get_joined_group_list", group.GetJoinedGroupList)
		groupRouterGroup.POST("/dismiss_group", group.DismissGroup) //
		groupRouterGroup.POST("/mute_group_member", group.MuteGroupMember)
		groupRouterGroup.POST("/cancel_mute_group_member", group.CancelMuteGroupMember) //MuteGroup
		groupRouterGroup.POST("/mute_group", group.MuteGroup)
		groupRouterGroup.POST("/cancel_mute_group", group.CancelMuteGroup)
		groupRouterGroup.POST("/set_group_member_nickname", group.SetGroupMemberNickname)
		groupRouterGroup.POST("/set_group_member_info", group.SetGroupMemberInfo)
		groupRouterGroup.POST("/get_group_abstract_info", group.GetGroupAbstractInfo)
		//groupRouterGroup.POST("/get_group_all_member_list_by_split", group.GetGroupAllMemberListBySplit)
	}
	superGroupRouterGroup := r.Group("/super_group")
	{
		superGroupRouterGroup.POST("/get_joined_group_list", group.GetJoinedSuperGroupList)
		superGroupRouterGroup.POST("/get_groups_info", group.GetSuperGroupsInfo)
	}
	////certificate
	authRouterGroup := r.Group("/auth")
	{
		authRouterGroup.POST("/user_register", apiAuth.UserRegister) //1
		authRouterGroup.POST("/user_token", apiAuth.UserToken)       //1
		authRouterGroup.POST("/parse_token", apiAuth.ParseToken)     //1
		authRouterGroup.POST("/force_logout", apiAuth.ForceLogout)   //1
	}
	////Third service
	thirdGroup := r.Group("/third")
	{
		thirdGroup.POST("/tencent_cloud_storage_credential", apiThird.TencentCloudStorageCredential)
		thirdGroup.POST("/ali_oss_credential", apiThird.AliOSSCredential)
		thirdGroup.POST("/minio_storage_credential", apiThird.MinioStorageCredential)
		thirdGroup.POST("/minio_upload", apiThird.MinioUploadFile)
		thirdGroup.POST("/upload_update_app", apiThird.UploadUpdateApp)
		thirdGroup.POST("/get_download_url", apiThird.GetDownloadURL)
		thirdGroup.POST("/get_rtc_invitation_info", apiThird.GetRTCInvitationInfo)
		thirdGroup.POST("/get_rtc_invitation_start_app", apiThird.GetRTCInvitationInfoStartApp)
		thirdGroup.POST("/fcm_update_token", apiThird.FcmUpdateToken)
		thirdGroup.POST("/aws_storage_credential", apiThird.AwsStorageCredential)
		thirdGroup.POST("/set_app_badge", apiThird.SetAppBadge)
	}
	////Message
	chatGroup := r.Group("/msg")
	{
		chatGroup.POST("/newest_seq", apiChat.GetSeq)
		chatGroup.POST("/send_msg", apiChat.SendMsg)
		chatGroup.POST("/pull_msg_by_seq", apiChat.PullMsgBySeqList)
		chatGroup.POST("/del_msg", apiChat.DelMsg)
		chatGroup.POST("/del_super_group_msg", apiChat.DelSuperGroupMsg)
		chatGroup.POST("/clear_msg", apiChat.ClearMsg)
		chatGroup.POST("/manage_send_msg", manage.ManagementSendMsg)
		chatGroup.POST("/batch_send_msg", manage.ManagementBatchSendMsg)
		chatGroup.POST("/check_msg_is_send_success", manage.CheckMsgIsSendSuccess)
		chatGroup.POST("/set_msg_min_seq", apiChat.SetMsgMinSeq)

		chatGroup.POST("/set_message_reaction_extensions", apiChat.SetMessageReactionExtensions)
		chatGroup.POST("/get_message_list_reaction_extensions", apiChat.GetMessageListReactionExtensions)
		chatGroup.POST("/add_message_reaction_extensions", apiChat.AddMessageReactionExtensions)
		chatGroup.POST("/delete_message_reaction_extensions", apiChat.DeleteMessageReactionExtensions)
	}
	////Conversation
	conversationGroup := r.Group("/conversation")
	{ //1
		conversationGroup.POST("/get_all_conversations", conversation.GetAllConversations)
		conversationGroup.POST("/get_conversation", conversation.GetConversation)
		conversationGroup.POST("/get_conversations", conversation.GetConversations)
		conversationGroup.POST("/set_conversation", conversation.SetConversation)
		conversationGroup.POST("/batch_set_conversation", conversation.BatchSetConversations)
		conversationGroup.POST("/set_recv_msg_opt", conversation.SetRecvMsgOpt)
		conversationGroup.POST("/modify_conversation_field", conversation.ModifyConversationField)
	}

	go apiThird.MinioInit()
	defaultPorts := config.Config.Api.GinPort
	ginPort := flag.Int("port", defaultPorts[0], "get ginServerPort from cmd,default 10002 as port")
	flag.Parse()
	address := "0.0.0.0:" + strconv.Itoa(*ginPort)
	if config.Config.Api.ListenIP != "" {
		address = config.Config.Api.ListenIP + ":" + strconv.Itoa(*ginPort)
	}
	fmt.Println("start api server, address: ", address, ", OpenIM version: ", constant.CurrentVersion)
	err := r.Run(address)
	if err != nil {
		log.Error("", "api run failed ", address, err.Error())
		panic("api start failed " + err.Error())
	}
}
