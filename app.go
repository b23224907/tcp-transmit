package main

import (
	"context"
	"tcp-transmit/service"
)

// App struct
type App struct {
	ctx context.Context
	t   *service.TcpTransmit
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.AppStop()
}

func (a *App) AppStart(targetIp, targetPort, listenIp, listenPort string) (result string) {
	a.t = nil
	t := service.NewTcpTransmit()
	a.t = t
	err := t.Start(targetIp, targetPort, listenIp, listenPort)
	if err != nil {
		return err.Error()
	}
	return "success"
}

func (a *App) AppStop() {
	if a.t != nil {
		a.t.Stop()
		a.t = nil
	}
}
func (a *App) ReadRemoteMsg() string {
	if a.t != nil {
		msg, _ := a.t.ReadRemoteChanMsg()
		if len(msg) > 0 {
			return string(msg)
		} else {
			return ""
		}
	} else {
		return ""
	}
}
func (a *App) ReadClientsMsg() string {
	if a.t != nil {
		msg, _ := a.t.ReadClientChanMsg()
		if len(msg) > 0 {
			return string(msg)
		} else {
			return ""
		}
	} else {
		return ""
	}
}

func (a *App) CheckAppStop() bool {
	if a.t != nil {
		return a.t.IsClose
	} else {
		return true
	}
}

func (a *App) GetClientsList() string {
	if a.t != nil {
		return a.t.GetClientsList()
	} else {
		return ""
	}
}
