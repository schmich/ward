package tests

import (
  "github.com/schmich/ward/store"
  . "gopkg.in/check.v1"
  "strconv"
)

type StoreSuite struct {
}

var _ = Suite(&StoreSuite{})

func (s *StoreSuite) TestCreate(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  c.Assert(db, NotNil)
  db, _ = store.Create(":memory:", "pass", 500)
  c.Assert(db, NotNil)
  db, _ = store.Create(":memory:", "", 1)
  c.Assert(db, IsNil)
  db, _ = store.Create(":memory:", "pass", 0)
  c.Assert(db, IsNil)
}

func (s *StoreSuite) TestEmptyAllCredentials(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  credentials := db.AllCredentials()
  c.Assert(len(credentials), Equals, 0)
}

func assertCredentialsEqual(c *C, p *store.Credential, q *store.Credential) {
  c.Assert(p.Login, Equals, q.Login)
  c.Assert(p.Password, Equals, q.Password)
  c.Assert(p.Realm, Equals, q.Realm)
  c.Assert(p.Note, Equals, q.Note)
}

func (s *StoreSuite) TestAddCredential(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  credential := &store.Credential {
    Login: "login",
    Password: "password",
    Realm: "realm",
    Note: "note",
  }
  db.AddCredential(credential)
  credentials := db.AllCredentials()
  c.Assert(len(credentials), Equals, 1)
  assertCredentialsEqual(c, credentials[0], credential)
}

func (s *StoreSuite) TestAddManyCredentials(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  credential := &store.Credential {
    Login: "login",
    Password: "password",
    Realm: "realm",
    Note: "note",
  }
  for i := 0; i < 1000; i++ {
    db.AddCredential(credential)
  }
  credentials := db.AllCredentials()
  c.Assert(len(credentials), Equals, 1000)
}

func (s *StoreSuite) TestFindCredentials(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  foo := &store.Credential {
    Login: "foo",
    Password: "waldo",
    Realm: "shared",
    Note: "quux",
  }
  bar := &store.Credential {
    Login: "bar",
    Password: "asdf",
    Realm: "shared",
    Note: "wsad",
  }
  db.AddCredential(foo)
  db.AddCredential(bar)
  found := db.FindCredentials([]string { "foo" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], foo)
  found = db.FindCredentials([]string { "bar" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], bar)
  found = db.FindCredentials([]string { "shared" })
  c.Assert(len(found), Equals, 2)
  if found[0].Login == foo.Login {
    assertCredentialsEqual(c, found[0], foo)
    assertCredentialsEqual(c, found[1], bar)
  } else {
    assertCredentialsEqual(c, found[0], bar)
    assertCredentialsEqual(c, found[1], foo)
  }
  found = db.FindCredentials([]string { "bar", "shared" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], bar)
  found = db.FindCredentials([]string { "waldo" })
  c.Assert(len(found), Equals, 0)
  found = db.FindCredentials([]string { "wsad" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], bar)
  found = db.FindCredentials([]string { "bar", "shared", "wsad" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], bar)
  found = db.FindCredentials([]string { "oo", "uu" })
  c.Assert(len(found), Equals, 1)
  assertCredentialsEqual(c, found[0], foo)
}

func (s *StoreSuite) TestUpdateCredential(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  foo := &store.Credential {
    Login: "foo",
    Password: "bar",
    Realm: "baz",
    Note: "quux",
  }
  db.AddCredential(foo)
  updated := db.AllCredentials()[0]
  updated.Login = "updated"
  db.UpdateCredential(updated)
  credentials := db.AllCredentials()
  c.Assert(len(credentials), Equals, 1)
  new := credentials[0]
  assertCredentialsEqual(c, new, updated)
}

func (s *StoreSuite) TestDeleteCredential(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  foo := &store.Credential {
    Login: "foo",
    Password: "bar",
    Realm: "baz",
    Note: "quux",
  }
  db.AddCredential(foo)
  foo = db.AllCredentials()[0]
  db.DeleteCredential(foo)
  credentials := db.AllCredentials()
  c.Assert(len(credentials), Equals, 0)
}

func (s *StoreSuite) TestUpdateMasterPassword(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  credentials := make(map[string]*store.Credential, 100)
  for i := 0; i < len(credentials); i++ {
    login := strconv.Itoa(i)
    credentials[login] = &store.Credential {
      Login: login,
      Password: "bar",
      Realm: "baz",
      Note: "quux",
    }
    db.AddCredential(credentials[login])
  }
  c.Assert(len(db.AllCredentials()), Equals, len(credentials))
  db.UpdateMasterPassword("newpass", 100)
  newCredentials := db.AllCredentials()
  c.Assert(len(newCredentials), Equals, len(credentials))
  for _, credential := range newCredentials {
    assertCredentialsEqual(c, credential, credentials[credential.Login])
  }
}

func (s *StoreSuite) TestClose(c *C) {
  db, _ := store.Create(":memory:", "pass", 1)
  db.Close()
}
