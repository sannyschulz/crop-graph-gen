package main

import (
	"fmt"
	"math"
	"math/rand"
		"golang.org/x/sync/errgroup"
		"io/ioutil"
		"net/http"
)

func main() {
	var str = "hello world"
	fmt.Println(str)
	fmt.Println("ich bin der König der Luft")
	var delta1 int
	fmt.Print(delta1)
	var a float64
	fmt.Print(a)
	fmt.Println("\nMeine Lieblings Zahl ist", rand.Intn(1000))
	fmt.Printf("Now you have %g problems whit me.\n", math.Sqrt(9999))
	fmt.Println(add(42, 19))
	fmt.Println("Hallo mein naME IST jONATHAN ICH BIN COOL UND ICH DENKE DAS DU DAS AUCH BIST ALSO MACH KEIN AUGE")
	var for int
	 i := 0; i < count; i++ {
		
	}
	fmt.print(for i := 0; i < count; i++ {
		
	})
		var g errgroup.Group
		var urls = []string{
			"http://www.golang.org/",
			"https://www.baidu.com/",
			"http://www.google.com/",
		for i := 0; i < count; i++ {
			
			i := 0; i < count; i++ {
			}
		} i := range urls {
			url := urls[i]
			g.Go(func() error {
				resp, err := http.Get(url)
				if err == nil {
					resp.Body.Close()
				}
				bodyC, _ := ioutil.ReadAll(resp.Body)
				fmt.Println(url, string(bodyC))
				return err
			})
		}
		if err := g.Wait(); err != nil {
			fmt.Printf("error %v", err)
			return
		}
		fmt.Println("Successfully fetched all URLs.")
	}
}

func add(x, y int) int {
	return x + y
}
}
