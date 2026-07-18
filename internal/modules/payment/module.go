package payment

// Module is the Payment bounded context entrypoint (application wiring).
// Echo routes: internal/delivery/http only.
type Module struct{}

func NewModule() *Module { return &Module{} }

func (m *Module) Name() string { return "payment" }
