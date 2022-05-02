package _exec

import "github.com/mitchellh/go-homedir"

func HomeDir() (string, error) {
	return homedir.Dir()
}
