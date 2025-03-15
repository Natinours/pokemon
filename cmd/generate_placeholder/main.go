package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func main() {
	// Créer une nouvelle image de 400x560 pixels (format carte Pokémon)
	width := 400
	height := 560
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Remplir l'arrière-plan avec une couleur grise foncée
	bgColor := color.RGBA{42, 42, 42, 255} // #2a2a2a
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Créer un cadre plus clair
	borderColor := color.RGBA{64, 64, 64, 255}
	borderWidth := 2
	// Bordure supérieure
	draw.Draw(img, image.Rect(0, 0, width, borderWidth), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// Bordure inférieure
	draw.Draw(img, image.Rect(0, height-borderWidth, width, height), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// Bordure gauche
	draw.Draw(img, image.Rect(0, 0, borderWidth, height), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	// Bordure droite
	draw.Draw(img, image.Rect(width-borderWidth, 0, width, height), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	// Sauvegarder l'image
	f, err := os.Create("static/img/series/card-placeholder.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
