// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package skeleton

import (
	"fmt"

	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/schema"
	"launchpad.net/juju-core/utils"
)

// boilerplateConfig will be shown in help output, so please keep it up to
// date when you change environment configuration below.
const boilerplateConfig = `skeleton:
  type: skeleton

  # this exists to demonstrate how to deal with sensitive values.
  skeleton-secret-field: <cloud credentials, for example>

  # this exists to demonstrate how to deal with values that can't change; and
  # also how to use Prepare to fill in values that don't makes sense with
  # static defaults but can be inferred or chosen at runtime.
  # skeleton-immutable-field: <a storage bucket name, for example>

  # this exists to demonstrate how to deal with static default values that
  # some users may wish to override
  # skeleton-default-field: <specific default value>

`

var configFields = schema.Fields{
	"skeleton-secret-field":    schema.String(),
	"skeleton-immutable-field": schema.String(),
	"skeleton-default-field":   schema.String(),
}

var configDefaultFields = schema.Defaults{
	"skeleton-default-field": "<specific default value>",
}

var configSecretFields = []string{
	"skeleton-secret-field",
}

var configImmutableFields = []string{
	"skeleton-immutable-field",
}

func prepareConfig(cfg *config.Config) (*config.Config, error) {
	// Turn an incomplete config into a valid one, if possible.
	attrs := cfg.UnknownAttrs()
	if _, ok := attrs["skeleton-immutable-field"]; !ok {
		uuid, err := utils.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("cannot generate skeleton-immutable-field")
		}
		attrs["skeleton-immutable-field"] = fmt.Sprintf("%x", uuid.Raw())
	}
	return cfg.Apply(attrs)
}

func validateConfig(cfg *config.Config, old *environConfig) (*environConfig, error) {
	// Check sanity of juju-level fields.
	var oldCfg *config.Config
	if old != nil {
		oldCfg = old.Config
	}
	if err := config.Validate(cfg, oldCfg); err != nil {
		return nil, err
	}

	// Extract validated provider-specific fields. All of configFields will be
	// present in validated, and defaults will be inserted if necessary. If the
	// schema you passed in doesn't quite express what you need, you can make
	// whatever checks you need here, before continuing.
	// In particular, if you want to extract (say) credentials from the user's
	// shell environment variables, you'll need to allow missing values to pass
	// through the schema by setting a value of schema.Omit in the configFields
	// map, and then to set and check them at this point. These values *must* be
	// stored in newAttrs: a Config will be generated on the user's machine only
	// to begin with, and will subsequently be used on a different machine that
	// will probably not have those variables set.
	newAttrs, err := cfg.ValidateUnknownAttrs(configFields, configDefaultFields)
	if err != nil {
		return nil, err
	}
	for field := range configFields {
		if newAttrs[field] == "" {
			return nil, fmt.Errorf("%s: must not be empty", field)
		}
	}

	// If an old config was supplied, check any immutable fields have not changed.
	if old != nil {
		for _, field := range configImmutableFields {
			if old.attrs[field] != newAttrs[field] {
				return nil, fmt.Errorf(
					"%s: cannot change from %v to %v",
					field, old.attrs[field], newAttrs[field],
				)
			}
		}
	}

	// Merge the validated provider-specific fields into the original config,
	// to ensure the object we return is internally consistent.
	newCfg, err := cfg.Apply(newAttrs)
	if err != nil {
		return nil, err
	}
	return &environConfig{
		Config: newCfg,
		attrs:  newAttrs,
	}, nil
}

type environConfig struct {
	*config.Config
	attrs map[string]interface{}
}

func (ecfg *environConfig) secretField() string {
	return ecfg.attrs["skeleton-secret-field"].(string)
}

func (ecfg *environConfig) immutableField() string {
	return ecfg.attrs["skeleton-immutable-field"].(string)
}

func (ecfg *environConfig) defaultField() string {
	return ecfg.attrs["skeleton-default-field"].(string)
}
