package main

import (
	"github.com/karlockhart/txelec"
	"github.com/sirupsen/logrus"
)

func main() {
	err := txelec.LoadConfiguration()
	if err != nil {
		logrus.Fatal(err)
	}

	a, err := txelec.NewAPI()
	if err != nil {
		logrus.Fatal(err)
	}

	a.Start()

}
