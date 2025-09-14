package config

var Protocol = "tcp"
var Port = ":3000"
var MaxConnection = 20000
var MaxKeyNumber int = 10
var EvictionRatio = 0.1

var EvictionPolicy string = "allkeys-random"
