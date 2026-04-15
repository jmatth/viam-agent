package main

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/getlantern/systray"
	"github.com/samber/mo"
)

//go:embed viam.ico
var viamIco []byte

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	var traywg sync.WaitGroup

	systray.Run(onReady(ctx, cancel, &traywg), onExit)
	traywg.Wait()
}

func onReady(ctx context.Context, cancel context.CancelCauseFunc, traywg *sync.WaitGroup) func() {
	return func() {
		systray.SetIcon(viamIco)

		adminMode := systray.AddMenuItem("Enter admin mode", "TODO")
		traywg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-adminMode.ClickedCh:
					triggerAdmin()
				}
			}
		})

		exit := systray.AddMenuItem("Exit", "TODO")
		traywg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-exit.ClickedCh:
					slog.Info("Exit pressed, shutting down")
					cancel(errors.New("Exit selected from tray"))
					systray.Quit()
				}
			}
		})
	}
}

func onExit() {}

func triggerAdmin() {
	const ipcPath = `C:\opt\viam\tmp\viam-agent-win.sock`
	dialRes := mo.TupleToResult(net.Dial("unix", ipcPath))
	if dialRes.IsError() {
		return
	}
	conn := dialRes.MustGet()
	defer conn.Close()
	conn.Write([]byte("ADMIN"))
	conn.Close()
}
