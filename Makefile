# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
 @go run ./cmd/api -jwt-secret=${JWT_SECRET}