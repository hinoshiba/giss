package main

import (
	"github.com/hinoshiba/goctx"
	"github.com/hinoshiba/termwindow"
	"giss/values"
)

func gissterm() error {
	own := goctx.NewOwner()

	if err := termwindow.Init(); err != nil {
		return err
	}
	go termwindow.Input(own.NewWorker())
	go termwindow.Start(own.NewWorker())
	defer termwindow.Close()

	termwindow.SetTitle(values.VersionText)
	termwindow.SetMsg("test init")

	wk := own.NewWorker()
	go func(){
		for {
			select {
			case <-wk.RecvCancel():
				wk.Done()
				return
			case  <-termwindow.Key:
			default:
			}
		}
	}()
	own.Wait()
	return nil
}
