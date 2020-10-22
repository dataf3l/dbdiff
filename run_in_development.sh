#!/bin/bash

source env.sh
go build && ./db_diff english2 chaufferjob ./private/config.js
#go build && ./db_diff english tutree_jobs ./private/config.js
