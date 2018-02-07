package main

import (
	"ethereum-front/front"
	"flag"
	"fmt"
	"github.com/spf13/viper"
)

var config_dir string

func main() {
	flag.StringVar(
		&config_dir,
		"config",
		"/app/confdir/",
		`example: -config=/app/confdir/`,
	)
	flag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(config_dir)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		s := fmt.Sprintf("Fatal error config file: %s \n", err)
		panic(s)
	}
	connect := viper.GetString("connect_url")
	fmt.Printf("connect url: %s\n", connect)
	sol_path := viper.GetString("sol_path")
	fmt.Printf("sol files path: %s\n", sol_path)
	keystore_path := viper.GetString("keystore_path")
	fmt.Printf("keystore path: %s\n", keystore_path)
	port := viper.GetInt("port")
	fmt.Printf("port: %d\n", port)
	gaslimit := viper.GetInt64("gaslimit")
	fmt.Printf("gas limit: %d\n", gaslimit)
	solc := viper.GetString("solc")
	fmt.Printf("solc file: %s\n", solc)

	front.Start(connect, sol_path, keystore_path, port, gaslimit, solc)
}
