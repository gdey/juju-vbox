// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package skeleton_test

import (
	"testing"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/provider/skeleton"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type SkeletonSuite struct{}

var _ = gc.Suite(&SkeletonSuite{})

func (*SkeletonSuite) TestRegistered(c *gc.C) {
	provider, err := environs.Provider("skeleton")
	c.Assert(err, gc.IsNil)
	c.Assert(provider, gc.Equals, skeleton.Provider)
}
