/*
Copyright 2011 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Example of using the rfb package.
//
// Author: Brad Fitzpatrick <brad@danga.com>

package main

import (
	"flag"
	"github.com/bradfitz/rfbgo/rfb"
	"github.com/kbinani/screenshot"
	"image"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"time"
)

var (
	listen  = flag.String("listen", ":5900", "listen on [ip]:port")
	profile = flag.Bool("profile", false, "write a cpu.prof file when client disconnects")
)

const (
//width  = 640
//height = 480
)

func main() {
	flag.Parse()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	width := screenshot.GetDisplayBounds(0).Dx()
	height := screenshot.GetDisplayBounds(0).Dy()
	s := rfb.NewServer(width, height)
	go func() {
		err = s.Serve(ln)
		log.Fatalf("rfb server ended with: %v", err)
	}()
	for c := range s.Conns {
		handleConn(c)
	}
}

func handleConn(c *rfb.Conn) {
	if *profile {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("profiling CPU")
		defer pprof.StopCPUProfile()
		defer log.Printf("stopping profiling CPU")
	}
	bounds := screenshot.GetDisplayBounds(0)
	im := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	li := &rfb.LockableImage{Img: im}

	closec := make(chan bool)
	go func() {
		slide := 0
		tick := time.NewTicker(time.Second / 10)
		defer tick.Stop()
		haveNewFrame := false
		for {
			feed := c.Feed
			if !haveNewFrame {
				feed = nil
			}
			_ = feed
			select {
			case feed <- li:
				haveNewFrame = false
			case <-closec:
				return
			case <-tick.C:
				slide++
				li.Lock()
				im = drawImage(slide)
				li.Img = im
				li.Unlock()
				haveNewFrame = true
			}
		}
	}()

	for e := range c.Event {
		log.Printf("got event: %#v", e)
	}
	close(closec)
	log.Printf("Client disconnected")
}

func drawImage(anim int) *image.RGBA {
	//pos := 0
	//const border = 50
	//for y := 0; y < height; y++ {
	//	for x := 0; x < width; x++ {
	//		var r, g, b uint8
	//		switch {
	//		case x < border*2.5 && x < int((1.1+math.Sin(float64(y+anim*2)/40))*border):
	//			r = 255
	//		case x > width-border*2.5 && x > width-int((1.1+math.Sin(math.Pi+float64(y+anim*2)/40))*border):
	//			g = 255
	//		case y < border*2.5 && y < int((1.1+math.Sin(float64(x+anim*2)/40))*border):
	//			r, g = 255, 255
	//		case y > height-border*2.5 && y > height-int((1.1+math.Sin(math.Pi+float64(x+anim*2)/40))*border):
	//			b = 255
	//		default:
	//			r, g, b = uint8(x+anim), uint8(y+anim), uint8(x+y+anim*3)
	//		}
	//		im.Pix[pos] = r
	//		im.Pix[pos+1] = g
	//		im.Pix[pos+2] = b
	//		pos += 4 // skipping alpha
	//	}
	//}
	img, err := screenshot.CaptureRect(screenshot.GetDisplayBounds(0))
	if err != nil {
		panic(err)
	}
	return img

}
