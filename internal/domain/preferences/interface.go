package preferences

// Interface groups the preferences that shape the TUI itself, independent of
// audio or color.
type Interface struct {
	// SidebarHidden is the sidebar's state at startup. Ctrl-B still toggles it
	// during the session; that toggle is not persisted on its own.
	SidebarHidden bool
	// Animations gates the small motion in the UI, such as the playing-stream
	// blink.
	Animations bool
}

func defaultInterface() Interface {
	return Interface{SidebarHidden: false, Animations: true}
}
