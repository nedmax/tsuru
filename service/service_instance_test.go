// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"github.com/globocom/tsuru/app/bind"
	"github.com/globocom/tsuru/auth"
	"labix.org/v2/mgo/bson"
	"launchpad.net/gocheck"
)

func (s *S) createServiceInstance() {
	s.service = &Service{Name: "MySQL"}
	s.service.Create()
	s.serviceInstance = &ServiceInstance{
		Name: s.service.Name,
	}
	s.serviceInstance.Create()
}

func (s *S) TestCreateServiceInstance(c *gocheck.C) {
	s.createServiceInstance()
	defer s.conn.Services().Remove(bson.M{"name": s.service.Name})
	var result ServiceInstance
	query := bson.M{
		"name": s.service.Name,
	}
	err := s.conn.ServiceInstances().Find(query).One(&result)
	c.Check(err, gocheck.IsNil)
	c.Assert(result.Name, gocheck.Equals, s.service.Name)
}

func (s *S) TestDeleteServiceInstance(c *gocheck.C) {
	s.createServiceInstance()
	defer s.conn.Services().Remove(bson.M{"name": s.service.Name})
	s.serviceInstance.Delete()
	query := bson.M{
		"name": s.service.Name,
	}
	qtd, err := s.conn.ServiceInstances().Find(query).Count()
	c.Assert(err, gocheck.IsNil)
	c.Assert(qtd, gocheck.Equals, 0)
}

func (s *S) TestRetrieveAssociatedService(c *gocheck.C) {
	service := Service{Name: "my_service"}
	service.Create()
	serviceInstance := &ServiceInstance{
		Name:        service.Name,
		ServiceName: service.Name,
	}
	serviceInstance.Create()
	rService := serviceInstance.Service()
	c.Assert(service.Name, gocheck.Equals, rService.Name)
}

func (s *S) TestAddApp(c *gocheck.C) {
	instance := ServiceInstance{
		Name: "myinstance",
		Apps: []string{},
	}
	err := instance.AddApp("app1")
	c.Assert(err, gocheck.IsNil)
	c.Assert(instance.Apps, gocheck.DeepEquals, []string{"app1"})
}

func (s *S) TestAddAppReturnErrorIfTheAppIsAlreadyPresent(c *gocheck.C) {
	instance := ServiceInstance{
		Name: "myinstance",
		Apps: []string{"app1"},
	}
	err := instance.AddApp("app1")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, "^This instance already has this app.$")
}

func (s *S) TestFindApp(c *gocheck.C) {
	instance := ServiceInstance{
		Name: "myinstance",
		Apps: []string{"app1", "app2"},
	}
	c.Assert(instance.FindApp("app1"), gocheck.Equals, 0)
	c.Assert(instance.FindApp("app2"), gocheck.Equals, 1)
	c.Assert(instance.FindApp("what"), gocheck.Equals, -1)
}

func (s *S) TestRemoveApp(c *gocheck.C) {
	instance := ServiceInstance{
		Name: "myinstance",
		Apps: []string{"app1", "app2", "app3"},
	}
	err := instance.RemoveApp("app2")
	c.Assert(err, gocheck.IsNil)
	c.Assert(instance.Apps, gocheck.DeepEquals, []string{"app1", "app3"})
	err = instance.RemoveApp("app3")
	c.Assert(err, gocheck.IsNil)
	c.Assert(instance.Apps, gocheck.DeepEquals, []string{"app1"})
}

func (s *S) TestRemoveAppReturnsErrorWhenTheAppIsNotBoundToTheInstance(c *gocheck.C) {
	instance := ServiceInstance{
		Name: "myinstance",
		Apps: []string{"app1", "app2", "app3"},
	}
	err := instance.RemoveApp("app4")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err, gocheck.ErrorMatches, "^This app is not bound to this service instance.$")
}

func (s *S) TestServiceInstanceIsABinder(c *gocheck.C) {
	var _ bind.Binder = &ServiceInstance{}
}

func (s *S) TestGetServiceInstancesByServices(c *gocheck.C) {
	srvc := Service{Name: "mysql"}
	err := srvc.Create()
	c.Assert(err, gocheck.IsNil)
	sInstance := ServiceInstance{Name: "t3sql", ServiceName: "mysql", Teams: []string{s.team.Name}}
	err = sInstance.Create()
	c.Assert(err, gocheck.IsNil)
	sInstance2 := ServiceInstance{Name: "s9sql", ServiceName: "mysql"}
	err = sInstance2.Create()
	c.Assert(err, gocheck.IsNil)
	sInstances, err := GetServiceInstancesByServices([]Service{srvc})
	c.Assert(err, gocheck.IsNil)
	expected := []ServiceInstance{{Name: "t3sql", ServiceName: "mysql"}, sInstance2}
	c.Assert(sInstances, gocheck.DeepEquals, expected)
}

func (s *S) TestGetServiceInstancesByServicesWithoutAnyExistingServiceInstances(c *gocheck.C) {
	srvc := Service{Name: "mysql"}
	err := srvc.Create()
	c.Assert(err, gocheck.IsNil)
	sInstances, err := GetServiceInstancesByServices([]Service{srvc})
	c.Assert(err, gocheck.IsNil)
	c.Assert(sInstances, gocheck.DeepEquals, []ServiceInstance(nil))
}

