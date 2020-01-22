#!/bin/bash

source env.sh
go build && ./db_diff fixed chaufferjob ./private/config.js
