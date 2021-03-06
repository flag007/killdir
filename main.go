package main

import (
	"net/url"
	"encoding/json"
	"os"
	"strings"
	"bufio"
	"fmt"
	"flag"
	"github.com/logrusorgru/aurora"
	"sync"
	"io/ioutil"
)

type (
	result struct {
		Status  int `json:"status"`
		Length  int    `json:"length"`
		Words   int    `json:"words"`
		Lines   int    `json:"lines"`
		Url     string `json:"url"`
	}

	FFUFResult struct {
		Time string `json:"time"`
		Results []result `json:"results"`
	}

)

var au aurora.Aurora
var details bool

func init() {
	au = aurora.NewAurora(true)
}

type SafeMap struct {
	sync.RWMutex
	Map map[string]int
}

func newSafeMap() *SafeMap {
	sm := new(SafeMap)
	sm.Map = make(map[string]int)
	return sm
}

func (sm *SafeMap) okMap(key string) bool {
	sm.RLock()
	_, ok := sm.Map[key]
	sm.RUnlock()
	return ok
}

func (sm *SafeMap) readMap(key string) int {
	sm.RLock()
	value := sm.Map[key]
	sm.RUnlock()
	return value
}

func (sm *SafeMap) writeMap(key string) {
	sm.Lock()
	value := sm.Map[key]
	sm.Map[key] = value+1
	sm.Unlock()
}


func main() {
	dir_dicc := newSafeMap()
	var dirAlls []string
	dirAllCh1 := make(chan string)
	var dirAllWG1 sync.WaitGroup

	dirAllCh2 := make(chan string)
	var dirAllWG2 sync.WaitGroup

	output := make(chan string)

	var file string
	flag.StringVar(&file, "f", "", "读取文件")

	flag.BoolVar(&details, "v", false, "输出详情")

	var concurrency int
	flag.IntVar(&concurrency, "c", 50, "设置线程")

	var threshold int
	flag.IntVar(&threshold, "t", 50, "设置阈值")

	flag.Parse()

	if details {
		str := `
    
▄███▄   ██      ▄▄▄▄▄ ▀▄    ▄ ▄ ▄   ▄█    ▄   
█▀   ▀  █ █    █     ▀▄ █  █ █   █  ██     █  
██▄▄    █▄▄█ ▄  ▀▀▀▀▄    ▀█ █ ▄   █ ██ ██   █ 
█▄   ▄▀ █  █  ▀▄▄▄▄▀     █  █  █  █ ▐█ █ █  █ 
▀███▀      █           ▄▀    █ █ █   ▐ █  █ █ 
          █                   ▀ ▀      █   ██ 
         ▀                                    

           `
    fmt.Println(au.Magenta(str))
  }

	for i := 0; i < concurrency; i++ {
		dirAllWG1.Add(1)
		go func() {
			for dirAll := range dirAllCh1 {
				s := strings.Fields(dirAll)
				u, err := url.Parse(s[2])
				if err != nil {
					continue
				}

				dir_dicc.writeMap(u.Host+"|"+s[1])

				//fmt.Println(u.Host+"|"+s[1])

				//fmt.Println(u.Path)
			}
			dirAllWG1.Done()
		}()
	}

	var data FFUFResult

	if file != "" {
		bytes, _ := ioutil.ReadFile(file)
		json.Unmarshal(bytes, &data)

		for _, i := range data.Results {
			dirAll := fmt.Sprintf("%-16d%-16d%s", i.Status, i.Length, i.Url)
			dirAllCh1 <- dirAll
			dirAlls = append(dirAlls, dirAll)
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			dirAll := strings.ToLower(sc.Text())
			dirAllCh1 <- dirAll
			dirAlls = append(dirAlls, dirAll)
		}

		if err := sc.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
		}
	}


	close(dirAllCh1)
	dirAllWG1.Wait()

	if details {
		for key, value := range dir_dicc.Map{
			fmt.Println(au.Yellow(key), au.Yellow(value))
		}
		fmt.Println()
		fmt.Println()
	}

	var inputWG sync.WaitGroup

	inputWG.Add(1)

	go func(){
		for _,dirAll := range dirAlls {
			dirAllCh2 <- dirAll
		}
		inputWG.Done()
	}()


	go func() {
		inputWG.Wait()
		close(dirAllCh2)
	}()


	for i := 0; i < concurrency; i++ {
		dirAllWG2.Add(1)

		go func() {
			for dirAll := range dirAllCh2 {
				s := strings.Fields(dirAll)
				u, err := url.Parse(s[2])
				if err != nil {
					continue
				}

				if u.Path == "/favicon.ico" {
					continue
				}

				if u.Path == "/health" {
					continue
				}

				if u.Path == "/crossdomain.xml" {
					continue
				}


				if  dir_dicc.readMap(u.Host+"|"+s[1]) <= threshold {
					output <- dirAll
				}

			}
			dirAllWG2.Done()
		}()
	}

	go func() {
		dirAllWG2.Wait()
		close(output)
	}()

	var outputWG sync.WaitGroup

	outputWG.Add(1)

	go func() {
		for o := range output {
			if details {
				fmt.Println("[!]", o)
			} else {
				fmt.Println(o)
			}
		}
		outputWG.Done()
	}()

	outputWG.Wait()

}
