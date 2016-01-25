package tests

import (
  . "gopkg.in/check.v1"
  "github.com/schmich/ward/passgen"
)

type PassgenSuite struct {
}

var _ = Suite(&PassgenSuite{})

func (s *PassgenSuite) TestNew(c *C) {
  g := passgen.New()
  c.Assert(g, NotNil)
}

func (s *PassgenSuite) TestEmptyError(c *C) {
  g := passgen.New()
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestNoLengthError(c *C) {
  g := passgen.New()
  abc := g.AddAlphabet("abc")
  abc.SetMinMax(0, 100)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestSingleCharGenerate(c *C) {
  g := passgen.New()
  g.AddAlphabet("a")
  g.SetLength(5, 5)
  p, _ := g.Generate()
  c.Assert(len(p), Equals, 5)
  c.Assert(p, Equals, "aaaaa")
}

func (s *PassgenSuite) TestMultiCharGenerate(c *C) {
  g := passgen.New()
  abc := g.AddAlphabet("abc")
  abc.SetMinMax(0, 100)
  g.SetLength(10, 20)
  p, _ := g.Generate()
  c.Assert(p, Matches, "[abc]{10,20}")
}

func (s *PassgenSuite) TestMinError(c *C) {
  g := passgen.New()
  abc := g.AddAlphabet("abc")
  abc.SetMin(100)
  g.SetLength(10, 10)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestMaxError(c *C) {
  g := passgen.New()
  abc := g.AddAlphabet("abc")
  abc.SetMax(1)
  g.SetLength(10, 10)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}
