#!/bin/bash

set -e

if [[ -z "$GITHUB_ACTIONS" ]]; then
	python3 -m venv venv
	source venv/bin/activate
fi

pip3 install -r requirements.txt
python3 -m pytest -n 2
