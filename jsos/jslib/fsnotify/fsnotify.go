package fsnotify

import (
	"github.com/fsnotify/fsnotify"
	"github.com/grafana/sobek"
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func New() jsvm.Module {
	return new(fsnotifyModule)
}

type fsnotifyModule struct {
	svm jsvm.Engineer
}

func (f *fsnotifyModule) Preload(svm jsvm.Engineer) (string, any, bool) {
	f.svm = svm
	obj := svm.Runtime().NewObject()
	_ = obj.Set("newWatcher", f.newWatcher)
	_ = obj.Set("create", fsnotify.Create)
	_ = obj.Set("remove", fsnotify.Remove)
	_ = obj.Set("rename", fsnotify.Rename)
	_ = obj.Set("write", fsnotify.Write)
	_ = obj.Set("chmod", fsnotify.Chmod)

	return "github.com/fsnotify/fsnotify", obj, true
}

func (f *fsnotifyModule) newWatcher() (sobek.Value, error) {
	svm := f.svm
	rt := svm.Runtime()
	ret := rt.NewObject()

	wch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	wm := &watchModule{svm: svm, wch: wch}
	fid := f.svm.Defer().Append(wm.finalize)
	wm.fid = fid

	select {
	case <-wch.Events:
	case <-wch.Errors:

	}

	return ret, nil
}

type watchModule struct {
	svm jsvm.Engineer
	wch *fsnotify.Watcher
	fid uint64
	//onEvent func(evt fsnotify.Event)
	//onError func(err error)
}

func (wm *watchModule) add(path string) error {
	return wm.wch.Add(path)
}

func (wm *watchModule) remove(path string) error {
	return wm.wch.Remove(path)
}

func (wm *watchModule) watchList() []string {
	return wm.wch.WatchList()
}

func (wm *watchModule) onerror(f func(error)) {

}

func (wm *watchModule) onevent(f func(event *watchEvent)) {
	evt := new(fsnotify.Event)
}

func (wm *watchModule) close() error {
	return wm.wch.Close()
}

func (wm *watchModule) addWithBufferSize(path string, size int) error {
	return wm.wch.AddWith(path, fsnotify.WithBufferSize(size))
}

func (wm *watchModule) finalize() error {
	if fn := wm.svm.Defer().Remove(wm.fid); fn != nil {
		return fn()
	}
	return nil
}

func (wm *watchModule) watch() {
	select {
	case err := <-wm.wch.Errors:
	case evt := <-wm.wch.Events:

	}
}

func (wm *watchModule) notifyEvent(evt *fsnotify.Event) {
}

type watchEvent struct {
	Name string      `json:"name"`
	Op   fsnotify.Op `json:"op"`
}
