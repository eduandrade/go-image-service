package main

import (
	"fmt"
	"strconv"
	"os"
	"image/png"
	"image/color"
    "log"
    "net/http"
    "io"
    "regexp"
    "image"
    "github.com/gin-gonic/gin"
    "github.com/disintegration/imaging"
)

func main() {
	log.Println("Starting app....")

	r := gin.Default()
    r.GET("/img/:id/:width/:height", resizeImageWithoutBg)
	r.GET("/img/:id/:width/:height/:bgcolor", resizeImageWithBg)
    r.POST("/img/upload", uploadImage)

    r.LoadHTMLGlob("templates/*")
    r.GET("/newlogo", func(c *gin.Context) {
        c.HTML(http.StatusOK, "newlogo.tmpl", nil)
    })
    
    port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	r.Run(":" + port)
}

func resizeImageWithoutBg(c *gin.Context) {
    resizeImage("", c)
}

func resizeImageWithBg(c *gin.Context) {
    resizeImage(c.Param("bgcolor"), c)
}

func resizeImage(bgcolor string, c *gin.Context) {
    id := c.Param("id")
	uwidth, err := strconv.Atoi(c.Param("width"))
	if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "Invalid width"})
        return
    }

    uheight, err := strconv.Atoi(c.Param("height"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "Invalid height"})
        return
    }

	var width int = uwidth
	var height int = uheight

    filename := fmt.Sprint("images/", id, "/original.png")
	file, err := os.Open(filename)
    if err != nil {
        log.Println("Error opening file", err)
        c.JSON(http.StatusNotFound, gin.H{"status": "Error opening file " + filename})
        return
    }
    defer file.Close()

    img, err := png.Decode(file)
    if err != nil {
        log.Println("Error decoding file", err)
        c.JSON(http.StatusNotFound, gin.H{"status": "Error decoding file " + filename})
        return
    }
    
   	m := imaging.Fit(img, width, height, imaging.Lanczos)
   	//m := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)
    //m := imaging.Resize(img, width, height, imaging.Lanczos)
    //m := imaging.Thumbnail(img, width, height, imaging.Lanczos)
   	
    var newImage *image.NRGBA
    if bgcolor == "" {
        newImage = imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
    } else {
        alpha := 0xff //fully opaque
        col, err := hex(bgcolor)
        if err != nil{
            c.JSON(http.StatusNotFound, gin.H{"status": "Invalid color"})
            return
        }
        newImage = imaging.New(width, height, color.NRGBA{col.R, col.G, col.B, uint8(alpha)})
    }   

    newImage = imaging.OverlayCenter(newImage, m, 1.0)

	newfilename := fmt.Sprint("images/", id, "/resized_", width, "_", height, ".png")
    out, err := os.Create(newfilename)
    if err != nil {
        log.Println("Error creating file", err)
        c.JSON(http.StatusNotFound, gin.H{"status": "Error creating file " + newfilename})
        return
    }
    defer out.Close()

    png.Encode(out, newImage)

    c.File(newfilename)	

}

func uploadImage(c *gin.Context) {
    id := c.PostForm("id")
    if id == "" {
        c.JSON(http.StatusNotFound, gin.H{"status": "File id not informed"})
        return
    }

    re := regexp.MustCompile("^[a-z]*$")
    if !re.MatchString(id) {
        c.JSON(http.StatusNotFound, gin.H{"status": "File id must have only lowercase letters"})
        return
    }


    file, header , err := c.Request.FormFile("uploadedFile")
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"status": "File not informed"})
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
        c.JSON(http.StatusNotFound, gin.H{"status": "Error creating file " + filename})
        return
    }
    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
        log.Println("Error saving file", err)
        c.JSON(http.StatusNotFound, gin.H{"status": "Error saving file " + filename})
        return
    }  

    c.HTML(http.StatusOK, "newlogo.tmpl", gin.H{"status": "File uploaded!",})
}

// A color is stored internally using sRGB (standard RGB) values in the range 0-1
type CustomColor struct {
    R, G, B uint8
}

// Hex parses a "html" hex color-string, either in the 3 "#f0c" or 6 "#ff1034" digits form.
func hex(scol string) (CustomColor, error) {
    format := "%02x%02x%02x"
    //factor := 1/255
    if len(scol) == 4 {
        format = "#%1x%1x%1x"
        //factor = 1/15
    }

    var r, g, b uint8
    n, err := fmt.Sscanf(scol, format, &r, &g, &b)
    if err != nil {
        return CustomColor{}, err
    }
    if n != 3 {
        return CustomColor{}, fmt.Errorf("color: %v is not a hex-color", scol)
    }

    //fmt.Printf(">RGB values: %v, %v, %v", r, g, b)
    //return CustomColor{r*uint8(factor), g*uint8(factor), b*uint8(factor)}, nil
    return CustomColor{r, g, b}, nil
}

