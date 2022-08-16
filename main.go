package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/flw-cn/go-version"
	"github.com/gdamore/tcell/v2"
	"github.com/nxadm/tail"
	"github.com/rivo/tview"
)

func main() {
	var ver bool
	flag.BoolVar(&ver, "version", false, "print version information")
	flag.Parse()

	if ver {
		version.PrintVersion(os.Stderr, "", "")
		return
	}

	app := tview.NewApplication()

	pages := tview.NewPages()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %v <file1> [<file2> ...]\n", os.Args[0])
		return
	}

	id2name := map[int]string{}
	name2id := map[string]int{}
	id := 0
	for _, fileName := range args {
		if _, ok := name2id[fileName]; ok {
			continue
		}

		file, err := tail.TailFile(
			fileName,
			tail.Config{Follow: true, ReOpen: true},
		)

		if err != nil {
			continue
		}

		id2name[id] = fileName
		name2id[fileName] = id

		textView := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetTextColor(tcell.ColorSilver).
			SetMaxLines(200).
			SetChangedFunc(func() {
				app.Draw()
			})

		pages.AddPage(fmt.Sprintf("%d", id), textView, true, id == 0)

		go func() {
			for line := range file.Lines {
				fmt.Fprintln(textView, tview.TranslateANSI(line.Text))
				textView.ScrollToEnd()
			}
		}()

		id++
	}

	if id == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %v <file1> [<file2> ...]\n", os.Args[0])
		return
	}

	title := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[::bl]" + os.Args[1])

	title.SetTextColor(tcell.ColorLightYellow).
		SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorDarkBlue)

	nextFile := func() {
		name, _ := pages.GetFrontPage()
		index, _ := strconv.Atoi(name)
		next := (index + 1) % pages.GetPageCount()
		title.SetText("[::bl]" + id2name[next])
		pages.SwitchToPage(fmt.Sprintf("%d", next))
	}

	frame := tview.NewPages().
		AddPage("files", pages, true, true).
		AddPage("title", tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(title, 1, 1, false), 20, 1, false), true, true)

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			nextFile()
			app.Draw()
		}
	}()

	if err := app.SetRoot(frame, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
