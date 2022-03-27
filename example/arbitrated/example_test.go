package arbitrated

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/btsomogyi/arbiter"
	pb "github.com/btsomogyi/arbiter/example/examplepb"
	"github.com/sirupsen/logrus/hooks/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var (
	randomSeed      = 1234567890
	defaultWorkTime = 10 * time.Microsecond
	bufSize         = 1024 * 1024
	dumpState       = false
)

func Benchmark_randomRequests(b *testing.B) {
	benchmarks := []struct {
		Requests    int
		Concurrency int
	}{
		{100, 1},
		{100, 2},
		{100, 4},
		{100, 8},
		{100, 12},
		{100, 24},
		{500, 1},
		{500, 2},
		{500, 4},
		{500, 8},
		{500, 12},
		{500, 24},
		{500, 48},
		{5000, 1},
		{5000, 2},
		{5000, 4},
		{5000, 8},
		{5000, 12},
		{5000, 24},
		{5000, 48},
		{5000, 96},
		{5000, 144},
		{5000, 192},
	}

	r := rand.New(rand.NewSource(int64(randomSeed)))
	var requests = []*pb.UpdateVersionRequest{}
	var getRequests = []*pb.GetVersionRequest{}

	var maxRequests int
	for _, bm := range benchmarks {
		if bm.Requests > maxRequests {
			maxRequests = bm.Requests
		}
	}

	// Precreate request sequence
	type request struct {
		key   int64
		value int64
	}

	for i := 1; i <= maxRequests; i++ {
		key := r.Intn(maxRequests) + 1
		req := pb.UpdateVersionRequest{
			Key:     &pb.Key{Id: int64(key)},
			Version: &pb.Version{Id: int64(i)}, // Always increasing
		}
		requests = append(requests, &req)

		getReq := pb.GetVersionRequest{
			Key: &pb.Key{Id: int64(key)},
		}
		getRequests = append(getRequests, &getReq)
	}

	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("%d reqs @ %d", bm.Requests, bm.Concurrency), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				logger, hook := test.NewNullLogger()

				// Start GRPC Server
				supervisorOptions := arbiter.SetLogger(logger)
				supervisor, err := arbiter.NewSupervisor(supervisorOptions)
				if err != nil {
					b.Fatal(err)
				}
				go supervisor.Process()
				defer supervisor.Terminate()

				lis := bufconn.Listen(bufSize)

				var opts []grpc.ServerOption
				grpcServer := grpc.NewServer(opts...)

				// UNUSED
				option := SetWorkFunc(func() error {
					time.Sleep(defaultWorkTime)
					return nil
				})
				pb.RegisterVersionerServer(grpcServer, NewVersioner(supervisor, option))
				go grpcServer.Serve(lis)

				// Setup client
				bufDialer := func(context.Context, string) (net.Conn, error) {
					return lis.Dial()
				}
				conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
				if err != nil {
					b.Fatalf("grpc connection failed: %v", err)
				}
				defer conn.Close()
				client := pb.NewVersionerClient(conn)

				// Setup concurrent client requests
				queue := make(chan *pb.UpdateVersionRequest)
				done := make(chan interface{})

				g := new(errgroup.Group)
				// Start work publisher
				g.Go(func() error {
					for _, r := range requests[0:bm.Requests] {
						queue <- r
					}
					close(done)
					return nil
				})

				b.ReportAllocs()
				b.StartTimer()
				// Start work consumers
				for w := 1; w <= bm.Concurrency; w++ {
					g.Go(func() error {
						for {
							select {
							case req := <-queue:
								_, err := client.UpdateVersion(context.Background(), req)

								//b.Logf("Received work response: %e", err)
								if err != nil {
									// Lookup grpc status code, and ignore if expected.
									st, ok := status.FromError(err)
									if !ok {
										return err
									}
									switch st.Code() {
									case codes.InvalidArgument, codes.AlreadyExists, codes.ResourceExhausted:
										continue
									default:
										return err
									}
								}
							case <-done:
								return nil
							}
						}
					})

				}

				if err := g.Wait(); err != nil {
					b.Fatal(err)
				}

				if dumpState && n == 0 {
					var reqdata string
					for x, r := range requests[0:bm.Requests] {
						reqdata = fmt.Sprintf("%s[%d:%d:%d]", reqdata, x, r.Key.Id, r.Version.Id)
					}
					b.Log(reqdata)
					for _, r := range getRequests[0:bm.Requests] {
						resp, err := client.GetVersion(context.Background(), r)
						if err != nil {
							b.Log(err)
						} else {
							b.Logf("key: %d version: %d", resp.Key.Id, resp.Version.Id)
						}
					}
				}

				b.StopTimer()

				grpcServer.GracefulStop()

				for _, entry := range hook.AllEntries() {
					b.Log(entry)
				}
			}
		})
	}
}
