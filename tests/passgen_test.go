package tests

import (
  . "gopkg.in/check.v1"
  "github.com/schmich/ward/passgen"
  "strings"
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
  g.SetLength(5, 5)
  g.AddAlphabet("a")
  p, _ := g.Generate()
  c.Assert(len(p), Equals, 5)
  c.Assert(p, Equals, "aaaaa")
}

func (s *PassgenSuite) TestMultiCharGenerate(c *C) {
  g := passgen.New()
  g.SetLength(10, 20)
  abc := g.AddAlphabet("abc")
  abc.SetMinMax(0, 100)
  p, _ := g.Generate()
  c.Assert(p, Matches, "[abc]{10,20}")
}

func (s *PassgenSuite) TestMinError(c *C) {
  g := passgen.New()
  g.SetLength(10, 10)
  abc := g.AddAlphabet("abc")
  abc.SetMin(100)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestMaxError(c *C) {
  g := passgen.New()
  g.SetLength(10, 10)
  abc := g.AddAlphabet("abc")
  abc.SetMax(1)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}

func (s *PassgenSuite) TestMultiAlphabet(c *C) {
  g := passgen.New()
  g.SetLength(8, 8)
  a := g.AddAlphabet("a")
  a.SetMinMax(3, 3)
  b := g.AddAlphabet("b")
  b.SetMinMax(5, 5)
  p, _ := g.Generate()
  c.Assert(strings.Replace(p, "b", "", -1), Equals, "aaa")
  c.Assert(strings.Replace(p, "a", "", -1), Equals, "bbbbb")
}

func (s *PassgenSuite) TestExclude(c *C) {
  g := passgen.New()
  g.SetLength(100, 100)
  g.AddAlphabet("abc")
  g.Exclude = "bc"
  p, _ := g.Generate()
  c.Assert(p, Matches, "a{100}")
}

func (s *PassgenSuite) TestImpossibleConstraint(c *C) {
  g := passgen.New()
  g.SetLength(10, 10)
  a := g.AddAlphabet("a")
  a.SetMinMax(6, 10)
  b := g.AddAlphabet("b")
  b.SetMinMax(6, 10)
  _, err := g.Generate()
  c.Assert(err, NotNil)
}
