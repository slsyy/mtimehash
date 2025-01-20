package main

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"runtime/pprof"

	"github.com/slsyy/mtimehash"
	"github.com/urfave/cli/v2"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	maxUnixTimeFlag = "max-unix-time"
	verboseFlag     = "verbose"
	cpuProfileFlag  = "cpu-profile-path"

	defaultMaxUnitTime = 1704067200 // read flag desc for more info
)

func main() {
	os.Exit(run())
}

func run() int {
	app := &cli.App{
		Name:  "mtimehash",
		Usage: "Set file modification times based on the hash of the file content",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:        maxUnixTimeFlag,
				Usage:       "modulo limit for mtime",
				Value:       defaultMaxUnitTime,
				DefaultText: fmt.Sprintf("%d, which is a reasonable limit as it is a date from the past (begining of the year 2024)", defaultMaxUnitTime),
			},
			&cli.BoolFlag{
				Name:    verboseFlag,
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
			},
			&cli.PathFlag{
				Name:  cpuProfileFlag,
				Usage: "Path to CPU profile output file",
			},
		},
		Action: mainApp,
	}

	if err := app.Run(os.Args); err != nil {
		return 1
	}
	return 0
}

// Only for tests
// nolint:gochecknoglobals
var silentLogs = false

// mainApp handles the CLI logic
func mainApp(c *cli.Context) error {
	maxUnixTime := c.Int64(maxUnixTimeFlag)
	verbose := c.Bool(verboseFlag)
	cpuProfile := c.Path(cpuProfileFlag)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: func() slog.Level {
			if verbose {
				return slog.LevelDebug
			}
			return slog.LevelWarn
		}(),
	}))
	if silentLogs {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: func() slog.Level {
				return 10000
			}(),
		}))
	}
	slog.SetDefault(logger)

	_, _ = maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		logger.Debug(fmt.Sprintf(s, i...))
	}))

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			logger.Error("failed to create CPU profile file", "err", err)
			return err
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Error("failed to start CPU profiling", "err", err)
			return err
		}
		defer pprof.StopCPUProfile()
	}

	if err := mtimehash.Process(streamLines(os.Stdin), maxUnixTime); err != nil {
		logger.Error("failed to process files", "err", err)
		return err
	}
	return nil
}

func streamLines(input io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				return // Stop iteration if yield returns false
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Default().Error("failed to read input", "err", err)
		}
	}
}
