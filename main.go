package main

import (
	"fmt"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vrecan/death"
	"net"
	"os"
	"syscall"
)

func init() {
	pflag.StringSliceP("address", "a", []string{}, "add an address")
	pflag.StringP("domain", "d", "", "domain name")
	pflag.IntP("port", "p", 53, "port")

	viper.BindPFlags(pflag.CommandLine)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
}

var rootCmd = &cobra.Command{
	Use:   "ensicoin-chito",
	Short: "DNS Bootstraping server for Ensicoin",
	Run: func(cmd *cobra.Command, args []string) {
		if err := launch(); err != nil {
			os.Exit(1)
		}
	},
}

var validAddresses []net.IP

func launch() error {
	log.Info("ensicoin-chito 1.0.0")

	addresses := viper.GetStringSlice("address")

	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip != nil {
			log.WithField("address", address).Info("advertizing")
			validAddresses = append(validAddresses, ip)
		} else {
			log.WithField("address", address).Warn("invalid address")
		}
	}

	server := &dns.Server{Addr: fmt.Sprintf(":%d", viper.GetInt("port")), Net: "udp"}
	dns.HandleFunc(".", handleRequest)

	go server.ListenAndServe()

	death.NewDeath(syscall.SIGINT, syscall.SIGTERM).WaitForDeath()

	return server.Shutdown()
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	log.Debug("handling request")

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func parseQuery(m *dns.Msg) {
	log.Debug("parsing query")

	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Debug("A")

			for _, address := range validAddresses {
				if address.To4() != nil {
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", viper.GetString("domain"), address.String()))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		case dns.TypeAAAA:
			log.Debug("AAAA")

			for _, address := range validAddresses {
				if address.To4() == nil {
					rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", viper.GetString("domain"), address.String()))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			}
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
