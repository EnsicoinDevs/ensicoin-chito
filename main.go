package main

import (
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

	server := &dns.Server{Addr: ":53", Net: "udp"}
	dns.HandleFunc(".", handleRequest)

	go server.ListenAndServe()

	death.NewDeath(syscall.SIGINT, syscall.SIGTERM).WaitForDeath()

	return server.Shutdown()
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	log.Info("salut")

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Compress = false

	switch r.Question[0].Qtype {
	case dns.TypeA:
		log.Info("hoho")

		msg.Authoritative = true

		for _, address := range validAddresses {
			if address.To4() != nil {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: viper.GetString("domain"), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP(address.String()),
				})

				log.Info("woaw")
			}
		}
	case dns.TypeAAAA:
		log.Info("huhu")

		msg.Authoritative = true

		for _, address := range validAddresses {
			if address.To4() == nil {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: viper.GetString("domain"), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP(address.String()),
				})
			}
		}
	}

	w.WriteMsg(msg)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
