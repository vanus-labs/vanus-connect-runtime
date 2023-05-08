package controller

type ConnectorOption func(opt *connectorOptions)

type connectorOptions struct {
	labelSelector string
	handler       ConnectorEventHandler
}

func newConnectorOptions(options ...ConnectorOption) connectorOptions {
	opts := defaultConnectorOptions()

	for _, apply := range options {
		apply(&opts)
	}
	return opts
}

func defaultConnectorOptions() connectorOptions {
	return connectorOptions{}
}

func WithFilter(filter string) ConnectorOption {
	return func(opt *connectorOptions) {
		opt.labelSelector = filter
	}
}

func WithEventHandler(handler ConnectorEventHandler) ConnectorOption {
	return func(opt *connectorOptions) {
		opt.handler = handler
	}
}
