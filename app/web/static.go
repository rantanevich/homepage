package web

import "embed"

//go:embed static/* templates/*
var WebFS embed.FS
