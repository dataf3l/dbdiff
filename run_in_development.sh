#!/bin/bash
go build && ./db_diff deliveryjobsnyc chaufferjob ./private/config.js
