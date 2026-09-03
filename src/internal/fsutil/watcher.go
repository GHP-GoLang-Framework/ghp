package fsutil

import (
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
	Op   fsnotify.Op
}

func Watch(path string, events chan<- Event) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	if err := w.Add(path); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}

			events <- Event{
				Path: event.Name,
				Op:   event.Op,
			}

		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}

			return err
		}
	}
}
