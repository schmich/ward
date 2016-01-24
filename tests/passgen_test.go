package tests

import (
  "github.com/schmich/ward/passgen"
  . "gopkg.in/check.v1"
)

type PassgenSuite struct {
}

var _ = Suite(&PassgenSuite{})

func (s *PassgenSuite) TestNew(c *C) {
  g := passgen.New()
  c.Assert(g, NotNil)
}

func (s *PassgenSuite) TestEmptyGenerate(c *C) {
  g := passgen.New()
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestBasicGenerate(c *C) {
  g := passgen.New()
  g.AddAlphabet("basic", "abc")
  g.SetLength(10, 10)
  g.SetMinMax("basic", 0, 100)
  p, _ := g.Generate()
  c.Assert(len(p) > 0, Equals, true)
}
