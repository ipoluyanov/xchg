package app

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ipoluianov/gomisc/logger"
	binserver "github.com/ipoluianov/xchg/bin_server"
	"github.com/ipoluianov/xchg/config"
	"github.com/ipoluianov/xchg/core"
	"github.com/ipoluianov/xchg/http_server"
	"github.com/kardianos/osext"
	"github.com/kardianos/service"
)

var ServiceName string
var ServiceDisplayName string
var ServiceDescription string
var ServiceRunFunc func() error
var ServiceStopFunc func()

func SetAppPath() {
	exePath, _ := osext.ExecutableFolder()
	err := os.Chdir(exePath)
	if err != nil {
		return
	}
}

func init() {
	SetAppPath()
}

func TryService() bool {
	serviceFlagPtr := flag.Bool("service", false, "Run as service")
	installFlagPtr := flag.Bool("install", false, "Install service")
	uninstallFlagPtr := flag.Bool("uninstall", false, "Uninstall service")
	startFlagPtr := flag.Bool("start", false, "Start service")
	stopFlagPtr := flag.Bool("stop", false, "Stop service")

	flag.Parse()

	if *serviceFlagPtr {
		runService()
		return true
	}

	if *installFlagPtr {
		InstallService()
		return true
	}

	if *uninstallFlagPtr {
		UninstallService()
		return true
	}

	if *startFlagPtr {
		StartService()
		return true
	}

	if *stopFlagPtr {
		StopService()
		return true
	}

	return false
}

func NewSvcConfig() *service.Config {
	var SvcConfig = &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceDisplayName,
		Description: ServiceDescription,
	}
	SvcConfig.Arguments = append(SvcConfig.Arguments, "-service")
	return SvcConfig
}

func InstallService() {
	fmt.Println("Service installing")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Install()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Service installed")
}

func UninstallService() {
	fmt.Println("Service uninstalling")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Uninstall()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Service uninstalled")
}

func StartService() {
	fmt.Println("Service starting")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Service started")
	}
}

func StopService() {
	fmt.Println("Service stopping")
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Stop()
	if err != nil {
		log.Fatal(err)
		return
	} else {
		fmt.Println("Service stopped")
	}
}

func runService() {
	prg := &program{}
	s, err := service.New(prg, NewSvcConfig())
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

type program struct{}

func (p *program) Start(_ service.Service) error {
	return ServiceRunFunc()
}

func (p *program) Stop(_ service.Service) error {
	ServiceStopFunc()
	return nil
}

/////////////////////////////

var c *core.Core
var httpServer *http_server.HttpServer
var binServer *binserver.Server

func Start() error {
	logger.Println("Application Started")
	TuneFDs()

	conf, err := config.LoadFromFile(logger.CurrentExePath() + "/" + "config.json")
	if err != nil {
		logger.Println("configuration error:", err)
		return err
	}

	c = core.NewCore(conf)

	httpServer = http_server.NewHttpServer(conf, c)
	httpServer.Start()

	binServer = binserver.NewServer(c)
	binServer.Start()

	return nil
}

func Stop() {
	_ = httpServer.Stop()
	binServer.Stop()
}

func RunConsole() {
	logger.Println("[app]", "Running as console application")
	err := Start()
	if err != nil {
		logger.Println("error:", err)
		return
	}
	_, _ = fmt.Scanln()
	logger.Println("[app]", "Console application exit")
}

func RunAsServiceF() error {
	return Start()
}

func StopServiceF() {
	Stop()
}
