package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

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

func main() {
	r := gin.Default()
	initDB()
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", register)
		userGroup.POST("/login", login)
	}

	friendGroup := r.Group("/friend")
	{
		friendGroup.POST("/", addFriend)
		friendGroup.GET("/list", getFriend)
		friendGroup.DELETE("/", deleteFriend)
	}

	r.Run()
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
