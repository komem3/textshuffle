package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

const meishi = "名詞"

type contents struct {
	sync.Mutex
	skips []int
	words []string
}

func (cs *contents) skip(i int) {
	cs.Lock()
	defer cs.Unlock()
	cs.skips = append(cs.skips, i)
	cs.words = append(cs.words, "")
}

func (cs *contents) insert(str string) {
	cs.Lock()
	defer cs.Unlock()
	cs.words[cs.skips[0]] = str
	cs.skips = cs.skips[1:]
}

func (cs *contents) add(str string) {
	cs.Lock()
	defer cs.Unlock()
	cs.words = append(cs.words, str)
}

var text = flag.String("t", "", "shuffle text.")
var file = flag.String("f", "", "shuffle file.")

func main() {
	flag.Parse()
	if *text == "" && *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	txt := *text
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		fbytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		txt = string(fbytes)
	}

	tokens := t.Tokenize(txt)
	var cs contents
	var wg sync.WaitGroup
	for i, token := range tokens {
		feat, _ := token.FeatureAt(0)
		if feat == meishi {
			cs.skip(i)
			wg.Add(1)
			go func(str string) {
				defer wg.Done()
				<-time.After(time.Duration(rand.Int()%1000) * time.Millisecond)
				cs.insert(str)
			}(token.Surface)
			continue
		}
		cs.add(token.Surface)
	}
	wg.Wait()
	fmt.Printf("%s\n", strings.Join(cs.words, ""))
}
