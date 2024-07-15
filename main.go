package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangchunlong818/go-email-shunt/sender"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()
	r.POST("/sendEmail", sendEmail)
	_ = r.Run(":8000")
}

func sendEmail(c *gin.Context) {
	type baseResponse struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
		Msg  string      `json:"msg"`
	}
	content := &sender.MailContent{
		EmailType: 1,
		From: sender.EmailAddress{
			Email: "returns@parcelpanel.org",
			Name:  "ParcelPanel Returns2",
		},
		To: sender.EmailAddress{
			Email: "a1257407454@gmail.com",
			Name:  "caoyanyan",
		},
		ReplyTo: sender.EmailAddress{
			Email: "Returns@channelwill.com",
			Name:  "ParcelPanel Returns2",
		},
		Subject:        "Important Updates and Limited Time Offer for ParcelPanel Returns Customers!",
		Html:           "<html><h2>This is a test warming email</h2></html>",
		UnsubscribeUrl: "https://www.baidu.com",
		Tags:           map[string]string{},
	}
	redisCli := getRedis()
	db := getDb()
	emailSenderSrv := sender.NewSenderManager(redisCli, db, 2)
	send, err := emailSenderSrv.Send(content)
	if err != nil {
		c.JSON(http.StatusForbidden, baseResponse{
			http.StatusForbidden,
			send,
			err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, baseResponse{
		http.StatusOK,
		send,
		"success",
	})
	return
}

// 初始化数据库连接
func getDb() *gorm.DB {
	dsn := fmt.Sprintf("root:root@tcp(192.168.15.194:3306)/returns_db?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}))
	if err != nil {
		panic("数据库连接失败,报错信息" + err.Error())
	}

	sqlDB, _ := db.DB()
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(20)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。maxLifeTime必须要比mysql服务器设置的wait_timeout小，否则会导致golang侧连接池依然保留已被mysql服务器关闭了的连接。
	//mysql服务器的wait_timeout默认是8 hour，可通过show variables like 'wait_timeout’查看。
	sqlDB.SetConnMaxLifetime(time.Hour * 24)

	//ping 检测
	if err := sqlDB.Ping(); err != nil {
		panic("数据库连接失败ping检测失败信息: " + err.Error())
	}

	// 自动迁移
	//db.AutoMigrate(&UserInfo{})
	return db
}

func getRedis() *redis.Client {
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", 6379)

	redisOptions := &redis.Options{
		Network:  "tcp", //网络类型，tcp or unix，默认tcp
		Addr:     addr,  // Redis地址
		Password: "",    // Redis账号
		DB:       8,     // Redis库

		//连接池容量及闲置连接数量
		PoolSize:     10,  // 连接池最大socket连接数
		MinIdleConns: 100, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量
	}
	//如果有用户名
	//redisOptions.Username = ""
	//更多配置查看 https://blog.csdn.net/pengpengzhou/article/details/105385666
	redisCli := redis.NewClient(redisOptions)

	_, err := redisCli.Ping(context.Background()).Result()
	if err == redis.Nil {
		panic("Redis异常:" + err.Error())
	} else if err != nil {
		panic("Redis失败:" + err.Error())
	}
	return redisCli
}
