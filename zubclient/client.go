// Package zubclient is an rpc-client that receives and prints the results from a publishing server.

package zubclient

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	pb "../proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// TODO define a newMessage() function for creating new message structs to server?

// dialAddress is the address that the client will dial to contact the grpc publishing server
var dialAddress = "localhost:11101"
var errEmptyArrayResponse = errors.New("Empty response from server")

// ZubClient is a grpc client that receives a list of top 10 channels from a publishing server
type ZubClient struct {
	stream pb.Subscription_SubscribeClient
	c      pb.SubscriptionClient
}

// ZubRequest asdf
type ZubRequest struct {
	Refreshinterval uint32
	Statistic       uint8
}

// NewZubClient returns a grpc subscription client
func NewZubClient() (*ZubClient, error) {
	var client ZubClient

	conn, err := grpc.Dial(dialAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client.c = pb.NewSubscriptionClient(conn)
	s, err := client.c.Subscribe(context.Background())
	if err != nil {
		return nil, err
	}

	client.stream = s
	return &client, nil
}

// RequestSub will request a subscription from a grpc publishing server with the given update frequency
func (zc ZubClient) RequestSub(r *ZubRequest) error {
	req := &pb.SubscribeMessage{
		RefreshRate: r.Refreshinterval,
		Statistics:  pb.SubscribeMessage_Statistics(r.Statistic),
	}

	return zc.stream.Send(req)
}

func (zc *ZubClient) dumpTop10(r *pb.NotificationMessage) error {

	if len(r.GetTop10()) < 1 {
		return errEmptyArrayResponse
	}

	/*
		// TODO handle different types of results better
		thelist := r.GetTop10()

		thelist[0].GetAvgDuration()
		thelist[0].GetChannelName()
		thelist[0].GetSampleSize()
		thelist[0].GetViewcount()
	*/

	fmt.Printf("\n    Channel\t  Viewers     AvgDur   SampSize\n")
	for i, ch := range r.GetTop10() {
		fmt.Printf(
			"%2v: %-18v%3v\t%v\t%v\n",
			i+1, ch.GetChannelName(),
			ch.GetViewcount(),
			ch.GetAvgDuration(),
			ch.GetSampleSize()
		)
	}

	return nil
}

// Listen() starts the client's listen loop and prints a top10 list
func (zc *ZubClient) Listen() error {
	retryDur := time.Duration(1) * time.Second

	for {
		r, err := zc.stream.Recv()
		switch err {
		case nil:
			break
		case io.EOF:
			fmt.Printf("[ZubClient] got EOF; retrying in %v\n", retryDur)
			time.Sleep(retryDur)
			continue
		default:
			return &unknownError{err: err}
		}

		// TODO cleaner way to do this; Restructure or separate into functions
		rStatus := r.GetStatus()

		if rStatus[0:1] != "1" {
			fmt.Printf("[ZubClient] Error response from server: [%v]\n[ZubClient] Reconnecting in %v\n", rStatus, retryDur)
			continue
		}

		err = zc.dumpTop10(r)
		switch err {
		case nil:
			break
		case errEmptyArrayResponse:
			fmt.Println(err.Error())
		default:
			return err
		}
	}
}

type unknownError struct {
	err error
}

func (e *unknownError) Error() string {
	return fmt.Sprintf("ZubClient encountered an unknown error while reading from the publishing server: %v", e.err)
}

type tooFewFieldsError struct {
	input []string
}

func (e *tooFewFieldsError) Error() string {
	csv := strings.Join(e.input, ", ")
	return fmt.Sprintf("top10 result from publishing server has too few fields.\n'%v'\n(Need 3, got %v)", csv, len(e.input))
}
