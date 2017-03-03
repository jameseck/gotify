package main

import (
	"encoding/json"
	"fmt"
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
	//	"github.com/fsnotify/fsnotify"
	"log"
	"os/exec"
	"sync"
	"time"
)

var (
	mutex = &sync.Mutex{}
)

type Rbd struct {
	Name        string
	Pool        string
	Map         RbdMap
	MappedSince time.Time
	Locks       []rbd.Locker
	Watchers    []Watcher
}

type RbdMap struct {
	Name   string `json:"name"`
	Pool   string `json:"pool"`
	Snap   string `json:"snap"`
	Device string `json:"device"`
}

// {"watchers":{"watcher":{"address":"192.168.3.2:0\/654154289","client":8937,"cookie":2}}}
type JSONWatchers struct {
	Watchers []JSONWatcher
}

type JSONWatcher struct {
	watcher string `json:"watcher"`
	Watch   Watcher
}

type Watcher struct {
	Address string `json:"address"`
	Client  int    `json:"client"`
	Cookie  int    `json:"cookie"`
}

func (rbd *Rbd) Duration() time.Duration {
	return time.Now().Sub(rbd.MappedSince)
}

func Locks(maps map[string]Rbd) map[string]Rbd {
	fmt.Println("Entering Locks func")

	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()
	iocx, err := conn.OpenIOContext("rbd")
	if err != nil {
		fmt.Println("Locks error opening iocontext: ", err)
	}
	log.Printf("maps: %-v\n", maps)

	log.Println("Getting rbd names")
	names, err := rbd.GetImageNames(iocx)
	if err != nil {
		fmt.Println("Locks error getting image names: ", err)
	}

	for i := range names {
		// populate maps with image names
		var r Rbd
		r.Name = names[i]
		r.Pool = "rbd"
		maps[names[i]] = r

		fmt.Println("Getting rbd image: ", names[i])
		img := rbd.GetImage(iocx, names[i])
		fmt.Println("Opening rbd image: ", names[i])
		img.Open(false) // false = read-only
		fmt.Println("ListLockers on rbd image: ", names[i])
		_, lockers, err := img.ListLockers()
		if err != nil {
			fmt.Printf("Error listing lockers for %s: %s\n", err, names[i])
		}
		for k := range lockers {
			fmt.Printf("Image %s is locked by %s\n", names[i], lockers[k])
			if r, exists := maps[names[i]]; exists {
				l := r.Locks
				l = append(l, lockers[k])
				r.Locks = l
				maps[names[i]] = r
			}
		}
		fmt.Println("Closing rbd image: ", names[i])
		img.Close()

		log.Printf("looking for watchers on %s\n", names[i])
		r = maps[names[i]]
		w := Listwatchers(r)
		rw := r.Watchers
		for ww := range w {
			log.Printf("Found watcher on rbd %s\n", names[i])
			rw = append(rw, rw[ww])
		}
		r.Watchers = rw
		maps[names[i]] = r
	}

	log.Printf("Finished Locks func\n")
	return maps
}

// Pass an Rbd and receive a slice of watchers
func Listwatchers(rbd Rbd) []Watcher {

	var watchers []Watcher

	//w := []Watcher{}

	out, err := exec.Command("/usr/bin/rbd", "-p", rbd.Pool, "status", rbd.Name, "--format", "json").Output()
	fmt.Printf("out: %s", out)
	if err != nil {
		log.Fatal(err)
	}

	var watchmap map[string]*json.RawMessage
	err = json.Unmarshal(out, &watchmap)

	var w JSONWatchers
	err = json.Unmarshal(*watchmap["watchers"], &w)

	for _, v := range w.Watchers {
		fmt.Printf("%v. %v. %v", v.Watch.Address, v.Watch.Client, v.Watch.Cookie)
		watchers = append(watchers, v.Watch)
	}

	return watchers
}

func Showmapped(rbds map[string]Rbd) map[string]Rbd {

	var maps []RbdMap

	out, err := exec.Command("/usr/bin/rbd", "showmapped", "--format", "json").Output()
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(out, &maps)
	if err != nil {
		log.Fatal(err)
	}
	for k, _ := range maps {
		r := rbds[maps[k].Name]
		r.MappedSince = time.Now()
		rbds[maps[k].Name] = r
	}
	return rbds
}

func main() {

	mutex.Lock()
	rbdmaps := make(map[string]Rbd)
	rbdmaps = Showmapped(rbdmaps)
	rbdmaps = Locks(rbdmaps)
	mutex.Unlock()

	// 	watcher, err := fsnotify.NewWatcher()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer watcher.Close()

	//	ticker := time.NewTicker(time.Millisecond * 2000)
	//	go func() {
	//		for t := range ticker.C {
	//			fmt.Printf("Tick at", t)
	//			mutex.Lock()
	//			for _, v := range rbdmaps {
	//				fmt.Printf("%v: mapped for: %v\n", v.Name, v.Duration())
	//			}
	//			mutex.Unlock()
	//		}
	//	}()

	// 	done := make(chan bool)
	// 	go func() {
	// 		for {
	// 			select {
	// 			case event := <-watcher.Events:
	// 				switch event.Op {
	// 				case fsnotify.Create:
	// 					mutex.Lock()
	// 					m := rbdmaps[event.Name]
	// 					m.MappedSince = time.Now()
	// 					rbdmaps[event.Name] = m
	// 					mutex.Unlock()
	// 					log.Printf("Create %s detected", event.Name)
	// 				case fsnotify.Remove:
	// 					mutex.Lock()
	// 					delete(rbdmaps, event.Name)
	// 					mutex.Unlock()
	// 					log.Printf("Remove %s detected", event.Name)
	// 				}
	// 			case err := <-watcher.Errors:
	// 				log.Println("error:", err)
	// 			}
	// 		}
	// 	}()

	// 	err = watcher.Add("/dev/rbd/rbd")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	<-done
}
