// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/SlinkyProject/slurm-client/pkg/client"
	"github.com/SlinkyProject/slurm-client/pkg/object"
	"github.com/SlinkyProject/slurm-client/pkg/types"
)

// Initialize the slurm client to talk to slurmrestd.
// Requires that the env SLURM_JWT is set.
func NewSlurmClient(server string, cacheFreq time.Duration) (client.Client, error) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	token, ok := os.LookupEnv("SLURM_JWT")
	if !ok || token == "" {
		return nil, errors.New("SLURM_JWT must be defined and not empty")
	}

	if cacheFreq <= 1*time.Second {
		return nil, errors.New("cache-freq >= 1s")
	}

	// Create slurm client
	config := &client.Config{
		Server:    server,
		AuthToken: token,
	}

	// Instruct the client to keep a cache of slurm objects
	clientOptions := client.ClientOptions{
		EnableFor: []object.Object{
			&types.V0043JobInfo{},
			&types.V0043Node{},
			&types.V0043PartitionInfo{},
		},
		CacheSyncPeriod: cacheFreq,
	}
	slurmClient, err := client.NewClient(config, &clientOptions)
	if err != nil {
		return nil, err
	}

	// Start client cache
	go slurmClient.Start(ctx)

	logger.Info("Created slurm client")

	return slurmClient, nil
}
