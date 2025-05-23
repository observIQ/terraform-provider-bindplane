// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package client provides a client wrapper for bindplane-op/client.NewBindPlane.
// It provides wrapper functions suitable for Terraform use. For example, when
// BindPlane server returns a 404, this wrapper package will gracefully handle it
// instead of returning an error.
package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/observiq/bindplane-op-enterprise/client"
	"github.com/observiq/bindplane-op-enterprise/model"
)

// BindPlane is a shim layer between Terraform and the
// BindPlane client interface
type BindPlane struct {
	Client client.BindPlane
}

// Apply creates or updates a single BindPlane resource and returns it's id.
// If rollout is true, any configuration which is updated by the Apply
// opteration will have a rollout started.
func (i *BindPlane) Apply(r *model.AnyResource, rollout bool) error {
	status, err := i.Client.Apply(context.Background(), []*model.AnyResource{r})
	if err != nil {
		return fmt.Errorf("failed to apply BindPlane resources: %w", err)
	}

	var errs error

	for _, status := range status {
		resource := status.Resource

		// Apply expects the resource to be unchanged, configured, or
		// created. All other statuses are unexpected and should result
		// in an error from this method.
		switch status.Status {
		case model.StatusUnchanged:
		case model.StatusConfigured, model.StatusCreated:
			if rollout && status.Resource.Kind == model.KindConfiguration {
				if err := i.Rollout(resource.Name()); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		default:
			err := fmt.Errorf(
				"unexpected status when applying resource: %s, status: %s: reason: %s",
				resource.Name(),
				status.Status,
				status.Reason)
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// ApplyWithRetry wraps Apply with the ability to retry on retryable errors
func (i *BindPlane) ApplyWithRetry(ctx context.Context, timeout time.Duration, r *model.AnyResource, rollout bool) error {
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		if err := i.Apply(r, rollout); err != nil {
			if retryableError(err) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("bindplane apply retries exhausted: %w", err)
	}
	return nil
}

// Rollout starts a rollout against a named config
// TODO(jsirianni): Should Rollout block until it has finished or failed?
func (i *BindPlane) Rollout(name string) error {
	_, err := i.Client.StartRollout(context.Background(), name, nil)
	return err
}

// Connector takes a name and returns the matching connector
func (i *BindPlane) Connector(name string) (*model.Connector, error) {
	r, err := i.Client.Resource(context.Background(), model.KindConnector, name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connector with name %s: %w", name, err)
	}

	// Bindplane should always return a connector but we should handle it
	// anyway considering we need to type assert it.
	switch c := r.(type) {
	case *model.Connector:
		return c, nil
	default:
		return nil, fmt.Errorf("unexpected response from bindplane, expected connector, got %T, this is a bug that should be reported", c)
	}
}

// DeleteConnector will delete a BindPlane connector
func (i *BindPlane) DeleteConnector(name string) error {
	err := i.Client.DeleteResource(context.Background(), model.KindConnector, name)
	if err != nil {
		return fmt.Errorf("error while deleting connector with name %s: %w", name, err)
	}
	return nil
}

// Configuration takes a name and returns the matching configuration
func (i *BindPlane) Configuration(name string) (*model.Configuration, error) {
	c, err := i.Client.Configuration(context.Background(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get configuration with name %s: %w", name, err)
	}
	return c, nil
}

// DeleteConfiguration will delete a BindPlane configuration
func (i *BindPlane) DeleteConfiguration(name string) error {
	err := i.Client.DeleteConfiguration(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error while deleting configuration with name %s: %w", name, err)
	}
	return nil
}

// Destination takes a name and returns the matching destination
func (i *BindPlane) Destination(name string) (*model.Destination, error) {
	r, err := i.Client.Destination(context.Background(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get destination with name %s: %w", name, err)
	}
	return r, nil
}

// DeleteDestination will delete a BindPlane destination
func (i *BindPlane) DeleteDestination(name string) error {
	err := i.Client.DeleteDestination(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error while deleting destination with name %s: %w", name, err)
	}
	return nil
}

// Source takes a name and returns the matching source
func (i *BindPlane) Source(name string) (*model.Source, error) {
	r, err := i.Client.Source(context.Background(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get source with name %s: %w", name, err)
	}
	return r, nil
}

// DeleteSource will delete a BindPlane source
func (i *BindPlane) DeleteSource(name string) error {
	err := i.Client.DeleteSource(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error while deleting source with name %s: %w", name, err)
	}
	return nil
}

// Processor takes a name and returns the matching processor
func (i *BindPlane) Processor(name string) (*model.Processor, error) {
	r, err := i.Client.Processor(context.Background(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get processor with name %s: %w", name, err)
	}
	return r, nil
}

// DeleteProcessor will delete a BindPlane processor
func (i *BindPlane) DeleteProcessor(name string) error {
	err := i.Client.DeleteProcessor(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error while deleting processor with name %s: %w", name, err)
	}
	return nil
}

// Extension takes a name and returns the matching extension
func (i *BindPlane) Extension(name string) (*model.Extension, error) {
	r, err := i.Client.Extension(context.Background(), name)
	if err != nil {
		// Do not return an error if the resource is not found. Terraform
		// will understand that the resource does not exist when it receives
		// a nil value, and will instead offer to create the resource.
		if isNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get extension with name %s: %w", name, err)
	}
	return r, nil
}

// DeleteExtension will delete a Bindplane extension
func (i *BindPlane) DeleteExtension(name string) error {
	err := i.Client.DeleteExtension(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error while deleting extension with name %s: %w", name, err)
	}
	return nil
}

// Delete will delete a Bindplane resource
func (i *BindPlane) Delete(k model.Kind, name string) error {
	switch k {
	case model.KindConfiguration:
		return i.DeleteConfiguration(name)
	case model.KindDestination:
		return i.DeleteDestination(name)
	case model.KindSource:
		return i.DeleteSource(name)
	case model.KindProcessor:
		return i.DeleteProcessor(name)
	case model.KindExtension:
		return i.DeleteExtension(name)
	case model.KindConnector:
		return i.DeleteConnector(name)
	default:
		return fmt.Errorf("Delete does not support bindplane kind '%s'", k)
	}
}

// GenericResource represents a Bindplane resource's
// id, name, version, and ParameterizedSpec.
type GenericResource struct {
	ID      string
	Name    string
	Version model.Version
	Spec    model.ParameterizedSpec
}

// GenericResource looks up a Bindplane resource and returns a GenericResource.
// The returned GenericResource will be nil if it does not exist. It is up to
// the caller to check.
func (i *BindPlane) GenericResource(k model.Kind, name string) (*GenericResource, error) {
	g := &GenericResource{}

	switch k {
	case model.KindDestination:
		r, err := i.Destination(name)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, nil
		}

		g.ID = r.ID()
		g.Name = r.Name()
		g.Version = r.Version()
		g.Spec = r.Spec
	case model.KindSource:
		r, err := i.Source(name)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, nil
		}

		g.ID = r.ID()
		g.Name = r.Name()
		g.Version = r.Version()
		g.Spec = r.Spec
	case model.KindProcessor:
		r, err := i.Processor(name)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, nil
		}

		g.ID = r.ID()
		g.Name = r.Name()
		g.Version = r.Version()
		g.Spec = r.Spec
	case model.KindExtension:
		r, err := i.Extension(name)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, nil
		}

		g.ID = r.ID()
		g.Name = r.Name()
		g.Version = r.Version()
		g.Spec = r.Spec
	case model.KindConnector:
		r, err := i.Connector(name)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, nil
		}

		g.ID = r.ID()
		g.Name = r.Name()
		g.Version = r.Version()
		g.Spec = r.Spec
	default:
		return nil, fmt.Errorf("GenericResource does not support bindplane kind '%s'", k)
	}

	return g, nil
}

// TODO(jsirianni): Bindplane should probably have error types so we can check
// error.Is.
func isNotFoundError(err error) bool {
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "404 not found")
}
