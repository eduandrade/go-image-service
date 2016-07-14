package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"os"
	"image/png"
	"image/color"
    "log"
    "net/http"
    "io"
    "regexp"

    //"github.com/eduandrade/imageutil"
    //"github.com/nfnt/resize"
    "github.com/disintegration/imaging"
)

func main() {
	log.Println("Starting app....")

	r := gin.Default()
	r.GET("/img/:id/:width/:height", resizeImage)
    r.POST("/img/upload", uploadImage)

    r.LoadHTMLGlob("templates/*")
    r.GET("/newlogo", func(c *gin.Context) {
        c.HTML(http.StatusOK, "newlogo.tmpl", nil)
    })
    
    port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func resizeImage(c *gin.Context) {
    id := c.Param("id")
	uwidth, err := strconv.Atoi(c.Param("width"))
	uheight, err := strconv.Atoi(c.Param("height"))
	var width int = uwidth
	var height int = uheight

    filename := fmt.Sprint("images/", id, "/original.png")
	file, err := os.Open(filename)
    if err != nil {
        log.Println("Error opening file", err)
        c.JSON(http.StatusBadRequest, gin.H{"status": "Error opening file " + filename})
        return
    }
    defer file.Close()

    img, err := png.Decode(file)
    if err != nil {
        log.Println("Error decoding file", err)
        c.JSON(http.StatusBadRequest, gin.H{"status": "Error decoding file " + filename})
        return
    }
    
   	m := imaging.Fit(img, width, height, imaging.Lanczos)
   	//m := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
    //m := imaging.Resize(img, width, height, imaging.Lanczos)
    //m := imaging.Thumbnail(img, width, height, imaging.Lanczos)

   	newImage := imaging.New(width, height, color.NRGBA{0, 255, 0, 255})
   	newImage = imaging.PasteCenter(newImage, m)

	newfilename := fmt.Sprint("images/", id, "/resized_", width, "_", height, ".png")
    out, err := os.Create(newfilename)
    if err != nil {
        log.Println("Error creating file", err)
        c.JSON(http.StatusBadRequest, gin.H{"status": "Error creating file " + newfilename})
        return
    }
    defer out.Close()

    png.Encode(out, newImage)

	c.File(newfilename)	

}

func uploadImage(c *gin.Context) {
    id := c.PostForm("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"status": "File id not informed"})
        return
    }

    re := regexp.MustCompile("^[a-z]*$")
    if !re.MatchString(id) {
        c.JSON(http.StatusBadRequest, gin.H{"status": "File id must have only lowercase letters"})
        return
    }


    file, header , err := c.Request.FormFile("uploadedFile")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "File not informed"})
        return
    }
    filename := header.Filename

    dirname := fmt.Sprint("images/", id)
    newfilename := fmt.Sprint(dirname, "/original.png")

    if _, err := os.Stat(dirname); os.IsNotExist(err) {
        os.Mkdir(dirname, 0700)
    }

    out, err := os.Create(newfilename)
    if err != nil {
        log.Println("Error creating file", err)
        c.JSON(http.StatusBadRequest, gin.H{"status": "Error creating file " + filename})
        return
    }
    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
        log.Println("Error saving file", err)
        c.JSON(http.StatusBadRequest, gin.H{"status": "Error saving file " + filename})
        return
    }  

    c.HTML(http.StatusOK, "newlogo.tmpl", gin.H{"status": "File uploaded!",})
}

