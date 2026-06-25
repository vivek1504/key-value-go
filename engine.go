package main

import (
	"errors"
)

type Engine struct {
	data map[string]string
}

func NewEngine() *Engine {
	return &Engine{
		data: make(map[string]string),
	}
}

func (e *Engine) Set(key, value string) error {
	e.data[key] = value
	return nil
}

func (e *Engine) Get(key string) (string, error) {
	value, ok := e.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return value, nil
}
