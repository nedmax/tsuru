// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
crane is a command line tool for service providers/administrators.

It provides some commands that allow the service administrator to register
himself/herself, manage teams, apps and services.

Usage:

	% crane <command> [args]

The currently available commands are (grouped by subject):

	target            retrive the current tsuru server
	target-add        add a new tsuru server to target-list
	target-set        set current target
	target-remove     remove a tsuru server from target-list

	version           displays current tsuru version

	user-create       creates a new user
	user-remove       removes your user from tsuru server
	login             authenticates the user with tsuru server
	logout            finishes the session with tsuru server
	change-password   changes your password
	key-add           adds a public key to tsuru deploy server
	key-remove        removes a public key from tsuru deploy server

	team-create       creates a new team (adding the current user to it automatically)
	team-remove       removes a team from tsuru
	team-list         list teams that the user is member
	team-user-add     adds a user to a team
	team-user-remove  removes a user from a team

	template          generates a new manifest file, so you can just fill information for your service
	create            creates a new service from a manifest file
	update            updates a service using a manifest file
	remove            removes a service
	list              list all services that the user is administrator of

	doc-add           updates service's documentation
	doc-get           gets current docs of the service

Use "crane help <command>" for more information about a command.


Managing remote crane server endpoints

Usage:

	% tsuru target
	% tsuru target-add <label> <address> [--set-current]
	% tsuru target-set <label>
	% tsuru target-remove <label>

The target is the crane server to which all operations will be directed to.

With this set of commands you are be able to check the current target, add a new labeled target, set a target for usage,
list the added targets and remove a target, respectively.


Check current version

Usage:

	% crane version

This command returns the current version of crane command.


Create a user

Usage:

	% crane user-create <email>

user-create creates a user within crane remote server. It will ask for the
password before issue the request.


Remove your user from tsuru server

Usage:

	% crane user-remove

user-remove will remove currently authenticated user from remote tsuru server.
since there cannot exist any orphan teams, tsuru will refuse to remove a user
that is the last member of some team. if this is your case, make sure you
remove the team using "team-remove" before removing the user.


Authenticate within remote crane server

Usage:

	% crane login <email>

Login will ask for the password and check if the user is successfully
authenticated. If so, the token generated by the crane server will be stored in
${HOME}/.crane_token.

All crane actions require the user to be authenticated (except login and
user-create, obviously).


Logout from remote crane server

Usage:

	% crane logout

Logout will delete the token file and terminate the session within crane
server.


Change user's password

Usage:

	% tsuru change-password

change-password will change the password of the logged in user. It will ask for
the current password, the new and the confirmation.


Create a new team for the user

Usage:

	% crane team-create <teamname>

team-create will create a team for the user. crane requires a user to be a
member of at least one team in order to create a service.

When you create a team, you're automatically member of this team.


Remove a team from tsuru

Usage:

	% crane team-remove <team-name>

team-remove will remove a team from tsuru server. You're able to remove teams
that you're member of. A team that has access to any app cannot be removed.
Before removing a team, make sure it does not have access to any app (see
"app-grant" and "app-revoke" commands for details).


List teams that the user is member of

Usage:

	% crane team-list

team-list will list all teams that you are member of.


Add a user to a team

Usage:

	% crane team-user-add <teamname> <useremail>

team-user-add adds a user to a team. You need to be a member of the team to be
able to add another user to it.


Remove a user from a team

Usage:

	% crane team-user-remove <teamname> <useremail>

team-user-remove removes a user from a team. You need to be a member of the
team to be able to remove a user from it.

A team can never have 0 users. If you are the last member of a team, you can't
remove yourself from it.


Create an empty manifest file

Usage:

	% crane template

Template will create a file named "manifest.yaml" with the following content:

	id: servicename
	endpoint:
	  production: production-endpoint.com
	  test: test-endpoint.com:8080

Change it at will to configure your service. Id is the id of your service, it
must be unique. You must provide a production endpoint that will be invoked by
tsuru when application developers ask for new instances and are binding their
apps to their instances. For more details, see the text "Services API
Workflow": http://tsuru.rtfd.org/services-api-workflow.


Create a new service

Usage:

	% crane create <manifest-file.yaml>

Create will create a new service with information present in the manifest file.
Here is an example of usage:

	% cat /home/gopher/projects/mysqlapi/manifest.yaml
	id: mysqlapi
	endpoint:
	  production: https://mysqlapi.com:7777
	% crane create /home/gopher/projects/mysqlapi/manifest.yaml
	success

You can use "crane template" to generate a template. Both id and production
endpoint are required fields.

When creating a new service, crane will add all user's teams as administrator
teams of the service.


Update a service

Usage:

	% crane update <manifest-file.yaml>

Update will update a service using a manifest file. Currently, it's only
possible to edit an endpoint, or add new endpoints. You need to be an
administrator of the team to perform an update.


Remove a service

Usage:

	% crane remove <service-id>

Remove will remove a service from crane server. You need to be an administrator
of the team to remove it.


List services that you administrate

Usage:

	% crane list

List will list all services that you administrate, and the instances of each
service, created by application developers.


Update service's documentation

Usage:

	% crane doc-add <service-id> <doc-file.txt>

doc-add will update service's doc. Example of usage:

	% cat doc.txt
	mysqlapi

	This service is used for mysql connections.

	Once bound, you will be able to use the following environment variables:

		- MYSQL_HOST: host of MySQL server
		- MYSQL_PORT: port of MySQL instance
		- MYSQL_DATABASE_NAME: name of the database
		- MYSQL_USER: MySQL user for connections
		- MYSQL_PASSWORD: MySQL password for connections
	% crane doc-add mysqlapi doc.txt
	Documentation for 'mysqlapi' successfully updated.

You need to be an administrator of the service to update its docs.


Retrieve service's documentation

Usage:

	% crane doc-get <service-id>

doc-get will retrieve the current documentation of the service.
*/
package main
