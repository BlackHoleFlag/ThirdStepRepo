package main										//all thanks to :
													//Jack Mott - https://www.twitch.tv/jackmott42
													//Stefan Gustavson, 2003-2005 Contact: stefan.gustavson@liu.se

import (
	"os"
	"fmt"
	"time"
	"image/png"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 800, 600			//using this constant for future window

type texture struct {
	pos
	pixels []byte
	w,h, pitch int									//pitch = width * size of pixel
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

func(tex *texture) draw(pixels []byte){
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >=0 && screenY < winHeight {
				texIndex := y *tex.pitch + x * 4					//reads data from texture
				screenIndex := screenY * winWidth * 4 + screenX * 4 //copy data on a screen

				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}
func(tex *texture) drawAlpha(pixels []byte){							//more expensive than "draw", but work with pics with alfa value
	for y := 0; y < tex.h; y++ {
		screenY := y + int(tex.y)
		for x := 0; x < tex.w; x++ {
			
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < winWidth && screenY >=0 && screenY < winHeight {
				texIndex := y *tex.pitch + x * 4					//reads data from texture
				screenIndex := screenY * winWidth * 4 + screenX * 4 //copy data on a screen

				srcR := int(tex.pixels[texIndex])
				srcG := int(tex.pixels[texIndex+1])
				srcB := int(tex.pixels[texIndex+2])
				srcA := int(tex.pixels[texIndex+3])

				dstR := int(pixels[screenIndex])
				dstG := int(pixels[screenIndex+1])
				dstB := int(pixels[screenIndex+2])
				
				rstrR := (srcR * 255 + dstR *(255 - srcA))/255   //taking out alfa and scale it back down to 255
				rstrG := (srcG * 255 + dstG *(255 - srcA))/255
				rstrB := (srcB * 255 + dstB *(255 - srcA))/255

				pixels[screenIndex] = byte(rstrR)
				pixels[screenIndex+1] = byte(rstrG)
				pixels[screenIndex+2] = byte(rstrB)
				//pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}


func loadBalloons()[]texture{

	balloonStrs := []string{"balloons/balloon_red.png", "balloons/balloon_green.png", "balloons/balloon_blue.png"}
	balloonTextures := make([]texture, len(balloonStrs))

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
	balloonTextures[i] = texture{pos{0,0}, balloonPixels, w, h, w *4}
	}
	return balloonTextures
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

tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))//using renderer to create textures
if err != nil {
	fmt.Println(err)
	return
}
defer tex.Destroy()

pixels := make([]byte, winWidth * winHeight * 4)  //we going to use "pixels"(slice of byte) to draw on a screen
balloonTextures := loadBalloons()						//balloon textures
dir := 1												//dirrection for balloon

for {
	frameStart := time.Now()
	for event := sdl.PollEvent(); event !=nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		}
	}
	clear(pixels)

	for _, tex := range balloonTextures {
		tex.drawAlpha(pixels)
	}
	balloonTextures[1].x += float32(1*dir)
	if balloonTextures[1].x > 400 || balloonTextures[1].x < 0 {
		dir = dir * -1
	}
	
	tex.Update(nil, pixels, winWidth * 4)
	renderer.Copy (tex, nil, nil)
	renderer.Present()
	elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
	fmt.Println("ms per frame: ", elapsedTime)
	if elapsedTime < 5 {
		sdl.Delay(5 - uint32(elapsedTime))

	}
	sdl.Delay(16)
}
}