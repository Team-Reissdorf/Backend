package endpoints

import (
	"github.com/LucaSchmitz2003/FlowWatch"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("EndpointTracer")
	logger = FlowWatch.GetLogHelper()
)
