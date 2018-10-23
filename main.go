package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

var transactionsRegex = regexp.MustCompile(`Transactions:[\t ]*([0-9.]+)`)
var availabilityRegex = regexp.MustCompile(`Availability:[\t ]*([0-9.]+)`)
var elapsedTimeRegex = regexp.MustCompile(`Elapsed time:[\t ]*([0-9.]+)`)
var dataTransferredRegex = regexp.MustCompile(`Data transferred:[\t ]*([0-9.]+)`)
var responseTimeRegex = regexp.MustCompile(`Response time:[\t ]*([0-9.]+)`)
var transactionRateRegex = regexp.MustCompile(`Transaction rate:[\t ]*([0-9.]+)`)
var throughputRegex = regexp.MustCompile(`Throughput:[\t ]*([0-9.]+)`)
var concurrencyRegex = regexp.MustCompile(`Concurrency:[\t ]*([0-9.]+)`)
var successfulTransactionsRegex = regexp.MustCompile(`Successful transactions:[\t ]*([0-9.]+)`)
var failedTransactionsRegex = regexp.MustCompile(`Failed transactions:[\t ]*([0-9.]+)`)
var longestTransactionRegex = regexp.MustCompile(`Longest transaction:[\t ]*([0-9.]+)`)
var shortestTransactionRegex = regexp.MustCompile(`Shortest transaction:[\t ]*([0-9.]+)`)

var transactions []float64
var availability []float64
var elapsedTime []float64
var dataTransferred []float64
var responseTime []float64
var transactionRate []float64
var throughput []float64
var concurrency []float64
var successfulTransactions []float64
var failedTransactions []float64
var longestTransaction []float64
var shortestTransaction []float64

func publicKeyFile(file string) (authMethod ssh.AuthMethod, err error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return
	}

	authMethod = ssh.PublicKeys(key)
	return
}

func theSiege(c *cli.Context) {
	args := c.Args()

	keyPath := "~/.ssh/id_rsa"
	username := "root"
	port := "22"
	var addrs []string
	var siegeParams []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ssh-key":
			i = i + 1

			if i >= len(args) {
				log.Fatal("missing value of SSH host address")
			}

			keyPath = args[i]
		case "--ssh-user":
			i = i + 1

			if i >= len(args) {
				log.Fatal("missing value of SSH username address")
			}

			username = args[i]
		case "--ssh-port":
			i = i + 1

			if i >= len(args) {
				log.Fatal("missing value of SSH username address")
			}

			port = args[i]
		case "--ssh-addr":
			i = i + 1

			if i >= len(args) {
				log.Fatal("missing value of SSH host address")
			}

			addrs = append(addrs, args[i])
		default:
			siegeParams = append(siegeParams, args[i])
		}
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	addrs = append(addrs, strings.Split(strings.TrimSpace(string(data)), "\n")...)

	authMethod, err := publicKeyFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)

		go func(addr string) {
			defer wg.Done()

			conn, err := ssh.Dial("tcp", addr+":"+port, &ssh.ClientConfig{
				User:            username,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Auth:            []ssh.AuthMethod{authMethod},
			})
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			session, err := conn.NewSession()
			if err != nil {
				log.Fatal(err)
			}
			defer session.Close()

			var stdoutBuf bytes.Buffer
			session.Stderr = &stdoutBuf

			session.Run("siege " + strings.Join(siegeParams, " "))

			str := stdoutBuf.String()

			matches := transactionsRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				transactions = append(transactions, v)
			}

			matches = availabilityRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				availability = append(availability, v)
			}

			matches = elapsedTimeRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				elapsedTime = append(elapsedTime, v)
			}

			matches = dataTransferredRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				dataTransferred = append(dataTransferred, v)
			}

			matches = responseTimeRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				responseTime = append(responseTime, v)
			}

			matches = transactionRateRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				transactionRate = append(transactionRate, v)
			}

			matches = throughputRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				throughput = append(throughput, v)
			}

			matches = concurrencyRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				concurrency = append(concurrency, v)
			}

			matches = successfulTransactionsRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				successfulTransactions = append(successfulTransactions, v)
			}

			matches = failedTransactionsRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				failedTransactions = append(failedTransactions, v)
			}

			matches = longestTransactionRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				longestTransaction = append(longestTransaction, v)
			}

			matches = shortestTransactionRegex.FindStringSubmatch(str)
			if len(matches) > 0 {
				v, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Fatal(err)
				}
				shortestTransaction = append(shortestTransaction, v)
			}
		}(addr)
	}
	wg.Wait()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Transactions:\t%11.0f\thits\n", sum(transactions...))
	fmt.Fprintf(w, "Availability:\t%11.2f\t%%\n", avg(availability...))
	fmt.Fprintf(w, "Elapsed time:\t%11.2f\tsecs\n", max(elapsedTime...))
	fmt.Fprintf(w, "Data transferred:\t%11.2f\tMB\n", sum(dataTransferred...))
	fmt.Fprintf(w, "Response time:\t%11.2f\tsecs\n", avg(responseTime...))
	fmt.Fprintf(w, "Transaction rate:\t%11.2f\ttrans/sec\n", sum(transactionRate...))
	fmt.Fprintf(w, "Throughput:\t%11.2f\tMB/sec\n", sum(throughput...))
	fmt.Fprintf(w, "Concurrency:\t%11.2f\n", sum(concurrency...))
	fmt.Fprintf(w, "Successful transactions:\t%11.0f\n", sum(successfulTransactions...))
	fmt.Fprintf(w, "Failed transactions:\t%11.0f\n", sum(failedTransactions...))
	fmt.Fprintf(w, "Longest transaction:\t%11.2f\n", max(longestTransaction...))
	fmt.Fprintf(w, "Shortest transaction:\t%11.2f\n", min(shortestTransaction...))
	w.Flush()
}

func sum(input ...float64) float64 {
	var sum float64

	for _, i := range input {
		sum += i
	}

	return sum
}

func avg(input ...float64) float64 {
	return sum(input...) / float64(len(input))
}

func max(input ...float64) float64 {
	var max float64

	for _, i := range input {
		if i > max {
			max = i
		}
	}

	return max
}

func min(input ...float64) float64 {
	var min float64

	for _, i := range input {
		if i < min {
			min = i
		}
	}

	return min
}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "ssh-key",
		Value: "~/.ssh/id_rsa",
		Usage: "SSH identity file (key) address",
	},
	cli.StringSliceFlag{
		Name:  "ssh-addr",
		Usage: "SSH host address",
	},
	cli.StringSliceFlag{
		Name:  "ssh-user",
		Usage: "SSH username",
	},
	cli.StringSliceFlag{
		Name:  "ssh-port",
		Usage: "SSH port",
	},
}

func main() {
	app := cli.NewApp()

	app.Name = "Candia"

	app.Version = "0.1.0"

	app.Usage = "Send parallel command to multiple sieges and combine their output"
	app.UsageText = "candia [siege parameters]"
	app.ArgsUsage = "[siege parameters]"
	app.Flags = flags
	app.Action = theSiege

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
