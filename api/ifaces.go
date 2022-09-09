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
	"time"

	"github.com/go-logr/logr"
	manifests "github.com/raptor-ml/raptor/api/v1alpha1"
)

// Engine is the main engine of the Core
// It is responsible for the low-level operation for the features against the feature store
type Engine interface {
	Metadata(ctx context.Context, fqn string) (Metadata, error)
	Get(ctx context.Context, fqn string, entityID string) (Value, Metadata, error)
	Set(ctx context.Context, fqn string, entityID string, val any, ts time.Time) error
	Append(ctx context.Context, fqn string, entityID string, val any, ts time.Time) error
	Incr(ctx context.Context, fqn string, entityID string, by any, ts time.Time) error
	Update(ctx context.Context, fqn string, entityID string, val any, ts time.Time) error
}
type MetadataGetter func(ctx context.Context, fqn string) (Metadata, error)

// Logger is a simple interface that returns a Logr.Logger
type Logger interface {
	Logger() logr.Logger
}

// FeatureManager is managing Feature(s) within Core
// It is responsible for managing features as well as operating on them
type FeatureManager interface {
	BindFeature(in *manifests.Feature) error
	UnbindFeature(fqn string) error
	HasFeature(fqn string) bool
}

// DataConnectorManager is managing DataConnector(s) within Core
// It is responsible for maintaining the DataConnector(s) in an internal store
type DataConnectorManager interface {
	BindDataConnector(md DataConnector) error
	UnbindDataConnector(fqn string) error
	HasDataConnector(fqn string) bool
}

// DataConnectorGetter is a simple interface that returns a DataConnector
type DataConnectorGetter interface {
	GetDataConnector(fqn string) (DataConnector, error)
}

// EngineWithConnector is an Engine that has a DataConnector
type EngineWithConnector interface {
	Engine
	DataConnectorGetter
}

// ManagerEngine is the business-logic implementation of the Core
type ManagerEngine interface {
	Logger
	FeatureManager
	DataConnectorManager
	Engine
}
