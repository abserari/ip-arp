package main

import (
	"github.com/abserari/ip-arp/fing"
)

func main() {
	f := new(fing.Fing)
	f.Detect()
	f.Show()
}
