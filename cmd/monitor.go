/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ztrade/ctpmonitor/config"
	"github.com/ztrade/ctpmonitor/monitor"
	"github.com/ztrade/ctpmonitor/pb"
	"github.com/ztrade/ctpmonitor/service"
)

var (
	logFile string
	cfgFile string
)

// monitorCmd represents the monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "monitor and record ctp market data",
	Long:  `monitor and record ctp market data"`,
	// Run:   runMonitor,
	Run: runService,
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	monitorCmd.PersistentFlags().StringVarP(&logFile, "log", "l", "monitor.log", "log file")
	monitorCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "config file")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// monitorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		viper.AddConfigPath(filepath.Join(exPath, "configs"))
		viper.SetConfigName("monitor")
	}
	if err := viper.ReadInConfig(); err == nil {
		logrus.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runService(cmd *cobra.Command, args []string) {
	initConfig()
	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}

	s, err := service.NewCtpService(&cfg)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	httpSrv := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			recovery.Recovery(),
		),
	)
	grpcSrv := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			recovery.Recovery(),
		),
	)
	pb.RegisterCtpHTTPServer(httpSrv, s)
	pb.RegisterCtpServer(grpcSrv, s)
	app := kratos.New(
		kratos.Name(""),
		kratos.Server(
			httpSrv,
			grpcSrv,
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func runMonitor(cmd *cobra.Command, args []string) {
	initConfig()
	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	m := monitor.NewCTPMonitor(&cfg)
	err = m.Start()
	if err != nil {
		fmt.Println("start error:", err.Error())
		return
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("begin to stop...")
	err = m.Stop()
	fmt.Println("stop...", err)
	if err != nil {
		fmt.Println("monitor stop error:", err.Error())
		return
	}
}
