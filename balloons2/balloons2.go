package main										//all thanks to :
													//Jack Mott - https://www.twitch.tv/jackmott42
													//Stefan Gustavson, 2003-2005 Contact: stefan.gustavson@liu.se

import (
	"os"
	"fmt"
	"time"
	"github.com/yawhoo/noise"
	"image/png"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600			//using this constant for future window

type balloon struct {
	tex *sdl.Texture
	pos
	scale float32
	w,h int
}

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	newW := int32(float32(balloon.w) * balloon.scale)
	newH := int32(float32(balloon.h) * balloon.scale)
	x := int32(balloon.x - float32(newW) / 2)
	y := int32(balloon.y - float32(newH) / 2)
	rect := &sdl.Rect{x,y, newW, newH}
	renderer.Copy(balloon.tex, nil, rect)
}

type rgba struct{
	r, g, b byte
}

type pos struct {
	x,y float32
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth +x)*4
	if index < len(pixels)-4 && index >=0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b 
	}
}

func pixelsToTextures(renderer *sdl.Renderer, pixels []byte, w, h int)*sdl.Texture{
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func loadBalloons(renderer *sdl.Renderer)[]balloon{

	balloonStrs := []string{"balloons/balloon_red.png", "balloons/balloon_green.png", "balloons/balloon_blue.png"}
	balloons:= make([]balloon, len(balloonStrs))

	for i, bstr := range balloonStrs {	
	infile, err := os.Open(bstr)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	balloonPixels := make([]byte, w*h*4)			//all ballon pixels would be extracted here
	bIndex := 0
	for y := 0; y < h; y++ {						
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x,y).RGBA()		//extract rgba colors from img (16byte val because have alfa)
			balloonPixels[bIndex] = byte(r/256)		
			bIndex++
			balloonPixels[bIndex] = byte(g/256)
			bIndex++
			balloonPixels[bIndex] = byte(b/256)
			bIndex++
			balloonPixels[bIndex] = byte(a/256)
			bIndex++
		}	
	}
	tex := pixelsToTextures(renderer, balloonPixels, w, h)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	balloons[i] = balloon{tex, pos{float32(i*120),float32(i*120)}, float32(1+i)/2, w, h}
	}
	return balloons
}

func lerp(b1, b2 byte, pct float32)byte {						//gives us linear interpolation between 2 bytes
	return byte(float32(b1) + pct *(float32(b2)- float32(b1))) 	//pct - percent
}

func colorLerp(c1, c2 rgba, pct float32) rgba {				//linear interpolation of each component (red,green,blue)
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}

}

func getGradient(c1,c2 rgba) []rgba{						//making gradient from 2 colors
	result:= make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 rgba) []rgba{					//second gradient
	result:= make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1,c2, pct * float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct * float32(1.5)- float32(0.5))
		}
		
	}
	return result
}

func clamp(min, max, v int) int {							//this func make sure values in between range of what we looking for
	if v < min {
		v = min
	}else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32,min, max float32, gradient []rgba, w, h int) []byte {		//rescale noise --> turn in to byte --> set pixel array with that byte
	result := make([]byte, w * h * 4)
	scale := 255.0 / (max-min)								//to expand our values  from 0 to 255 using min-max
	offset := min * scale

	for i := range noise {									//going thru slice of noise
		noise[i] = noise[i] * scale - offset				//modifying and putting back
		c := gradient[clamp(0,255,int(noise[i]))]			//passing min/max and 
		p := i * 4
		result[p] = c.r									//set those byte in to pixels slice
		result[p + 1] = c.g
		result[p+ 2] = c.b
	}
	return result
}

func main(){

	err := sdl.Init(sdl.INIT_EVERYTHING)		//initializing sdl, wont work without it
	if err != nil {
		fmt.Println(err)
		return
	}

defer sdl.Quit() 								//closing properly after everything is done

window, err := sdl.CreateWindow("My name is what? my name is who? my name is window", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,		// creating window with sdl and our const for window
int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
if err != nil {
	fmt.Println(err)
	return
}
defer window.Destroy()							//all "defers" initializing from bottom to top

renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)  	// creating renderer
if err != nil {
	fmt.Println(err)							//always checking for mistakes
	return
}
defer renderer.Destroy()						//and cleaning up

sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))//using renderer to create textures
if err != nil {
	fmt.Println(err)
	return
}
defer tex.Destroy()

cloudNoise, min, max := noise.MakeNoise(noise.FBM, 0.009, 0.4, 3, 3, winWidth, winHeight)			//noise for background
cloudGradient := getGradient(rgba{240,0,0}, rgba{250,250,250})										//grad for bg
cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight) 
cloudTexture := pixelsToTextures(renderer, cloudPixels, winWidth, winHeight)

balloons := loadBalloons(renderer)						//balloon textures
dir := 1												//dirrection for balloon

for {
	frameStart := time.Now()
	for event := sdl.PollEvent(); event !=nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		}
	}
	renderer.Copy(cloudTexture, nil, nil)

	for _, balloon := range balloons {
		balloon.draw(renderer)
	}
	balloons[1].x += float32(1*dir)
	if balloons[1].x > 400 || balloons[1].x < 0 {
		dir = dir * -1
	}
	
	renderer.Present()
	elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
	fmt.Println("ms per frame: ", elapsedTime)
	if elapsedTime < 5 {
		sdl.Delay(5 - uint32(elapsedTime))

	}
	sdl.Delay(16)
}
}