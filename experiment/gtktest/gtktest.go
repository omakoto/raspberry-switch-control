package main

// See: https://github.com/BurntSushi/xgbutil/blob/master/keybind/doc.go
import (
	"fmt"
	_ "github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/omakoto/go-common/src/common"
)

func main() {
	xu, err := xgbutil.NewConn()
	common.Checke(err)

	keybind.Initialize(xu)

	// This isn't quite "press/release" because key repeats trigger them.
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			fmt.Printf("Pressed: %v\n", ev)
		}).Connect(xu, xu.RootWin(), "a", true)

	keybind.KeyReleaseFun(func(xu *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
		fmt.Printf("Released: %v\n", ev)
	}).Connect(xu, xu.RootWin(), "a", true)

	//xevent.KeyPressFun(
	//	func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
	//		// keybind.LookupString does the magic of implementing parts of
	//		// the X Keyboard Encoding to determine an english representation
	//		// of the modifiers/keycode tuple.
	//		// N.B. It's working for me, but probably isn't 100% correct in
	//		// all environments yet.
	//		modStr := keybind.ModifierString(e.State)
	//		keyStr := keybind.LookupString(X, e.State, e.Detail)
	//		if len(modStr) > 0 {
	//			fmt.Printf("Key: %s-%s\n", modStr, keyStr)
	//		} else {
	//			fmt.Println("Key:", keyStr)
	//		}
	//
	//		if keybind.KeyMatch(X, "Escape", e.State, e.Detail) {
	//			if e.State&xproto.ModMaskControl > 0 {
	//				fmt.Println("Control-Escape detected. Quitting...")
	//				xevent.Quit(X)
	//			}
	//		}
	//	}).Connect(xu, xu.RootWin())

	xevent.Main(xu)
}
