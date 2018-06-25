// +build !solution

// Zap Collection Server

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"../zlog"
	"../zubclient"
	"../zubpub"

	zap "github.com/ltlian/glabs/lab7"
)

var (
	maddr      = flag.String("mcast", "224.0.1.130:10000", "multicast ip:port")
	labnum     = flag.String("lab", "c2", "which lab exercise to run")
	showHelp   = flag.Bool("h", false, "show this help message and exit")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
	printTime  = flag.Bool("time", false, "log execution times to console")
	ztore      zlog.ZapLogger
)

// The buffer size to be used for reading from the event listener
const bufSize = 4096

func runLab() error {

	listener, err := startServer()
	if err != nil {
		return err
	}

	go readFromServer(listener)

	// Create logger
	switch *labnum {
	case "a", "c1", "c2", "d", "e":
		ztore = zlog.NewSimpleZapLogger()
	case "f":
		ztore = zlog.NewViewersZapLogger()
	case "grpc":
		ztore = zlog.NewAdvancedZapLogger()
	}

	// Toggle whether the logger should print the time taken to fetch and
	// process various data
	switch *labnum {
	case "c2", "d":
		*printTime = true
	}

	// TODO more readable task selection and manual selection of attributes
	zlog.PrintTimes = *printTime

	log.Printf("Logger:\t\t%s", ztore)
	log.Printf("Time measurements:\t%v\n\n", zlog.PrintTimes)

	// Select task
	switch *labnum {
	case "c1", "c2", "d":
		go showViewers("NRK1")
	case "e", "f":
		go showTopTen()
	case "grpc":
		go zubpub.NewPublisher(&ztore)

		client, err := zubclient.NewZubClient()
		if err != nil {
			return err
		}

		//clientRequest := &zubclient.ZubRequest{}
		clientRequest := &zubclient.ZubRequest{
			Refreshinterval: 2,
			Statistic:       3,
		}

		err = client.RequestSub(clientRequest)
		if err != nil {
			return err
		}

		go client.Listen()
	}

	// Join with previous switch to show viewers for NRK1 and TV2 Norge
	if *labnum == "c2" {
		go showViewers("TV2 Norge")
	}

	return nil
}

func startServer() (*net.UDPConn, error) {

	addr, err := net.ResolveUDPAddr("udp", *maddr)
	if err != nil {
		return nil, err
	}

	server, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	log.Printf("ZapServer listening for multicast address %v", addr)

	return server, nil
}

// showViewers() shows the amount of viewers for a given channel every second
// TODO error handling
func showViewers(chName string) {
	for {
		viewers := ztore.Viewers(chName)
		fmt.Println(fmt.Sprintf("%-18s%d viewers", chName+":", viewers))
		time.Sleep(time.Second)
	}
}

// readFromServer() launches the client which listens for, parses, and logs zap
// events
func readFromServer(listener *net.UDPConn) {

	b := make([]byte, bufSize)

	for {
		// Listen to UDP address and write result to b
		n, addr, err := listener.ReadFromUDP(b)
		if err != nil {
			panic(err)
		}

		res := string(b[:n])

		// Dump to console and skip logging
		if *labnum == "a" {
			log.Printf("Received from %v: %v", addr, res)
			continue
		}

		zCh, ztat, err := zap.NewSTBEvent(res)
		if err != nil {
			panic(err)
		} else if zCh != nil {
			ztore.LogZap(*zCh)
		} else if ztat != nil {
			// Statuschange is ignored, however the HDMI status can be used for channel view statistics
			// ztore.LogZap(*ztat)
		} else {
			panic(fmt.Errorf("Nothing to handle from NewSTBEvent response"))
		}
	}
}

// showTopTen() prints the top 10 channels in terms of viewer count every 5
// seconds
func showTopTen() {
	retryInterval := time.Second * 5

	for {
		listLength := ztore.Entries()

		if listLength < 2 {
			fmt.Printf("Not enough entries in log to build top10 (have %v). Retrying in %v\n", listLength, retryInterval)
			time.Sleep(retryInterval)
			continue
		}

		sortedChannels := ztore.FetchSorted(10)
		println("len: ", len(sortedChannels))
		fmt.Printf("\n    Channel\t     Viewers\n")
		/* TODO
		for as := sortedChannels.ChanViewersList [
			println(as)
		]*/

		for r, ch := range sortedChannels {
			fmt.Printf("%2v: %-18v%3v\n", r+1, ch.Channel, ch.Viewers)
			if r > 8 { // Break after rank (9+1) or higher
				break
			}
		}

		time.Sleep(retryInterval)
	}
}
