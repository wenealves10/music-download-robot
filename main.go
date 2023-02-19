package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

func main() {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create context
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var links []string

	err := chromedp.Run(ctx, getLinks(&links))

	if err != nil {
		panic(err)
	}

	channel := make(chan bool)
	lastIndex := 0
	countChannel := 0

	for i := 1; i <= len(links); i++ {
		go func(i int) {
			err := downloadMusic(ctx, links[i-1])

			if err != nil {
				log.Println(err)
			}
			channel <- true
		}(i)

		countChannel++

		if countChannel%50 == 0 {

			for k := lastIndex; k < countChannel; k++ {
				log.Println(<-channel)
			}

			lastIndex = countChannel
		}
	}

	for i := lastIndex; i < countChannel; i++ {
		log.Println(<-channel)
	}
}

func getLinks(links *[]string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(`https://www.youtube.com/playlist?list=PLQEUs9xtYT6o8QYY-NYg5hRuX-8fLrcy3`),
		chromedp.WaitVisible(`#contents > ytd-playlist-video-renderer #video-title`, chromedp.ByID),
		chromedp.EvaluateAsDevTools(`Array.from(document.querySelectorAll("#contents > ytd-playlist-video-renderer #video-title")).map(element => element.href)`, &links),
	}
}

func downloadMusic(ctx context.Context, url string) error {

	var urlDowmload string

	ctxx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := chromedp.Run(ctxx, downloadMusicTasks(url, &urlDowmload))

	if err != nil {
		return err
	}

	fmt.Println(urlDowmload)

	err = downloadToFile(urlDowmload)

	if err != nil {
		return err
	}

	return nil
}

func downloadMusicTasks(url string, urlDowmload *string) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.Navigate(`https://yt1s.com/pt138/youtube-to-mp3`),
		chromedp.WaitVisible(`#s_input`, chromedp.ByID),
		chromedp.SendKeys(`#s_input`, url, chromedp.ByID),
		chromedp.Click(`#btn-convert`, chromedp.ByID),
		chromedp.WaitVisible(`#btn-action`, chromedp.ByID),
		chromedp.Click(`#btn-action`, chromedp.ByID),
		chromedp.WaitVisible(`#asuccess`, chromedp.ByID),
		chromedp.EvaluateAsDevTools(`document.querySelector("a#asuccess").href`, &urlDowmload),
	}
}

func downloadToFile(url string) error {

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	nameAleatorio := uuid.New().String()

	// dir
	_, err = os.Stat("musicas")

	if os.IsNotExist(err) {
		os.Mkdir("musicas", 0755)
	}

	out, err := os.Create("musicas/" + nameAleatorio + ".mp3")

	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return err
	}

	return nil
}
