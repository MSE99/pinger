package main

type config struct {
	apps []appDef
}

type appDef struct {
	appName   string
	statusURL string
	onError   errorHandlingDef
}

type errorHandlingDef struct {
	alertURL string
	body     any
}
