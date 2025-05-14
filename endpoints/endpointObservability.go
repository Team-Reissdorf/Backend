package endpoints

import (
	"github.com/LucaSchmitz2003/FlowWatch"
	"go.opentelemetry.io/otel"
)

var (
	Tracer = otel.Tracer("EndpointTracer")
	Logger = FlowWatch.GetLogHelper()
)
