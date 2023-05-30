package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB
var Rdb *redis.Client

type User struct {
	Id       int    `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"column:username;unique"`
	Password string `json:"password" gorm:"column:password"`
}

type Friend struct {
	Id         int    `json:"id" gorm:"primaryKey"`
	UserName   string `json:"username" gorm:"column:username"`
	FriendName string `json:"friend_name" gorm:"column:friend_name"`
}

type Post struct {
	Id       int    `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"column:username"`
	Content  string `json:"content" gorm:"column:content"`
}

func main() {
	r := gin.Default()
	initDB()
	initRedis()
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", register)
		userGroup.POST("/login", login)
		userGroup.GET("/", getUser)
	}

	friendGroup := r.Group("/friend")
	{
		friendGroup.POST("/", addFriend)
		friendGroup.GET("/list", getFriend)
		friendGroup.DELETE("/", deleteFriend)
	}

	postGroup := r.Group("/post")
	{
		postGroup.POST("/", addPost)
		postGroup.POST("/like", likePost)
	}

	r.Run()
}

func initRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	Rdb = rdb
}

func initDB() {
	dsn := "root:123456@tcp(localhost:3306)/csa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	//migration
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Friend{})
	db.AutoMigrate(&Post{})
	Db = db
}

func register(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//encode password
	h := md5.New()
	h.Write([]byte(u.Password))
	u.Password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	if err := Db.Create(&u).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})

}

func login(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//encode password
	h := md5.New()
	h.Write([]byte(u.Password))
	u.Password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	//get password from db
	var user User
	if err := Db.Where("username = ?", u.Username).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if user.Password != u.Password {
		c.JSON(400, gin.H{"error": "wrong password"})
		return
	}
	c.JSON(200, gin.H{"message": "success"})

}

func getUser(c *gin.Context) {
	var u User
	u.Username = c.Query("username")

	var user User
	res, err := Rdb.HGet(c, "user", user.Username).Result()
	if err != nil {
		if err != redis.Nil {
			c.JSON(400, gin.H{"error": "get user from cache error"})
			return
		}
	} else {
		user.Password = res
		c.JSON(200, gin.H{
			"message": "success",
			"user":    user,
		})
	}

	//get user from mysql
	if err = Db.Where("username = ?", u.Username).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//set user to cache
	if err = Rdb.HSet(c, "user", user.Username, user.Password).Err(); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"user":    user,
	})
}

func addFriend(c *gin.Context) {
	var f Friend
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("bind json error: %s", err.Error())})
		return
	}

	//check user exist
	var user User
	if err := Db.Where("username = ?", f.UserName).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, gin.H{"error": "user not exist"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//check friend exist
	var friend User
	if err := Db.Where("username = ?", f.FriendName).First(&friend).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, gin.H{"error": "friend not exist"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//add friend to db
	if err := Db.Create(&f).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func getFriend(c *gin.Context) {
	username := c.Query("username")
	//get friend from db
	var friends []Friend
	if err := Db.Where("username = ?", username).Find(&friends).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "success", "data": friends})

}

func deleteFriend(c *gin.Context) {
	var f Friend
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//check friendship exist
	var friend Friend
	if err := Db.Where("username = ? and friend_name = ?", f.UserName, f.FriendName).First(&friend).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, gin.H{"error": "friend not exist"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//delete friend from db
	if err := Db.Where("username = ? and friend_name = ?", f.UserName, f.FriendName).Delete(&f).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func addPost(c *gin.Context) {
	var p Post
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := Db.Create(&p).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"post":    p,
	})
}

type Like struct {
	Username string `json:"username"`
	PostId   int    `json:"post_id"`
}

func likePost(c *gin.Context) {
	var like Like
	if err := c.ShouldBindJSON(&like); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//check post exist
	var post Post
	if err := Db.Where("id = ?", like.PostId).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(400, gin.H{"error": "post not exist"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//like post in redis
	err := Rdb.SAdd(c, "post:"+"user:", like.Username).Err()

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})

}
