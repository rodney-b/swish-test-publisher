package publisher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/minio/minio-go/v7"
	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/time/rate"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/healthcheck"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/kafka"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/logger"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/objstore"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/telemetry"
)

func Run(cp config.ConfigProvider) error {
	log := logger.New("publisher")

	err := healthcheck.Start(cp)
	if err != nil {
		return err
	}

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer ctxCancel()

	minioClient, err := objstore.NewMinioClient(ctx, cp)
	if err != nil {
		return errors.Join(errors.New("error connecting minio client"), err)
	}

	tel, err := telemetry.NewTelemetry(ctx, cp, log)
	if err != nil {
		return errors.Join(err, errors.New("error initializing telemetry"))
	}
	defer tel.Shutdown()

	err = publish(ctx, cp, minioClient, log, tel)
	if err != nil {
		log.Error("error producing to kafka", "error", err.Error())
		return errors.Join(errors.New("error producing to message queue"), err)
	}

	return nil
}

func publish(ctx context.Context, cp config.ConfigProvider, minioClient *minio.Client, log *slog.Logger, tel *telemetry.Telemetry) error {
	obj, err := minioClient.GetObject(ctx, cp.GetDataSourceBucket(), cp.GetDataObjectPath(), minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	zDecoder, err := zstd.NewReader(obj)
	if err != nil {
		return errors.Join(errors.New("error creating new zstd decoder"), err)
	}
	defer zDecoder.Close()

	kafkaClient, err := kafka.NewClient(ctx, cp)
	if err != nil {
		return errors.Join(errors.New("error creating kafka client"), err)
	}
	defer kafkaClient.Close()

	if err := kafkaClient.Ping(ctx); err != nil {
		return errors.Join(errors.New("error pinging kafka client"), err)
	}

	const initBufferCap = 64 * 1024      // 64KiB
	const maxTokenSize = 2 * 1024 * 1024 // 2MiB

	// Using a scanner rather than a reader so enforce a max token size
	scanner := bufio.NewScanner(zDecoder)
	buf := make([]byte, initBufferCap)
	scanner.Buffer(buf, maxTokenSize)

	// continously publish for 30 seconds
	pubCtx, pubCtxCancel := context.WithTimeout(ctx, 30*time.Second)
	defer pubCtxCancel()

	limit := rate.Limit(cp.GetMessageRate())
	burst := 1
	limiter := rate.NewLimiter(limit, burst)

	for {
		select {
		case <-pubCtx.Done():
			return nil
		default:
			for scanner.Scan() {
				lineBytes := scanner.Bytes()
				if len(lineBytes) == 0 {
					continue
				}

				msg := &kgo.Record{
					Topic: cp.GetMessageQueueTopic(),
					Value: lineBytes,
				}

				if err := limiter.Wait(ctx); err != nil {
					log.Debug("rate limited wait cancelled", "error", err.Error())
					break
				}

				results := kafkaClient.ProduceSync(ctx, msg)
				if results.FirstErr() != nil {
					log.Error("failed to publish message to message queue", "error", results.FirstErr().Error())
					continue
				}

				tel.IncrementMessageCounter(ctx, cp)

				log.Info("Producing...", "line", string(lineBytes))
			}

			if err = scanner.Err(); err != nil {
				if errors.Is(err, bufio.ErrTooLong) {
					return errors.Join(fmt.Errorf("data line exceeded scanner max buffer size of %d bytes", maxTokenSize), err)
				} else {
					return errors.Join(errors.New("scanner error while reading data set"), err)
				}
			}

			if pubCtx.Err() != nil {
				log.Debug("pubCtx cancelled")
				return nil
			}

			if _, err := obj.Seek(0, io.SeekStart); err != nil {
				return errors.Join(errors.New("data object seeker error"), err)
			}
			if err := zDecoder.Reset(obj); err != nil {
				return errors.Join(errors.New("zstd decoder reset error"), err)
			}

			buf = make([]byte, initBufferCap)
			scanner = bufio.NewScanner(zDecoder)
			scanner.Buffer(buf, maxTokenSize)
		}
	}
}
