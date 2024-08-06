package main

import (
    "github.com/gin-gonic/gin"
    "github.com/plivo/plivo-go"
    "github.com/joho/godotenv"
    "log"
    "net/http"
    "os"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    plivoAuthID := os.Getenv("PLIVO_AUTH_ID")
    plivoAuthToken := os.Getenv("PLIVO_AUTH_TOKEN")

    router := gin.Default()

    router.POST("/call", func(c *gin.Context) {
        var json struct {
            From  string `json:"from" binding:"required"`
            To    string `json:"to" binding:"required"`
            Title string `json:"title" binding:"required"`
            Name  string `json:"name" binding:"required"`
        }

        if err := c.ShouldBindJSON(&json); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        otp := generateRandomOTP()

        client, err := plivo.NewClient(plivoAuthID, plivoAuthToken, &plivo.ClientOptions{})
        if err != nil {
            log.Fatalf("Error creating Plivo client: %v", err)
        }

        _, err = client.Calls.Create(plivo.CallCreateParams{
            From: json.From,
            To:   json.To,
            AnswerURL: "https://your-ngrok-url/call-handler?from=" + json.From + "&to=" + json.To + "&title=" + json.Title + "&name=" + json.Name + "&otp=" + otp,
            AnswerMethod: "GET",
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"status": "Call initiated"})
    })

    router.GET("/call-handler", func(c *gin.Context) {
        from := c.Query("from")
        to := c.Query("to")
        title := c.Query("title")
        name := c.Query("name")
        otp := c.Query("otp")

        response := "<Response><Say>Hello " + name + ". This is the " + title + ". We have sent this automated call because of an attempt related to your account. If this is not you, please press 1.</Say>"
        response += "<Gather action='/handle-keypress' method='POST' numDigits='1'><Say>To verify your identity, please enter the 6-digit security code that we have sent to your mobile device: " + otp + "</Say></Gather>"
        response += "</Response>"

        c.Header("Content-Type", "application/xml")
        c.String(http.StatusOK, response)
    })

    router.POST("/handle-keypress", func(c *gin.Context) {
        digits := c.PostForm("Digits")
        response := "<Response>"

        if digits == "" {
            response += "<Say>Invalid input. Please try again.</Say><Redirect method='GET'>/call-handler</Redirect>"
        } else {
            response += "<Say>Please wait as we verify the code.</Say><Hangup/>"
        }

        response += "</Response>"

        c.Header("Content-Type", "application/xml")
        c.String(http.StatusOK, response)
    })

    router.Run(":8080")
}

func generateRandomOTP() string {
    otp := ""
    for i := 0; i < 6; i++ {
        otp += string('0' + rune(rand.Intn(10)))
    }
    return otp
}
