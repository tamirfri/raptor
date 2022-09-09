/*
Copyright (c) 2022 RaptorML authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	manifests "github.com/raptor-ml/raptor/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FeatureSetBuilder = "featureset"
	ExpressionBuilder = "expression"
)

// DataConnector is a parsed abstracted representation of a manifests.DataConnector
type DataConnector struct {
	FQN    string                 `json:"fqn"`
	Kind   string                 `json:"kind"`
	Config manifests.ParsedConfig `json:"config"`
}

// DataConnectorFromManifest returns a DataConnector from a manifests.DataConnector
func DataConnectorFromManifest(ctx context.Context, dc *manifests.DataConnector, r client.Reader) (DataConnector, error) {
	pc, err := dc.ParseConfig(ctx, r)
	if err != nil {
		return DataConnector{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return DataConnector{
		FQN:    dc.FQN(),
		Kind:   dc.Spec.Kind,
		Config: pc,
	}, nil
}

// Metadata is the metadata of a feature.
type Metadata struct {
	FQN           string        `json:"FQN"`
	Primitive     PrimitiveType `json:"primitive"`
	Aggr          []WindowFn    `json:"aggr"`
	Freshness     time.Duration `json:"freshness"`
	Staleness     time.Duration `json:"staleness"`
	Timeout       time.Duration `json:"timeout"`
	Builder       string        `json:"builder"`
	DataConnector string        `json:"connector"`
}

// ValidWindow checks if the feature have aggregation enabled, and if it is valid
func (md Metadata) ValidWindow() bool {
	if md.Freshness < 1 {
		return false
	}
	if md.Staleness < md.Freshness {
		return false
	}
	if len(md.Aggr) == 0 {
		return false
	}
	if !(md.Primitive == PrimitiveTypeInteger || md.Primitive == PrimitiveTypeFloat) {
		return false
	}
	return true
}

func aggrsToStrings(a []manifests.AggrType) []string {
	res := make([]string, 0, len(a))
	for _, v := range a {
		res = append(res, string(v))
	}
	return res
}

// MetadataFromManifest returns a Metadata from a manifests.Feature
func MetadataFromManifest(in *manifests.Feature) (*Metadata, error) {
	primitive := StringToPrimitiveType(in.Spec.Primitive)
	if primitive == PrimitiveTypeUnknown {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPrimitiveError, in.Spec.Primitive)
	}
	aggr, err := StringsToWindowFns(aggrsToStrings(in.Spec.Builder.Aggr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregation functions: %w", err)
	}
	if len(aggr) > 0 && primitive != PrimitiveTypeInteger && primitive != PrimitiveTypeFloat {
		return nil, fmt.Errorf("%w with Aggregation: %s", ErrUnsupportedPrimitiveError, in.Spec.Primitive)
	}
	if in.Spec.Builder.AggrGranularity.Milliseconds() > 0 && len(aggr) > 0 {
		in.Spec.Freshness = in.Spec.Builder.AggrGranularity
	}

	md := &Metadata{
		FQN:       in.FQN(),
		Primitive: primitive,
		Aggr:      aggr,
		Freshness: in.Spec.Freshness.Duration,
		Staleness: in.Spec.Staleness.Duration,
		Timeout:   in.Spec.Timeout.Duration,
		Builder:   strings.ToLower(in.Spec.Builder.Kind),
	}
	if in.Spec.DataConnector != nil {
		md.DataConnector = in.Spec.DataConnector.FQN()
	}
	if md.Builder == "" {
		// Todo auto-detect builder
		md.Builder = ExpressionBuilder
	}

	if len(md.Aggr) > 0 && !md.ValidWindow() {
		return nil, fmt.Errorf("invalid feature specification for windowed feature")
	}
	return md, nil
}

func NormalizeFQN(fqn, defaultNamespace string) string {
	ns := strings.Index(fqn, ".")
	if ns != -1 {
		return fqn
	}

	fn := strings.Index(fqn, "[")
	if fn != -1 {
		return fmt.Sprintf("%s.%s%s", fqn, defaultNamespace, fqn[fn:])
	}
	return fmt.Sprintf("%s.%s", fqn, defaultNamespace)
}
