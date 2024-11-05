package krools

type ActionStack struct {
	actions []Action
}

func NewActionStack(actions ...Action) *ActionStack {
	return &ActionStack{actions: actions}
}

func (s *ActionStack) Then(ctx Context) error {
	for _, action := range s.actions {
		if err := action.Then(ctx); err != nil {
			return err
		}
	}

	return nil
}
