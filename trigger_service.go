package patrol

type TriggerService struct {
	Start func(
		id string,
		app *PatrolService,
	)
	Started func(
		id string,
		app *PatrolService,
	)
	StartFailed func(
		id string,
		app *PatrolService,
	)
	Running func(
		id string,
		app *PatrolService,
	)
}

func (self *TriggerService) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
