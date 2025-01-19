package authHelper

import "context"

func isUserActive(ctx context.Context, userId string) (bool, error) {
	ctx, span := tracer.Start(ctx, "isUserActive")
	defer span.End()

	// ToDo: Implement the function to check if the user exists and the status is active (pay attention to the admin user for backendSettings endpoints)

	return true, nil
}
