package main

import (
	"database/sql"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateShortURL() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type URL struct {
	ShortURL    string
	OriginalURL string
}

func main() {
	db, err := sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        short_url TEXT NOT NULL,
        original_url TEXT NOT NULL
    );`)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.LoadHTMLFiles("templates/index.html")

	shortly := router.Group("/shortly")
	{
		shortly.GET("/", func(c *gin.Context) {
			rows, err := db.Query("SELECT short_url, original_url FROM urls")
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			defer rows.Close()

			var urls []URL
			for rows.Next() {
				var url URL
				if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				urls = append(urls, url)
			}

			c.HTML(http.StatusOK, "index.html", gin.H{
				"URLs": urls,
			})
		})

		shortly.POST("/shorten", func(c *gin.Context) {
			originalURL := c.PostForm("url")
			if !strings.HasPrefix(originalURL, "https://") {
				originalURL = "https://" + originalURL
			}
			shortURL := generateShortURL()

			_, err := db.Exec("INSERT INTO urls (short_url, original_url) VALUES (?, ?)", shortURL, originalURL)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			shortURLFull := "http://" + c.Request.Host + "/" + "shortly/" + shortURL

			response := struct {
				ShortURL string
			}{
				ShortURL: shortURLFull,
			}

			t, err := template.New("result").Parse(`<p>Shortened URL: <a href="{{.ShortURL}}" target="_blank">{{.ShortURL}}</a></p>`)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.Header("Content-Type", "text/html")
			t.Execute(c.Writer, response)
		})

		shortly.GET("/:shortURL", func(c *gin.Context) {
			shortURL := c.Param("shortURL")

			var originalURL string
			err := db.QueryRow("SELECT original_url FROM urls WHERE short_url = ?", shortURL).Scan(&originalURL)
			if err == sql.ErrNoRows {
				c.String(http.StatusNotFound, "URL not found")
			} else if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.Redirect(http.StatusFound, originalURL)
			}
		})

		shortly.DELETE("/delete/:shortURL", func(c *gin.Context) {
			shortURL := c.Param("shortURL")

			_, err := db.Exec("DELETE FROM urls WHERE short_url = ?", shortURL)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.Status(http.StatusOK)
		})
	}

	log.Println("Server started at :8080")
	router.Run(":8080")
}
