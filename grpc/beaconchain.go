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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// FetchChainConfig fetches the chain configuration from the beacon node.
// It tweaks the output to make it easier to work with by setting appropriate
// types.
func FetchChainConfig(conn *grpc.ClientConn) (map[string]interface{}, error) {
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()
	config, err := beaconClient.GetBeaconConfig(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	results := make(map[string]interface{})
	for k, v := range config.Config {
		// Handle integers
		if v == "0" {
			results[k] = uint64(0)
			continue
		}
		intVal, err := strconv.ParseUint(v, 10, 64)
		if err == nil && intVal != 0 {
			results[k] = intVal
			continue
		}

		// Handle byte arrays
		if strings.HasPrefix(v, "[") {
			vals := strings.Split(v[1:len(v)-1], " ")
			res := make([]byte, len(vals))
			for i, val := range vals {
				intVal, err := strconv.Atoi(val)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert value %q for %s", v, k)
				}
				res[i] = byte(intVal)
			}
			results[k] = res
			continue
		}

		// String (or unhandled format)
		results[k] = v
	}
	return results, nil
}

// FetchValidator fetches the validator definition from the beacon node.
func FetchValidator(conn *grpc.ClientConn, account wtypes.Account) (*ethpb.Validator, error) {
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	req := &ethpb.GetValidatorRequest{
		QueryFilter: &ethpb.GetValidatorRequest_PublicKey{
			PublicKey: account.PublicKey().Marshal(),
		},
	}
	return beaconClient.GetValidator(ctx, req)
}

// FetchValidatorInfo fetches current validator info from the beacon node.
func FetchValidatorInfo(conn *grpc.ClientConn, account wtypes.Account) (*ethpb.ValidatorInfo, error) {
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
	defer cancel()

	stream, err := beaconClient.StreamValidatorsInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to contact beacon node")
	}

	changeSet := &ethpb.ValidatorChangeSet{
		Action:     ethpb.SetAction_SET_VALIDATOR_KEYS,
		PublicKeys: [][]byte{account.PublicKey().Marshal()},
	}
	err = stream.Send(changeSet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send validator public key")
	}
	return stream.Recv()
}
