// external ip :http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// mac: http://godoc.org/github.com/j-keck/arping
//      http://golangtc.com/t/52d26aa7320b5237d1000044
// vendor: http://zhidao.baidu.com/question/37072459.html

package fing

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/abserari/quicksilver/fing"
	oui "github.com/dutchcoders/go-ouitools"
	"github.com/j-keck/arping"
)

func examplemain() {
	fing := new(Fing)
	fing.Detect()
	fing.Show()
}

type Fing struct {
	Devices []*Device
}

type Device struct {
	Ip       string
	Mac      string
	Vendor   string
	Type     int
	HostName string
}

func NewDevice(ip, mac, vendor, hostname string, t int) *Device {
	device := new(Device)
	device.Ip = ip
	device.Mac = mac
	device.Vendor = vendor
	device.Type = t
	device.HostName = hostname
	return device
}

func (this *Fing) Detect() {
	// Get own IP
	maps, err := ExternalIP()
	if err != nil {
		log.Println(err)
		return
	}
	var wg sync.WaitGroup
	for ip, ownmac := range maps {
		vendor, err := Vendor(ownmac)
		if err != nil {
			log.Println("err :", err, "\n", vendor, " this", ip, "and mac", ownmac, "not match vendor")
			vendor = "unknown"
		}
		dnsname, err := net.LookupAddr(ip)
		if err != nil {
			dnsname = []string{"unknown"}
		}
		this.Devices = append(this.Devices, NewDevice(ip, ownmac, vendor, dnsname[0], TYPE_OWN_DEVICE))

		ipFormat := ip[:strings.LastIndex(ip, ".")+1] + "%d"

		for i := 1; i <= 255; i++ {
			nextIp := fmt.Sprintf(ipFormat, i)
			if nextIp != ip {
				wg.Add(1)
				go func() {
					hwAddr, _, err := fing.Mac(nextIp)
					if err == arping.ErrTimeout {
						log.Printf("IP %s is offline.\n", nextIp)
					} else if err != nil && err.Error() == "The network name cannot be found." {

					} else if err != nil {
						log.Printf("IP %s :%s\n", nextIp, err.Error())
					} else {
						// log.Printf("%s (%s) %d usec\n", nextIp, hwAddr, duration/1000)
						vendor, err := Vendor(hwAddr.String())
						if err != nil {
							// log.Println("err :", err, "\n", vendor, " this", ip, "and mac", ownmac, "not match and jump")
							vendor = "unknown"
						}
						dnsname, err := net.LookupAddr(ip)
						// log.Println(dnsname)
						if err != nil {
							dnsname = []string{"unknown"}
						}
						this.Devices = append(this.Devices, NewDevice(nextIp, hwAddr.String(), vendor, dnsname[0], TYPE_OTHER_DEVICE))
					}
					wg.Done()
				}()
			}
		}
	}

	wg.Wait()
}

func (this *Fing) Show() {
	fmt.Printf("%3s|%15s|%17s|%20s|%4s|%20s\n", "#", "IP", "MAC", "VENDOR", "TYPE", "HOSTNAME")
	for i, device := range this.Devices {
		fmt.Printf("%3d|%15s|%17s|%20s|%4s|%20s\n", i, device.Ip, device.Mac, device.Vendor, this.showType(device.Type), device.HostName)
	}
}

func (this *Fing) showType(t int) string {
	switch t {
	case TYPE_OWN_DEVICE:
		return "OWN"
	}
	return ""
}

const (
	TYPE_OWN_DEVICE = iota
	TYPE_OTHER_DEVICE
)

var db *oui.OuiDb = oui.New("./oui.txt")

func Vendor(mac string) (string, error) {
	if db == nil {
		log.Panicln("database not initialized")
	}

	macs := strings.Split(mac, ":")
	if len(macs) != 6 {
		return "", fmt.Errorf("MAC Error: %s", mac)
	}
	vendor, err := db.VendorLookup(mac)
	if err != nil {
		return "", fmt.Errorf("Vendor Error: %s", err)
	}
	return vendor, nil
	// return strings.Split(vendor, "\n")[0], nil
}

func ExternalIP() (map[string]string, error) {
	maps := make(map[string]string)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		// log.Println(iface)
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println("error ", err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			maps[ip.String()] = iface.HardwareAddr.String()
		}
	}
	// return "", "", errors.New("are you connected to the network?")
	return maps, nil
}
