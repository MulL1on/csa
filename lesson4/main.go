package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()
	{
		r.POST("/register", register)
		r.POST("/login", login)
		r.PUT("/password", checkAnswer)
	}

	r.POST("/message", addMessage)
	{
		r.POST("/safeQuestion", addSafeQuestion)
		r.GET("/safeQuestion", getSafeQuestion)
	}

	r.Run()
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type SafeQuestion struct {
	Username string `json:"username"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type CheckAnswer struct {
	Username string `json:"username"`
	Answer   string `json:"answer"`
	Password string `json:"password"`
}

func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := ioutil.ReadFile("user.txt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var users []User

	if len(data) != 0 {
		//unmarshal users
		err = json.Unmarshal(data, &users)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	//encode password
	h := md5.New()
	h.Write([]byte(user.Password))
	user.Password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	//append user
	users = append(users, user)

	//marshal users
	data, err = json.Marshal(users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//write file
	err = ioutil.WriteFile("user.txt", data, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//encode password
	h := md5.New()
	h.Write([]byte(user.Password))
	user.Password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	//read file
	data, err := ioutil.ReadFile("user.txt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//unmarshal users
	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//check user
	for _, u := range users {
		if u.Username == user.Username && u.Password == user.Password {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, gin.H{"message": "username or password is wrong"})
}

func addMessage(c *gin.Context) {
	var message Message
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := os.OpenFile("message.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	//marsahal message
	data, err := json.Marshal(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = f.Write(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	f.Write([]byte("\n"))
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func addSafeQuestion(c *gin.Context) {
	var sq SafeQuestion
	if err := c.ShouldBindJSON(&sq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := os.OpenFile("safeQuestion.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	//marsahal sq
	data, err := json.Marshal(sq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = f.Write(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f.Write([]byte("\n"))
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getSafeQuestion(c *gin.Context) {
	username := c.Query("username")
	//read file
	data, err := ioutil.ReadFile("safeQuestion.txt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// scan file in line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var sq SafeQuestion
		err := json.Unmarshal(scanner.Bytes(), &sq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if sq.Username == username {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "question": sq.Question})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "username incorrect"})
			return
		}
	}

}

func checkAnswer(c *gin.Context) {
	var ca CheckAnswer
	if err := c.ShouldBindJSON(&ca); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//encode password
	h := md5.New()
	h.Write([]byte(ca.Password))
	ca.Password = base64.StdEncoding.EncodeToString(h.Sum(nil))

	//read  safe question  file
	data, err := ioutil.ReadFile("safeQuestion.txt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// scan file in line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var caSubject CheckAnswer
		err := json.Unmarshal(scanner.Bytes(), &caSubject)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ca.Username == caSubject.Username {
			if ca.Answer == caSubject.Answer {

				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				err = updatePassword(ca.Username, ca.Password)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
				return
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"status": "answer incorrect"})
				return
			}
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"status": "username incorrect"})

}

func updatePassword(username, password string) error {
	//read file
	data, err := ioutil.ReadFile("user.txt")
	if err != nil {
		return err
	}

	//unmarshal users
	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	//update user
	for i, u := range users {
		if u.Username == username {
			users[i].Password = password
			break
		}
	}

	//marshal users
	data, err = json.Marshal(users)
	if err != nil {
		return err
	}

	//write file
	err = ioutil.WriteFile("user.txt", data, 0644)
	if err != nil {
		return err
	}
	return nil
}
