package cmd

import (
	"fmt"
	"image-syncer/pkg/web"
	"os"
	"os/signal"
	"syscall"

	"image-syncer/pkg/client"
	"github.com/spf13/cobra"
)

var (
	logPath, configFile, recordPath, defaultRegistry, defaultNamespace string

	procNum, retries int
)

// RootCmd describes "image-syncer" command
var RootCmd = &cobra.Command{
	Use:     "image-syncer",
	Aliases: []string{"image-syncer"},
	Short:   "A docker registry image synchronization tool",
	Long: `A Fast and Flexible docker registry image synchronization tool implement by Go.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// work starts here
		_, err := client.NewSyncClient(configFile, logPath, recordPath, procNum, retries, defaultRegistry, defaultNamespace)
		if err != nil {
			return fmt.Errorf("init sync client error: %v", err)
		}
		//client.Run()
		s := web.HttpServer()
		chs := web.HttpStart(s)
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		for {
			select {
			case <-chs:
				s.Close()
				return fmt.Errorf("server is close: %v", 1)
			case si := <-c:
				fmt.Println("message", fmt.Sprint("get a signal %s", si.String()))
				switch si {
				case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
					s.Close()
					return fmt.Errorf("server is close: %v", 2)
				case syscall.SIGHUP:
				default:
				}
			}
		}
		return nil
	},
}

func init() {
	var defaultLogPath, defaultConfigFile, defaultRecordPath,registry string

	pwd, err := os.Getwd()
	if err == nil {
		defaultLogPath = ""
		defaultConfigFile = pwd + "/" + "image-syncer.json"
		defaultRecordPath = pwd + "/" + "records"
	}
	//registry = "docker.oa.com:8080"
	registry = "localhost:5002"
	RootCmd.PersistentFlags().StringVar(&configFile, "config", defaultConfigFile, "config file path")
	RootCmd.PersistentFlags().StringVar(&logPath, "log", defaultLogPath, "log file path (default in os.Stderr)")
	RootCmd.PersistentFlags().StringVar(&recordPath, "records", defaultRecordPath,
		"records file path, to record the blobs that have been synced, auto generated if not exist")
	RootCmd.PersistentFlags().StringVar(&defaultRegistry, "registry", registry,
		"default destinate registry url when destinate registry is not given in the config file, can also be set with docker.oa.com:8080 environment value")
	RootCmd.PersistentFlags().StringVar(&defaultNamespace, "namespace", os.Getenv("DEFAULT_NAMESPACE"),
		"default destinate namespace when destinate namespace is not given in the config file, can also be set with DEFAULT_NAMESPACE environment value")
	RootCmd.PersistentFlags().IntVarP(&procNum, "proc", "p", 5, "numbers of working goroutines")
	RootCmd.PersistentFlags().IntVarP(&retries, "retries", "r", 2, "times to retry failed task")
}

// Execute executes the RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
