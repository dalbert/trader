/*
 *   Copyright 2019 Tero Vierimaa
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/twidgets"
	"tryffel.net/pkg/bookmarker/config"
	"tryffel.net/pkg/bookmarker/storage"
	"tryffel.net/pkg/bookmarker/storage/models"
	"tryffel.net/pkg/bookmarker/ui/modals"
)

var navBarLabels = make([]string, 0)
var navBarShortucts = make([]tcell.Key, 0)

type Window struct {
	app *tview.Application
	db  *storage.Database

	layout   *twidgets.ModalLayout
	grid     *tview.Grid
	gridAxis []int
	gridSize int

	navBar    *twidgets.NavBar
	project   *Projects
	tags      *Tags
	bookmarks *BookmarkTable
	metadata  *Metadata

	help         *modals.Help
	bookmarkForm *modals.BookmarkForm

	hasModal  bool
	modal     twidgets.Modal
	lastFocus tview.Primitive

	createFunc func(bookmark *models.Bookmark)

	metadataOpen bool
}

func (w *Window) Draw(screen tcell.Screen) {
	w.grid.Draw(screen)
}

func (w *Window) GetRect() (int, int, int, int) {
	return w.grid.GetRect()
}

func (w *Window) SetRect(x, y, width, height int) {
	w.grid.SetRect(x, y, width, height)
}

func (w *Window) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		//key := event.Key()
		//if key == tcell.KeyCtrlSpace {
		//	w.openMetadata()
		//} else {
		w.grid.InputHandler()(event, setFocus)
		//}
	}
}

func (w *Window) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	navbar := config.Configuration.Shortcuts.NavBar
	key := event.Key()
	switch key {
	case navbar.Menu:
	case navbar.Help:
		if !w.hasModal {
			w.addModal(w.help, 10, 40, true)
			w.help.Update()
		}
	case navbar.NewBookmark:
		w.addModal(w.bookmarkForm, 10, 40, false)
	case tcell.KeyEscape:
		if w.hasModal {
			w.layout.RemoveModal(w.modal)
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			w.modal = nil
			w.hasModal = false
		} else if w.metadataOpen {
			w.closeMetadata(false, nil)
		}
	case tcell.KeyCtrlSpace:
		if !w.metadataOpen || !w.hasModal {
			w.openMetadata()
		}
	default:
		return event
	}
	return nil

}

func (w *Window) addModal(modal twidgets.Modal, h, width uint, lockSize bool) {
	if !w.hasModal {
		w.layout.AddModal(modal, h, width, lockSize)

		w.lastFocus = w.app.GetFocus()
		w.app.SetFocus(modal)
		w.modal = w.bookmarkForm
		w.hasModal = true
	}
}

func (w *Window) Focus(delegate func(p tview.Primitive)) {
	w.grid.Focus(delegate)
}

func (w *Window) Blur() {
	w.grid.Blur()
}

func (w *Window) GetFocusable() tview.Focusable {
	return w.layout.GetFocusable()
}

//func (w *Window) HasFocus() bool {
//	return w.layout
//	focus :=  w.grid.HasFocus() || w.layout.GetFocusable().HasFocus()
//	return focus
//}

func NewWindow(colors config.Colors, shortcuts *config.Shortcuts, db *storage.Database) *Window {
	w := &Window{
		app:     tview.NewApplication(),
		db:      db,
		layout:  twidgets.NewModalLayout(),
		grid:    tview.NewGrid(),
		project: NewProjects(),
		tags:    NewTags(),
		help:    modals.NewHelp(),
	}

	w.app.SetRoot(w, true)
	w.app.SetInputCapture(w.inputCapture)

	w.bookmarks = NewBookmarkTable(w.openBookmark)
	w.metadata = NewMetadata(w.closeMetadata)

	w.bookmarkForm = modals.NewBookmarkForm(w.createBookmark)
	w.grid.SetBackgroundColor(colors.Background)

	w.gridSize = 6
	w.grid.SetRows(1, -1)
	w.grid.SetColumns(-1)
	w.grid.SetMinSize(2, 2)

	col := colors.NavBar.ToNavBar()

	w.metadata = NewMetadata(w.closeMetadata)
	w.navBar = twidgets.NewNavBar(col, w.navBarClicked)
	navBarLabels = []string{"Help", "New Bookmark", "Menu", "Quit"}

	sc := shortcuts.NavBar
	navBarShortucts = []tcell.Key{sc.Help, sc.NewBookmark, sc.Menu, sc.Quit}

	for i, v := range navBarLabels {
		btn := tview.NewButton(v)
		w.navBar.AddButton(btn, navBarShortucts[i])
	}

	w.grid.AddItem(w.navBar, 0, 0, 1, 1, 1, 10, false)
	w.grid.AddItem(w.layout, 1, 0, 1, 1, 4, 4, true)

	w.layout.Grid().AddItem(w.project, 0, 0, 3, 2, 5, 5, false)
	w.layout.Grid().AddItem(w.tags, 3, 0, 3, 2, 5, 5, false)
	w.layout.Grid().AddItem(w.bookmarks, 0, 2, 6, 4, 10, 10, true)

	w.app.SetFocus(w.bookmarks)

	return w
}

func (w *Window) navBarClicked(label string) {
	logrus.Info("User pressed: ", label)

}

func (w *Window) closeMetadata(save bool, bookmark *models.Bookmark) {
	if !save {
		w.layout.Grid().RemoveItem(w.bookmarks)
		w.layout.Grid().AddItem(w.bookmarks, 0, 2, 6, 4, 10, 10, true)
		w.layout.Grid().AddItem(w.project, 0, 0, 3, 2, 5, 5, false)
		w.layout.Grid().AddItem(w.tags, 3, 0, 3, 2, 5, 5, false)
	}
	w.app.SetFocus(w.lastFocus)
	w.lastFocus = nil
	w.metadataOpen = false
}

func (w *Window) openBookmark(b *models.Bookmark) {
	w.openMetadata()
	w.metadata.setData(b)
}

func (w *Window) openMetadata() {
	w.lastFocus = w.app.GetFocus()
	w.app.SetFocus(w.metadata)

	//w.grid.Blur()
	//w.metadata.Focus(func(p tview.Primitive){})
	w.layout.Grid().RemoveItem(w.bookmarks)
	w.layout.Grid().RemoveItem(w.project)
	w.layout.Grid().RemoveItem(w.tags)

	w.layout.Grid().AddItem(w.bookmarks, 0, 0, 6, 4, 10, 10, false)
	w.layout.Grid().AddItem(w.metadata, 0, 4, 6, 2, 10, 10, true)

	index, _ := w.bookmarks.table.GetSelection()
	bookmark := w.bookmarks.items[index-1]

	w.app.QueueUpdateDraw(func() { w.metadata.setData(bookmark) })
	w.metadataOpen = true
}

func (w *Window) createBookmark(bookmark *models.Bookmark) {
	logrus.Info("Got new bookmark: ", bookmark)

	err := w.db.NewBookmark(bookmark)
	if err != nil {
		logrus.Error("Failed to create bookmark: ", err)
	} else {
		bookmarks, err := w.db.GetAllBookmarks()
		if err != nil {
			return
		}
		w.bookmarks.SetData(bookmarks)
		if w.hasModal {
			w.layout.RemoveModal(w.modal)
			w.app.SetFocus(w.lastFocus)
			w.lastFocus = nil
			w.modal = nil
			w.hasModal = false
		}
	}
}
