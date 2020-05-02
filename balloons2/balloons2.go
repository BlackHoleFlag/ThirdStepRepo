package main										//all thanks to :
													//Jack Mott - https://www.twitch.tv/jackmott42
													//Stefan Gustavson, 2003-2005 Contact: stefan.gustavson@liu.se

import (
	"os"
	"fmt"
	"time"
	"sort"
	"math/rand"
	"image/png"
	"github.com/yawhoo/noise"
	."github.com/yawhoo/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight, winDepth int = 800, 600, 100			//using this constant for future window

type mouseState struct {
	leftButton bool
	rightButton bool
	x,y int
}

func getMouseState() mouseState{
	
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY) 
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)
	return result
}

type balloon struct {
	tex *sdl.Texture
	pos Vector3
	dir Vector3
	w,h int
}

type balloonArray []*balloon

func (balloons balloonArray) Len() int {
	return len(balloons)
}

func (balloons balloonArray) Swap(i, j int) {
	balloons[i], balloons[j] = balloons[j], balloons[i]
}

func (balloons balloonArray) Less(i,j int)bool{
	diff := balloons[i].pos.Z - balloons[j].pos.Z
	return diff < 1
}

func (balloon * balloon) update(elapsedTime float32) {
	p := Add(balloon.pos, Mult(balloon.dir, elapsedTime))

	if p.X < 0 || p.X > float32(winWidth) {
		balloon.dir.X = -balloon.dir.X
	}
	if p.Y < 0 || p.Y > float32(winHeight) {
		balloon.dir.Y = -balloon.dir.Y
	}
	if p.Z < 0 || p.Z > float32(winDepth) {
		balloon.dir.Z = -balloon.dir.Z
	}

	balloon.pos = Add(balloon.pos, Mult(balloon.dir, elapsedTime))
}	

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	scale := (balloon.pos.Z / 200 + 1)/2
	newW := int32(float32(balloon.w) * scale)
	newH := int32(float32(balloon.h) * scale)
	x := int32(balloon.pos.X - float32(newW) / 2)
	y := int32(balloon.pos.Y - float32(newH) / 2)
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

func loadBalloons(renderer *sdl.Renderer, numBalloons int)[]*balloon{

	balloonStrs := []string{"balloons/balloon_red.png", "balloons/balloon_green.png", "balloons/balloon_blue.png"}
	balloonTextures:= make([]*sdl.Texture, len(balloonStrs))

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
	balloonTextures[i] = tex
	}
	balloons := make([]*balloon, numBalloons)
	for i := range balloons {
		tex := balloonTextures[i % 3]			//to rotate between pictures
		pos := Vector3{rand.Float32() * float32(winWidth), rand.Float32() * float32(winHeight), rand.Float32() * float32(winDepth)}
		dir := Vector3 {rand.Float32(), rand.Float32(), rand.Float32()}
		_, _, w, h, err := tex.Query()
		if err != nil{
			panic(err)
		}
		balloons[i] = &balloon{tex, pos, dir, int(w), int(h)}
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

cloudNoise, min, max := noise.MakeNoise(noise.FBM, 0.009, 0.4, 3, 3, winWidth, winHeight)			//noise for background
cloudGradient := getGradient(rgba{240,0,0}, rgba{250,250,250})										//grad for bg
cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight) 
cloudTexture := pixelsToTextures(renderer, cloudPixels, winWidth, winHeight)

balloons := loadBalloons(renderer, 20)						//balloon textures
var elapsedTime float32

currentMouseState := getMouseState()
prevMouseState := currentMouseState

for {
	frameStart := time.Now()
	for event := sdl.PollEvent(); event !=nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		}
	}

	currentMouseState := getMouseState()

	if !currentMouseState.leftButton && prevMouseState.leftButton {
		fmt.Println("left click")
	}

	renderer.Copy(cloudTexture, nil, nil)

	for _, balloon := range balloons {
		balloon.update(elapsedTime)
	}

	sort.Stable(balloonArray(balloons))				//to sort bigger in fron(close), smaller on back(far)

	for _, balloon := range balloons {
		balloon.draw(renderer)
	}
		
	renderer.Present()
	elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
	fmt.Println("ms per frame: ", elapsedTime)
	if elapsedTime < 5 {
		sdl.Delay(5 - uint32(elapsedTime))
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)

	}
	prevMouseState = currentMouseState
}
}