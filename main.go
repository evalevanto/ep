// Demo code for the Table primitive.
package main

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Emoji struct {
	Keywords []string `json:"keywords"`
	Char     string   `json:"char"`
}

func getEmojis() (map[string][]Emoji, error) {
	raw, err := ioutil.ReadFile("emojis.json")
	if err != nil {
		return nil, err
	}

	nameMap := make(map[string]Emoji)
	err = json.Unmarshal(raw, &nameMap)
	if err != nil {
		return nil, err
	}

	keywordMap := make(map[string][]Emoji)
	for name, emoji := range nameMap {
		keywordMap[name] = append(keywordMap[name], emoji)
		for _, keyword := range emoji.Keywords {
			keywordMap[keyword] = append(keywordMap[keyword], emoji)
		}
	}

	return keywordMap, nil
}

func filterEmojis(emojis map[string][]Emoji, query string) []string {
	justEmojis := []string{}
	for key, e := range emojis {
		if !strings.Contains(key, query) {
			continue
		}
		for _, emoji := range e {
			justEmojis = append(justEmojis, emoji.Char)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(justEmojis)))
	return justEmojis
}

func drawEmojis(table *tview.Table, emojis map[string][]Emoji, query string) {
	filteredEmojis := filterEmojis(emojis, query)
	numCols := 10
	table.Clear()
	for word := 0; word < len(filteredEmojis); word++ {
		r, c := word/numCols, word%numCols
		table.SetCell(r, c,
			tview.NewTableCell(" "+filteredEmojis[word]+" "))
	}
	table.ScrollToBeginning()
	table.Select(0, 0)
}

func main() {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, true).
		SetFixed(0, 0)

	inputField := tview.NewInputField().
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(table)
			table.SetSelectable(true, true)
		})

	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown {
			app.SetFocus(table)
			table.SetSelectable(true, true)
		}
		return event
	})

	grid := tview.NewGrid().
		SetRows(1, 1).
		SetColumns(1, 1).
		AddItem(inputField, 0, 0, 1, 3, 0, 0, true).
		AddItem(table, 2, 0, 1, 3, 0, 0, false)
	grid.SetBorder(true).SetRect(0, 0, 60, 25)
	grid.SetTitle("Emoji Picker")

	emojis, _ := getEmojis()
	drawEmojis(table, emojis, "")

	inputField.SetChangedFunc(func(text string) {
		drawEmojis(table, emojis, text)
	})

	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
			app.Stop()
		}
	}).SetSelectedFunc(func(row int, column int) {
		cell := table.GetCell(row, column)
		clipboard.WriteAll(cell.Text)
		app.Stop()
	}).SetSelectable(false, false)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		if event.Key() == tcell.KeyUp && row == 0 {
			app.SetFocus(inputField)
			table.SetSelectable(false, false)
		} else if event.Key() == tcell.KeyRune {
			inputField.SetText(inputField.GetText() + string(event.Rune()))
			app.SetFocus(inputField)
			table.SetSelectable(false, false)
		}
		return event
	})

	if err := app.SetRoot(grid, false).Run(); err != nil {
		panic(err)
	}
}