package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))

	defer cancel()

	var nodes []*cdp.Node

	selector := "#main ul li a"

	pageURL := "https://notepad-plus-plus.org/downloads/"

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(pageURL),
		chromedp.WaitEnabled(selector),
		chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll),
	}); err != nil {
		log.Fatal(err)
	}

	var waitGroup = new(sync.WaitGroup)

	waitGroup.Add(len(nodes))

	for _, node := range nodes {
		go func(node *cdp.Node) {
			defer waitGroup.Done()

			if err := screenshotPage(ctx, node.AttributeValue("href")); err != nil {
				log.Fatal(err)
			}

		}(node)
	}

	waitGroup.Wait()

}

func screenshotPage(ctx context.Context, url string) error {

	clone, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// run task list
	var buf []byte
	err := chromedp.Run(clone,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Screenshot(`body`, &buf, chromedp.NodeVisible, chromedp.ByQuery),
	)
	if err != nil {
		return err
	}

	// if dir screenshots not exist, create it
	if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
		os.Mkdir("screenshots", 0755)
	}

	// save to disk
	filename := "screenshots/" + time.Now().Format("2006-01-02 15:04:05") + ".png"

	return os.WriteFile(filename, buf, 0644)
}
