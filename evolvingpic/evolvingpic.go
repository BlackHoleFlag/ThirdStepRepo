package main

import(
	"fmt"
	"time"
	."github.com/yawhoo/evolvingpic/apt"
	"github.com/veandco/go-sdl2/sdl"
)


const winWidth, winHeight, winDepth int = 1280, 960, 100			//using this constant for future window

type audioState struct {
	explosionBytes []byte
	deviceID sdl.AudioDeviceID
	audioSpec *sdl.AudioSpec
}

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

func pixelsToTextures(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture{
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func aptToTexture(redNode,  greenNode, blueNode Node, w, h int, renderer *sdl.Renderer) *sdl.Texture{
	scale := float32(255/2)		//scale to 0-255 colors
	offset := float32(-1.0 * scale)
	pixels := make([]byte, w * h * 4)
	pixelIndex := 0

	for yi := 0; yi < h; yi++ {
		y := float32(yi) / float32(h) * 2 - 1
		for xi := 0; xi < w; xi++{
			x := float32(xi) / float32(w) * 2 - 1

			r := redNode.Eval(x, y)
			g := greenNode.Eval(x, y)
			b := blueNode.Eval(x,y) 

			pixels[pixelIndex] = byte(r * scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g * scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b * scale - offset)
			pixelIndex++		//skip alfa 
			pixelIndex++
		}
	}
	return pixelsToTextures(renderer, pixels, w, h)
}

func main(){

	err := sdl.Init(sdl.INIT_EVERYTHING)		//initializing sdl, wont work without it
	if err != nil {
		fmt.Println(err)
		return
	}

defer sdl.Quit() 								//closing properly after everything is done

window, err := sdl.CreateWindow("Eye", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,		// creating window with sdl and our const for window
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
/*audio
var audioSpec sdl.AudioSpec
explosionBytes,_ := sdl.LoadWAV("balloons2/explode.wav")
audioID, err := sdl.OpenAudioDevice("", false, &audioSpec, nil, 0)
if err != nil {
	fmt.Println(err)
} 
defer sdl.FreeWAV(explosionBytes)

audioState := audioState{explosionBytes, audioID, &audioSpec}
*/
sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")	
var elapsedTime float32
currentMouseState := getMouseState()
//prevMouseState := currentMouseState

x := &OpX{}
y := &OpY{}
sine := &OpSin{}
noise := &OpNoise{}
atan2 := &OpMult{}
plus := &OpPlus{}
atan2.LeftChild = x
atan2.RightChild = noise
noise.LeftChild = x
noise.RightChild = y

sine.Child = atan2
plus.LeftChild = y
plus.RightChild = sine



tex := aptToTexture(sine, plus, noise, winWidth, winHeight, renderer)

for {
	frameStart := time.Now()
	currentMouseState = getMouseState()

	for event := sdl.PollEvent(); event !=nil; event = sdl.PollEvent() {
		switch e:= event.(type) {
		case *sdl.QuitEvent:
			return
		case *sdl.TouchFingerEvent:
			if e.Type == sdl.FINGERDOWN{
			touchX := int(e.X * float32(winWidth))
			touchY := int(e.Y * float32(winHeight))
			currentMouseState.x = touchX
			currentMouseState.y = touchY
			currentMouseState.leftButton = true
			}
		}
	}

	

	renderer.Copy(tex, nil, nil)

	renderer.Present()
	elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
	fmt.Println("ms per frame: ", elapsedTime)
	if elapsedTime < 5 {
		sdl.Delay(5 - uint32(elapsedTime))
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)

	}
	//prevMouseState = currentMouseState
}
}
