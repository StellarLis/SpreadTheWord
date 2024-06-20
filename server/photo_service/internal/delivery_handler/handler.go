package deliveryhandler

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"photo_service/internal/repository"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type CreatePhotoMessage struct {
	UserId   int
	Username string
}

func Handle(d amqp.Delivery, repository *repository.Repository) {
	message := CreatePhotoMessage{}
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		logrus.Fatalln(err)
	}

	size := 200
	avatar, err := createAvatar(size, message.Username[:2])
	if err != nil {
		logrus.Fatalln(err)
	}
	buffer := &bytes.Buffer{}
	png.Encode(buffer, avatar)
	repository.UpdateDbAvatar(buffer.Bytes(), message.UserId)
}

func createAvatar(size int, initials string) (*image.RGBA, error) {
	width, height := size, size
	colors := []string{"#fe6f47", "#6b82e2", "#b0b0b0", "#f08080", "#ff80be"}
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		logrus.Fatalln(err)
	}
	bgColor, err := ParseHexColor(colors[randomNum.Int64()])
	if err != nil {
		logrus.Fatalln(err)
	}
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(background, background.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
	err = drawText(background, initials)
	if err != nil {
		logrus.Fatalln(err)
	}
	return background, err
}

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	return
}

func drawText(canvas *image.RGBA, text string) error {
	var (
		fgColor  image.Image
		fontFace *truetype.Font
		err      error
		fontSize = 128.0
	)
	fgColor = image.White
	fontFace, err = freetype.ParseFont(goregular.TTF)
	fontDrawer := &font.Drawer{
		Dst: canvas,
		Src: fgColor,
		Face: truetype.NewFace(fontFace, &truetype.Options{
			Size:    fontSize,
			Hinting: font.HintingFull,
		}),
	}
	textBounds, _ := fontDrawer.BoundString(text)
	xPosition := (fixed.I(canvas.Rect.Max.X) - fontDrawer.MeasureString(text)) / 2
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	yPosition := fixed.I((canvas.Rect.Max.Y)-textHeight.Ceil())/2 + fixed.I(textHeight.Ceil())
	fontDrawer.Dot = fixed.Point26_6{
		X: xPosition,
		Y: yPosition,
	}
	fontDrawer.DrawString(text)
	return err
}