func (s *S) TestGetServiceInstancesByServicesWithTwoServices(c *gocheck.C) {
	srvc := Service{Name: "mysql"}
	err := srvc.Create()
	c.Assert(err, gocheck.IsNil)
	defer srvc.Delete()
	srvc2 := Service{Name: "mongodb"}
	err = srvc2.Create()
	c.Assert(err, gocheck.IsNil)
	defer srvc.Delete()
	sInstance := ServiceInstance{Name: "t3sql", ServiceName: "mysql", Teams: []string{s.team.Name}}
	err = sInstance.Create()
	c.Assert(err, gocheck.IsNil)
	sInstance2 := ServiceInstance{Name: "s9nosql", ServiceName: "mongodb"}
	err = sInstance2.Create()
	c.Assert(err, gocheck.IsNil)
	sInstances, err := GetServiceInstancesByServices([]Service{srvc, srvc2})
	c.Assert(err, gocheck.IsNil)
	expected := []ServiceInstance{{Name: "t3sql", ServiceName: "mysql"}, sInstance2}
	c.Assert(sInstances, gocheck.DeepEquals, expected)
}

func (s *S) TestGenericServiceInstancesFilter(c *gocheck.C) {
	srvc := Service{Name: "mysql"}
	teams := []string{s.team.Name}
	q, f := genericServiceInstancesFilter(srvc, teams)
	c.Assert(q, gocheck.DeepEquals, bson.M{"service_name": srvc.Name, "teams": bson.M{"$in": teams}})
	c.Assert(f, gocheck.DeepEquals, bson.M{"name": 1, "service_name": 1, "apps": 1})
}

func (s *S) TestGenericServiceInstancesFilterWithServiceSlice(c *gocheck.C) {
	services := []Service{
		{Name: "mysql"},
		{Name: "mongodb"},
	}
	names := []string{"mysql", "mongodb"}
	teams := []string{s.team.Name}
	q, f := genericServiceInstancesFilter(services, teams)
	c.Assert(q, gocheck.DeepEquals, bson.M{"service_name": bson.M{"$in": names}, "teams": bson.M{"$in": teams}})
	c.Assert(f, gocheck.DeepEquals, bson.M{"name": 1, "service_name": 1, "apps": 1})
}

func (s *S) TestGenericServiceInstancesFilterWithoutSpecifingTeams(c *gocheck.C) {
	services := []Service{
		{Name: "mysql"},
		{Name: "mongodb"},
	}
	names := []string{"mysql", "mongodb"}
	teams := []string{}
	q, f := genericServiceInstancesFilter(services, teams)
	c.Assert(q, gocheck.DeepEquals, bson.M{"service_name": bson.M{"$in": names}})
	c.Assert(f, gocheck.DeepEquals, bson.M{"name": 1, "service_name": 1, "apps": 1})
}

func (s *S) TestGetServiceInstancesByServicesAndTeams(c *gocheck.C) {
	srvc := Service{Name: "mysql", Teams: []string{s.team.Name}, IsRestricted: true}
	srvc.Create()
	defer srvc.Delete()
	srvc2 := Service{Name: "mongodb", Teams: []string{s.team.Name}, IsRestricted: false}
	srvc2.Create()
	defer srvc2.Delete()
	sInstance := ServiceInstance{
		Name:        "j4sql",
		ServiceName: srvc.Name,
		Teams:       []string{s.team.Name},
	}
	sInstance.Create()
	defer sInstance.Delete()
	sInstance2 := ServiceInstance{
		Name:        "j4nosql",
		ServiceName: srvc2.Name,
		Teams:       []string{s.team.Name},
	}
	sInstance2.Create()
	defer sInstance2.Delete()
	sInstance3 := ServiceInstance{
		Name:        "f9nosql",
		ServiceName: srvc2.Name,
	}
	sInstance3.Create()
	defer sInstance3.Delete()
	expected := []ServiceInstance{
		{
			Name:        sInstance.Name,
			ServiceName: sInstance.ServiceName,
			Teams:       []string(nil),
			Apps:        []string{},
		},
		{
			Name:        sInstance2.Name,
			ServiceName: sInstance2.ServiceName,
			Teams:       []string(nil),
			Apps:        []string{},
		},
	}
	sInstances, err := GetServiceInstancesByServicesAndTeams([]Service{srvc, srvc2}, s.user)
	c.Assert(err, gocheck.IsNil)
	c.Assert(sInstances, gocheck.DeepEquals, expected)
}

func (s *S) TestGetServiceInstancesByServicesAndTeamsForUsersThatAreNotMembersOfAnyTeam(c *gocheck.C) {
	u := auth.User{Email: "noteamforme@globo.com", Password: "123"}
	err := u.Create()
	c.Assert(err, gocheck.IsNil)
	defer s.conn.Users().Remove(bson.M{"email": u.Email})
	srvc := Service{Name: "mysql", Teams: []string{s.team.Name}, IsRestricted: true}
	err = srvc.Create()
	c.Assert(err, gocheck.IsNil)
	defer srvc.Delete()
	instance := ServiceInstance{
		Name:        "j4sql",
		ServiceName: srvc.Name,
	}
	err = instance.Create()
	c.Assert(err, gocheck.IsNil)
	defer instance.Delete()
	instances, err := GetServiceInstancesByServicesAndTeams([]Service{srvc}, &u)
	c.Assert(err, gocheck.IsNil)
	c.Assert(instances, gocheck.IsNil)
}
