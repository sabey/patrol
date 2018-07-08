package patrol

type TriggerApp struct {
	Start func(
		id string,
		app *PatrolApp,
	)
	Started func(
		id string,
		app *PatrolApp,
	)
	StartFailed func(
		id string,
		app *PatrolApp,
	)
	Running func(
		id string,
		app *PatrolApp,
	)
}

func (self *TriggerApp) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
