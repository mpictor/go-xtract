// +build !release

package xlate

// TestingClearSetupCheck clears loaded bool so that we can load multiple datasets in a testing binary
func TestingClearSetupCheck() {
	loaded = false
}
