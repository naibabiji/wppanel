package main

import "embed"

//go:embed templates/*
var TemplatesFS embed.FS

//go:embed static/css/* static/js/*
var StaticFS embed.FS

//go:embed wp-panel-cache-helper/*
var PluginFS embed.FS
