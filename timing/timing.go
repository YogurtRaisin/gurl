package timing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
)

// When receive all performance timing information

// type PerformanceTiming struct {
// 	NavigationStart            int64 `json:"navigationStart"`
// 	UnloadEventStart           int64 `json:"unloadEventStart"`
// 	UnloadEventEnd             int64 `json:"unloadEventEnd"`
// 	RedirectStart              int64 `json:"redirectStart"`
// 	RedirectEnd                int64 `json:"redirectEnd"`
// 	FetchStart                 int64 `json:"fetchStart"`
// 	DomainLookupStart          int64 `json:"domainLookupStart"`
// 	DomainLookupEnd            int64 `json:"domainLookupEnd"`
// 	ConnectStart               int64 `json:"connectStart"`
// 	ConnectEnd                 int64 `json:"connectEnd"`
// 	SecureConnectionStart      int64 `json:"secureConnectionStart"`
// 	RequestStart               int64 `json:"requestStart"`
// 	ResponseStart              int64 `json:"responseStart"`
// 	ResponseEnd                int64 `json:"responseEnd"`
// 	DomLoading                 int64 `json:"domLoading"`
// 	DomInteractive             int64 `json:"domInteractive"`
// 	DomContentLoadedEventStart int64 `json:"domContentLoadedEventStart"`
// 	DomContentLoadedEventEnd   int64 `json:"domContentLoadedEventEnd"`
// 	DomComplete                int64 `json:"domComplete"`
// 	LoadEventStart             int64 `json:"loadEventStart"`
// 	LoadEventEnd               int64 `json:"loadEventEnd"`
// }

var (
	msgChann = make(chan cdproto.Message)
	ctxt     context.Context
)

func devToolHandler(s string, is ...interface{}) {
	/*
	 Uncomment the following line to have a log of the events
	 log.Printf(s, is...)
	*/
	/*
	 We need this to be on a separate gorutine
	 otherwise we block the browser and we don't receive messages
	*/
	go func() {

		for _, elem := range is {
			var msg cdproto.Message
			// The CDP messages are sent as strings so we need to convert them back
			json.Unmarshal([]byte(fmt.Sprintf("%s", elem)), &msg)
			msgChann <- msg
		}
	}()
}

func Timing(url string) {
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// path := `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`
	// if runtime.GOOS != "windows" {
	// 	path = "/usr/bin/google-chrome"
	// }

	// create chrome instance
	c, err := chromedp.New(ctxt,
		chromedp.WithRunnerOptions(
			runner.Flag("disable-infobars", true),
			// When headless
			// runner.Path(path),
			// runner.Flag("headless", true),
			// runner.Flag("disable-gpu", true),
			// runner.Flag("no-first-run", true),
			// runner.Flag("no-sandbox", true),
			// runner.Flag("no-default-browser-check", true),
		),
		chromedp.WithLog(devToolHandler))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Run(ctxt, network.Enable())
	if err != nil {
		log.Println(err)
	}

	c.Run(ctxt, chromedp.ActionFunc(func(_ context.Context, h cdp.Executor) error {
		// This is in an action function because the Executor is needed in some method
		go func() {
			for {
				/*
					Right now I have no guaranties (did not run enough tests) that the messages are evaluated
					in the same order they are received
				*/
				msg := <-msgChann

				switch msg.Method.String() {
				case "Page.frameScheduledNavigation":
					var schednavevent page.EventFrameScheduledNavigation
					json.Unmarshal(msg.Params, &schednavevent)
					// fmt.Println("Page.frameScheduledNavigation :", schednavevent)
					// fmt.Println(schednavevent.URL)

				case "Network.requestWillBeSent":
					var reqWillSend network.EventRequestWillBeSent
					json.Unmarshal(msg.Params, &reqWillSend)
					// fmt.Println("Network.requestWillBeSent :", reqWillSend)
					// fmt.Println(reqWillSend.Request.URL)

				case "Network.loadingFinished":
					var loadFinished network.EventLoadingFinished
					json.Unmarshal(msg.Params, &loadFinished)
					// fmt.Println("Network.loadingFinished :", loadFinished)
					// fmt.Println(loadFinished.Timestamp.Time())

				case "Network.responseReceived":
					var respevent network.EventResponseReceived
					json.Unmarshal(msg.Params, &respevent)
					// This contains a bunch of data, like the response headers
					// fmt.Println("Network.responseReceived :", respevent)
					// fmt.Println(respevent.Response)
					// Uncomment the following lines if you want to print the response body
					// rbp := network.GetResponseBody(respevent.RequestID)
					// b, e := rbp.Do(ctxt, h)
					// if e != nil {
					// 	fmt.Println(e)
					// 	continue
					// }
					// fmt.Printf("%s\n", b)

				default:
					continue
				}
			}
		}()
		return nil
	}))

	err = c.Run(ctxt, chromedp.Navigate(url))
	if err != nil {
		log.Fatal(err)
	}

	/*
		Since chromedp.Navigate does not wait for the page to be fully loaded
		we wait manually, there may be a better and more reliable way to do this
	*/
	// var performance PerformanceTiming
	var navStart int64
	var resEnd int64
	var loadEnd int64
	state := "notloaded"
	for {
		script := `document.readyState`
		err = c.Run(ctxt, chromedp.EvaluateAsDevTools(script, &state))
		if err != nil {
			log.Println(err)
		}
		// fmt.Println("state :", state)
		if strings.Compare(state, "complete") == 0 {
			err = c.Run(ctxt, chromedp.EvaluateAsDevTools(`window.performance.timing.navigationStart`, &navStart))
			if err != nil {
				log.Println(err)
			}
			err = c.Run(ctxt, chromedp.EvaluateAsDevTools(`window.performance.timing.responseEnd`, &resEnd))
			if err != nil {
				log.Println(err)
			}
			err = c.Run(ctxt, chromedp.EvaluateAsDevTools(`window.performance.timing.loadEventEnd`, &loadEnd))
			if err != nil {
				log.Println(err)
			}
			resEndTime := resEnd - navStart
			loadEndTime := loadEnd - navStart
			fmt.Println("responseEnd :", resEndTime, "ms")
			fmt.Println("loadEnd :", loadEndTime, "ms")
			break
		}
	}

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish the shutdown
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	// Wait() will hang on Windows and Linux with Chrome headless mode
	// we'll need to exit the program when this happens
	// ch := make(chan error)
	// go func() {
	// 	c.Wait()
	// 	ch <- nil
	// }()

	// select {
	// case err = <-ch:
	// 	log.Println("chrome closed")
	// case <-time.After(10 * time.Second):
	// 	log.Println("chrome didn't shutdown within 10s")
	// }
}
