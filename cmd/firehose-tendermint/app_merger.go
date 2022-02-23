package main

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	mergerApp "github.com/streamingfast/merger/app/merger"
)

func init() {
	flags := func(cmd *cobra.Command) error {
		cmd.Flags().Duration("merger-time-between-store-lookups", 5*time.Second, "delay between source store polling (should be higher for remote storage)")
		cmd.Flags().String("merger-state-file", "{fh-data-dir}/merger/merger.seen.gob", "Path to file containing last written block number, as well as a map of all 'seen blocks' in the 'max-fixable-fork' range")
		cmd.Flags().String("merger-grpc-listen-addr", MergerServingAddr, "Address to listen for incoming gRPC requests")
		cmd.Flags().Duration("merger-writers-leeway", 10*time.Second, "how long we wait after seeing the upper boundary, to ensure that we get as many blocks as possible in a bundle")
		cmd.Flags().Int("merger-one-block-deletion-threads", 10, "number of parallel threads used to delete one-block-files (more means more stress on your storage backend)")
		cmd.Flags().Int("merger-max-one-block-operations-batch-size", 2000, "max number of 'good' (mergeable) files to look up from storage in one polling operation")
		cmd.Flags().Int("merger-next-exclusive-highest-block-limit", 0, "for next bundle boundary")

		return nil
	}

	initFunc := func(runtime *launcher.Runtime) error {
		sfDataDir := runtime.AbsDataDir

		if err := mkdirStorePathIfLocal(mustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url"))); err != nil {
			return err
		}

		if err := mkdirStorePathIfLocal(mustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url"))); err != nil {
			return err
		}

		return mkdirStorePathIfLocal(mustReplaceDataDir(sfDataDir, viper.GetString("merger-state-file")))
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {
		sfDataDir := runtime.AbsDataDir

		return mergerApp.New(&mergerApp.Config{
			StorageMergedBlocksFilesPath:   mustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url")),
			StorageOneBlockFilesPath:       mustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url")),
			TimeBetweenStoreLookups:        viper.GetDuration("merger-time-between-store-lookups"),
			GRPCListenAddr:                 viper.GetString("merger-grpc-listen-addr"),
			WritersLeewayDuration:          viper.GetDuration("merger-writers-leeway"),
			StateFile:                      mustReplaceDataDir(sfDataDir, viper.GetString("merger-state-file")),
			MaxOneBlockOperationsBatchSize: viper.GetInt("merger-max-one-block-operations-batch-size"),
			OneBlockDeletionThreads:        viper.GetInt("merger-one-block-deletion-threads"),
			NextExclusiveHighestBlockLimit: viper.GetUint64("merger-next-exclusive-highest-block-limit"),
		}), nil
	}

	launcher.RegisterApp(&launcher.AppDef{
		ID:            "merger",
		Title:         "Merger",
		Description:   "Produces merged block files from single-block files",
		MetricsID:     "merger",
		Logger:        launcher.NewLoggingDef("merger.*", nil),
		RegisterFlags: flags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
}
