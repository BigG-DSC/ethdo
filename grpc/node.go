// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// FetchGenesisTime fetches the genesis time.
func FetchGenesisTime(conn *grpc.ClientConn) (time.Time, error) {
	if conn == nil {
		return time.Now(), errors.New("no connection to beacon node")
	}
	client := ethpb.NewNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	res, err := client.GetGenesis(ctx, &types.Empty{})
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(res.GetGenesisTime().Seconds, 0), nil
}

// FetchGenesisValidatorsRoot fetches the genesis validators root.
func FetchGenesisValidatorsRoot(conn *grpc.ClientConn) ([]byte, error) {
	if conn == nil {
		return nil, errors.New("no connection to beacon node")
	}
	client := ethpb.NewNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	res, err := client.GetGenesis(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}
	return res.GetGenesisValidatorsRoot(), nil
}

// FetchVersion fetches the version and metadata from the server.
func FetchVersion(conn *grpc.ClientConn) (string, string, error) {
	if conn == nil {
		return "", "", errors.New("no connection to beacon node")
	}
	client := ethpb.NewNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	version, err := client.GetVersion(ctx, &types.Empty{})
	if err != nil {
		return "", "", err
	}
	return version.Version, version.Metadata, nil
}

// FetchSyncing returns true if the node is syncing, otherwise false.
func FetchSyncing(conn *grpc.ClientConn) (bool, error) {
	if conn == nil {
		return false, errors.New("no connection to beacon node")
	}
	client := ethpb.NewNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	syncStatus, err := client.GetSyncStatus(ctx, &types.Empty{})
	if err != nil {
		return false, err
	}
	return syncStatus.Syncing, nil
}
