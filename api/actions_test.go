// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/tsuru/action"
	"github.com/globocom/tsuru/auth"
	"github.com/globocom/tsuru/db"
	"labix.org/v2/mgo/bson"
	"launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

type ActionsSuite struct {
	conn *db.Storage
}

var _ = gocheck.Suite(&ActionsSuite{})

func (s *ActionsSuite) SetUpSuite(c *gocheck.C) {
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "tsuru_api_actions_test")
	config.Set("auth:salt", "tsuru-salt")
	config.Set("auth:token-key", "TSURU-SALT")
	var err error
	s.conn, err = db.Conn()
	c.Assert(err, gocheck.IsNil)
}

func (s *ActionsSuite) startGandalfTestServer(h http.Handler) *httptest.Server {
	ts := httptest.NewServer(h)
	pieces := strings.Split(ts.URL, "://")
	protocol := pieces[0]
	hostPart := strings.Split(pieces[1], ":")
	port := hostPart[1]
	host := hostPart[0]
	config.Set("git:host", host)
	portInt, _ := strconv.ParseInt(port, 10, 0)
	config.Set("git:port", portInt)
	config.Set("git:protocol", protocol)
	return ts
}

func (s *ActionsSuite) TestAddKeyInGandalfActionForward(c *gocheck.C) {
	h := testHandler{}
	ts := s.startGandalfTestServer(&h)
	defer ts.Close()
	key := &auth.Key{Name: "mysshkey", Content: "my-ssh-key"}
	u := &auth.User{Email: "me@gmail.com", Password: "123456"}
	ctx := action.FWContext{
		Params: []interface{}{key, u},
	}
	result, err := addKeyInGandalfAction.Forward(ctx)
	c.Assert(err, gocheck.IsNil)
	c.Assert(result, gocheck.IsNil) // we're not gonna need the result
	c.Assert(len(h.url), gocheck.Equals, 1)
	expected := fmt.Sprintf("/user/%s/key", u.Email)
	c.Assert(h.url[0], gocheck.Equals, expected)
}

func (s *ActionsSuite) TestAddKeyInGandalfActionBackward(c *gocheck.C) {
	h := testHandler{}
	ts := s.startGandalfTestServer(&h)
	defer ts.Close()
	key := &auth.Key{Name: "mysshkey", Content: "my-ssh-key"}
	u := &auth.User{Email: "me@gmail.com", Password: "123456"}
	ctx := action.BWContext{
		Params: []interface{}{key, u},
	}
	addKeyInGandalfAction.Backward(ctx)
	c.Assert(len(h.url), gocheck.Equals, 1)
	expected := fmt.Sprintf("/user/%s/key/%s", u.Email, key.Name)
	c.Assert(h.url[0], gocheck.Equals, expected)
}

func (s *ActionsSuite) TestAddKeyInDatabaseActionForward(c *gocheck.C) {
	key := &auth.Key{Name: "mysshkey", Content: "my-ssh-key"}
	u := &auth.User{Email: "me@gmail.com", Password: "123456"}
	err := u.Create()
	defer s.conn.Users().Remove(bson.M{"email": u.Email})
	c.Assert(err, gocheck.IsNil)
	ctx := action.FWContext{
		Params: []interface{}{key, u},
	}
	result, err := addKeyInDatabaseAction.Forward(ctx)
	c.Assert(err, gocheck.IsNil)
	c.Assert(result, gocheck.IsNil) // we do not need it
	u.Get()
	c.Assert(u.Keys, gocheck.DeepEquals, []auth.Key{*key})
}

