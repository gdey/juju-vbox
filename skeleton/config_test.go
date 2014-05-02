// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package skeleton_test

import (
	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/provider/skeleton"
	"launchpad.net/juju-core/testing"
	"launchpad.net/juju-core/testing/testbase"
)

func newConfig(c *gc.C, attrs testing.Attrs) *config.Config {
	attrs = testing.FakeConfig().Merge(attrs)
	cfg, err := config.New(config.NoDefaults, attrs)
	c.Assert(err, gc.IsNil)
	return cfg
}

func validAttrs() testing.Attrs {
	return testing.FakeConfig().Merge(testing.Attrs{
		"type":                     "skeleton",
		"skeleton-secret-field":    "seekrit",
		"skeleton-immutable-field": "static",
	})
}

type ConfigSuite struct {
	testbase.LoggingSuite
}

var _ = gc.Suite(&ConfigSuite{})

var newConfigTests = []struct {
	info   string
	insert testing.Attrs
	remove []string
	expect testing.Attrs
	err    string
}{{
	info:   "skeleton-immutable-field is required",
	remove: []string{"skeleton-immutable-field"},
	err:    "skeleton-immutable-field: expected string, got nothing",
}, {
	info:   "skeleton-immutable-field cannot be empty",
	insert: testing.Attrs{"skeleton-immutable-field": ""},
	err:    "skeleton-immutable-field: must not be empty",
}, {
	info:   "skeleton-secret-field is required",
	remove: []string{"skeleton-secret-field"},
	err:    "skeleton-secret-field: expected string, got nothing",
}, {
	info:   "skeleton-secret-field cannot be empty",
	insert: testing.Attrs{"skeleton-secret-field": ""},
	err:    "skeleton-secret-field: must not be empty",
}, {
	info:   "skeleton-default-field is inserted if missing",
	expect: testing.Attrs{"skeleton-default-field": "<specific default value>"},
}, {
	info:   "skeleton-default-field cannot be empty",
	insert: testing.Attrs{"skeleton-default-field": ""},
	err:    "skeleton-default-field: must not be empty",
}, {
	info:   "skeleton-default-field is untouched if present",
	insert: testing.Attrs{"skeleton-default-field": "<user value>"},
	expect: testing.Attrs{"skeleton-default-field": "<user value>"},
}, {
	info:   "unknown field is not touched",
	insert: testing.Attrs{"unknown-field": 12345},
	expect: testing.Attrs{"unknown-field": 12345},
}}

func (*ConfigSuite) TestNewEnvironConfig(c *gc.C) {
	for i, test := range newConfigTests {
		c.Logf("test %d: %s", i, test.info)
		attrs := validAttrs().Merge(test.insert).Delete(test.remove...)
		testConfig := newConfig(c, attrs)
		environ, err := environs.New(testConfig)
		if test.err == "" {
			c.Check(err, gc.IsNil)
			attrs := environ.Config().AllAttrs()
			for field, value := range test.expect {
				c.Check(attrs[field], gc.Equals, value)
			}
		} else {
			c.Check(environ, gc.IsNil)
			c.Check(err, gc.ErrorMatches, test.err)
		}
	}
}

func (*ConfigSuite) TestValidateNewConfig(c *gc.C) {
	for i, test := range newConfigTests {
		c.Logf("test %d: %s", i, test.info)
		attrs := validAttrs().Merge(test.insert).Delete(test.remove...)
		testConfig := newConfig(c, attrs)
		validatedConfig, err := skeleton.Provider.Validate(testConfig, nil)
		if test.err == "" {
			c.Check(err, gc.IsNil)
			attrs := validatedConfig.AllAttrs()
			for field, value := range test.expect {
				c.Check(attrs[field], gc.Equals, value)
			}
		} else {
			c.Check(validatedConfig, gc.IsNil)
			c.Check(err, gc.ErrorMatches, "invalid config: "+test.err)
		}
	}
}

func (*ConfigSuite) TestValidateOldConfig(c *gc.C) {
	knownGoodConfig := newConfig(c, validAttrs())
	for i, test := range newConfigTests {
		c.Logf("test %d: %s", i, test.info)
		attrs := validAttrs().Merge(test.insert).Delete(test.remove...)
		testConfig := newConfig(c, attrs)
		validatedConfig, err := skeleton.Provider.Validate(knownGoodConfig, testConfig)
		if test.err == "" {
			c.Check(err, gc.IsNil)
			attrs := validatedConfig.AllAttrs()
			for field, value := range validAttrs() {
				c.Check(attrs[field], gc.Equals, value)
			}
		} else {
			c.Check(validatedConfig, gc.IsNil)
			c.Check(err, gc.ErrorMatches, "invalid base config: "+test.err)
		}
	}
}

var changeConfigTests = []struct {
	info   string
	insert testing.Attrs
	remove []string
	expect testing.Attrs
	err    string
}{{
	info:   "no change, no error",
	expect: validAttrs(),
}, {
	info:   "can change skeleton-secret-field",
	insert: testing.Attrs{"skeleton-secret-field": "okkult"},
	expect: testing.Attrs{"skeleton-secret-field": "okkult"},
}, {
	info:   "can change skeleton-default-field",
	insert: testing.Attrs{"skeleton-default-field": "different"},
	expect: testing.Attrs{"skeleton-default-field": "different"},
}, {
	info:   "cannot change skeleton-immutable-field",
	insert: testing.Attrs{"skeleton-immutable-field": "mutant"},
	err:    "skeleton-immutable-field: cannot change from static to mutant",
}, {
	info:   "can insert unknown field",
	insert: testing.Attrs{"unknown": "ignoti"},
	expect: testing.Attrs{"unknown": "ignoti"},
}}

func (s *ConfigSuite) TestValidateChange(c *gc.C) {
	baseConfig := newConfig(c, validAttrs())
	for i, test := range changeConfigTests {
		c.Logf("test %d: %s", i, test.info)
		attrs := validAttrs().Merge(test.insert).Delete(test.remove...)
		testConfig := newConfig(c, attrs)
		validatedConfig, err := skeleton.Provider.Validate(testConfig, baseConfig)
		if test.err == "" {
			c.Check(err, gc.IsNil)
			attrs := validatedConfig.AllAttrs()
			for field, value := range test.expect {
				c.Check(attrs[field], gc.Equals, value)
			}
		} else {
			c.Check(validatedConfig, gc.IsNil)
			c.Check(err, gc.ErrorMatches, "invalid config change: "+test.err)
		}
	}
}

func (s *ConfigSuite) TestSetConfig(c *gc.C) {
	baseConfig := newConfig(c, validAttrs())
	for i, test := range changeConfigTests {
		c.Logf("test %d: %s", i, test.info)
		environ, err := environs.New(baseConfig)
		c.Assert(err, gc.IsNil)
		attrs := validAttrs().Merge(test.insert).Delete(test.remove...)
		testConfig := newConfig(c, attrs)
		err = environ.SetConfig(testConfig)
		newAttrs := environ.Config().AllAttrs()
		if test.err == "" {
			c.Check(err, gc.IsNil)
			for field, value := range test.expect {
				c.Check(newAttrs[field], gc.Equals, value)
			}
		} else {
			c.Check(err, gc.ErrorMatches, test.err)
			for field, value := range baseConfig.UnknownAttrs() {
				c.Check(newAttrs[field], gc.Equals, value)
			}
		}
	}
}
