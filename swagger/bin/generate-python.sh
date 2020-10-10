#!/bin/sh

set -e

docker run --rm \
	-v ${PWD}:/local openapitools/openapi-generator-cli generate \
	-i /local/swagger/swagger.yaml \
	-g python \
	-o /local/sdktest/python/gnomock \
	-c /local/swagger/config/python.yaml \
	--git-user-id orlangure \
	--git-repo-id gnomock-python-sdk