func (s *ActionsSuite) TestAddKeyInDatabaseActionBackward(c *gocheck.C) {
	key := &auth.Key{Name: "mysshkey", Content: "my-ssh-key"}
	u := &auth.User{Email: "me@gmail.com", Password: "123456"}
	u.AddKey(*key)
	err := u.Create()
	c.Assert(err, gocheck.IsNil)
	defer s.conn.Users().Remove(bson.M{"email": u.Email})
	c.Assert(u.Keys, gocheck.DeepEquals, []auth.Key{*key}) // just in case
	ctx := action.BWContext{
		Params: []interface{}{key, u},
	}
	addKeyInDatabaseAction.Backward(ctx)
	u.Get()
	c.Assert(u.Keys, gocheck.DeepEquals, []auth.Key{})
}

func (s *ActionsSuite) TestAddUserToTeamInGandalfActionForward(c *gocheck.C) {
	h := testHandler{}
	ts := s.startGandalfTestServer(&h)
	defer ts.Close()
	u := &auth.User{Email: "nobody@gmail.com", Password: "123456"}
	t := &auth.Team{Name: "myteam"}
	ctx := action.FWContext{
		Params: []interface{}{"me@gmail.com", u, t},
	}
	result, err := addUserToTeamInGandalfAction.Forward(ctx)
	c.Assert(err, gocheck.IsNil)
	c.Assert(result, gocheck.IsNil)
	c.Assert(len(h.url), gocheck.Equals, 1)
	c.Assert(h.url[0], gocheck.Equals, "/repository/grant")
}

func (s *ActionsSuite) TestAddUserToTeamInGandalfActionBackward(c *gocheck.C) {
	h := testHandler{}
	ts := s.startGandalfTestServer(&h)
	defer ts.Close()
	u := &auth.User{Email: "nobody@gmail.com", Password: "123456"}
	t := &auth.Team{Name: "myteam"}
	ctx := action.BWContext{
		Params: []interface{}{"me@gmail.com", u, t},
	}
	addUserToTeamInGandalfAction.Backward(ctx)
	c.Assert(len(h.url), gocheck.Equals, 1)
	c.Assert(h.url[0], gocheck.Equals, "/repository/revoke")
}

func (s *ActionsSuite) TestAddUserToTeamInDatabaseActionForward(c *gocheck.C) {
	u := &auth.User{Email: "nobody@gmail.com", Password: "123456"} // it's not used in this action
	newUser := &auth.User{Email: "me@gmail.com", Password: "123456"}
	err := newUser.Create()
	c.Assert(err, gocheck.IsNil)
	t := &auth.Team{Name: "myteam"}
	err = s.conn.Teams().Insert(t)
	c.Assert(err, gocheck.IsNil)
	defer s.conn.Teams().RemoveId(t.Name)
	defer s.conn.Users().Remove(bson.M{"email": newUser.Email})
	ctx := action.FWContext{
		Params: []interface{}{newUser.Email, u, t},
	}
	result, err := addUserToTeamInDatabaseAction.Forward(ctx)
	c.Assert(err, gocheck.IsNil)
	c.Assert(result, gocheck.IsNil)
	err = s.conn.Teams().FindId(t.Name).One(&t)
	c.Assert(err, gocheck.IsNil)
	c.Assert(t, ContainsUser, newUser)
}

func (s *ActionsSuite) TestAddUserToTeamInDatabaseActionBackward(c *gocheck.C) {
	u := &auth.User{Email: "nobody@gmail.com", Password: "123456"}
	err := u.Create()
	c.Assert(err, gocheck.IsNil)
	uEmail := "me@gmail.com"
	t := &auth.Team{Name: "myteam", Users: []string{uEmail}}
	err = s.conn.Teams().Insert(t)
	c.Assert(err, gocheck.IsNil)
	defer s.conn.Teams().RemoveId(t.Name)
	defer s.conn.Users().Remove(bson.M{"email": u.Email})
	ctx := action.BWContext{
		Params: []interface{}{uEmail, u, t},
	}
	addUserToTeamInDatabaseAction.Backward(ctx)
	err = s.conn.Teams().FindId(t.Name).One(&t)
	c.Assert(err, gocheck.IsNil)
	c.Assert(t, gocheck.Not(ContainsUser), &auth.User{Email: uEmail})
}
