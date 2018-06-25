// Package zubpub is an rpc-based publishing server that serves zap statistics
// from the zapserver

package zubpub

import (
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	pb "github.com/ltlian/glabs/lab7/proto"
	"github.com/ltlian/glabs/lab7/zlog"
	"google.golang.org/grpc"
)

// If the log has too few entries, the server will wait for this duration before
// trying again
const retryInterval = time.Duration(3) * time.Second

var errIterationTermination = errors.New("Iterator did not terminate correctly")

type pubZerver struct {
	logs zlog.ZapLogger
}

// NewPublisher launches a gRPC publishing server
func NewPublisher(zlogger *zlog.ZapLogger) error {
	grpcServer := grpc.NewServer()
	zubserver := newPubServer(zlogger)
	pb.RegisterSubscriptionServer(grpcServer, zubserver)
	listener, err := net.Listen("tcp", "localhost:11101")
	if err != nil {
		return err
	}

	// TODO Necessary to keep a reference to this server after launching?
	return grpcServer.Serve(listener)
}

func newPubServer(zlogger *zlog.ZapLogger) pb.SubscriptionServer {
	zs := new(pubZerver)
	zs.logs = *zlogger
	return zs
}

// Subscribe is called when the server receives a new request from a client
func (zs *pubZerver) Subscribe(stream pb.Subscription_SubscribeServer) error {
	res := new(pb.NotificationMessage)
	// res.Top10 = make([]*pb.NotificationMessage_Top10, 10)

	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	freq := time.Duration(msg.GetRefreshRate()) * time.Second
	r := msg.GetStatistics()

	fmt.Printf("[ZubServer] Got subscription request with a refresh interval of %v and an enum of '%v'\n", freq, r.String())

	for {
		if zs.logs.Entries() < 1 {
			// TODO error codes as int field?
			res.Status = fmt.Sprintf("2: The server has not yet logged any channels. Retrying in %v", retryInterval)
		} else {
			res.Status = "1"

			res.Top10, err = parseTop10(r, zs.logs)
			if err != nil {
				return err
			}
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}

		time.Sleep(freq)
	}
}

func parseTop10(msg pb.SubscribeMessage_Statistics, zl zlog.ZapLogger) ([]*pb.NotificationMessage_Top10, error) {

	var r int
	listLength := 10

	sortedList := zl.FetchSorted(10)
	statList := *zl.FetchStats()

	if sortedList.Len() < listLength {
		listLength = sortedList.Len()
	}

	top10 := make([]*pb.NotificationMessage_Top10, listLength)

	/* Determine if the logger supports statistics.
	 * This can be (edit: has been) refactored into the logger interface which would eliminate the need for this step,
	 * but simpler loggers need to return nil values to satisfy the interface. */

	switch zl.(type) {
	case *zlog.AdvancedZapLogger, *zlog.ZapsMap, *zlog.Zaps:
		// Assert logger type before trying to fetch statistics
		// Potentially not needed if FetchStats returns well formed non-values for loggers that do not support statistics
		break
	default:
		return nil, &unknownLoggerTypeErr{loggerType: fmt.Sprintf("%T", zl)}
	}

	for _, cv := range sortedList {

		/* Do not include channels with no current viewers or 'OFF' entries
		 * 'OFF' is added as a channel when someone turns off their zapbox (or tv?)
		 * TODO prevent 'OFF' from being logged as a channel */

		if cv.Channel == "OFF" || cv.Viewers < 1 {
			continue
		}

		field := new(pb.NotificationMessage_Top10)

		field.ChannelName = cv.Channel

		/* TODO Clients might want to be able to choose between no durations, simple durations which are summarized by the server (ie.
		 * do not show durations if sample size is too low), or precise durations + sample sizes
		 * field.AvgDuration = fmt.Sprintf("Too few samples (Have %v, but want %v)", statList[cv.Channel].SampleSize, sampleMinimum) */
		if msg == pb.SubscribeMessage_SUMMARY {
			// TODO simplified summary string
		}

		if msg == pb.SubscribeMessage_SAMPLESIZE {
			// TODO currently returns 1?
			field.SampleSize = statList[cv.Channel].SampleSize
		}

		if msg == pb.SubscribeMessage_AVGDURATIONS {
			field.AvgDuration = trimDuration(statList[cv.Channel].AvgDur)
		}

		if msg == pb.SubscribeMessage_VIEWERCOUNT {
			field.Viewcount = uint32(cv.Viewers)
		}

		/* TODO DEBUG /**/
		field.SampleSize = statList[cv.Channel].SampleSize
		field.AvgDuration = trimDuration(statList[cv.Channel].AvgDur)
		field.Viewcount = uint32(cv.Viewers)

		top10[r] = field

		r++
	}

	return top10, nil
}

// trimDuration strips decimals from a duration.String() result to avoid repeating decimals
// p = 9 => 1 second precision
func trimDuration(d time.Duration) string {
	p := 9
	x := time.Duration(math.Pow10(p))

	durtime := time.Duration(
		math.Ceil(
			float64(d/x),
		),
	) * x

	return durtime.String()
}

type unknownLoggerTypeErr struct {
	loggerType string
}

func (e *unknownLoggerTypeErr) Error() string {
	return fmt.Sprintf("Publisher could not determine logger type while building top10 list: %v", e.loggerType)
}
