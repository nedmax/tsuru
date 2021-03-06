.. Copyright 2013 tsuru authors. All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.

++++++++++++++++++++++++++++++++++++++
Build your own PaaS with tsuru and lxc
++++++++++++++++++++++++++++++++++++++

This document describes how to create a private PaaS service using tsuru and lxc.

This document assumes that tsuru is being installed on a Ubuntu (12.10) machine. You
can use equivalent packages for beanstalkd, git, MongoDB and other tsuru
dependencies. Please make sure you satisfy minimal version requirements.

Overview
========

The Tsuru PaaS is composed by multiple components:

* tsuru api server
* tsuru collector
* lxc
* gandalf api server
* gandalf wrapper
* git daemon
* charms
* mongodb (database)
* beanstalkd (queue server)

Installing
==========

lxc
---

Install the lxc, by doing this:

.. highlight:: bash

::

    $ sudo apt-get install -y lxc

MongoDB
-------

Tsuru needs MongoDB stable, distributed by 10gen. `It's pretty easy to
get it running on Ubuntu <http://docs.mongodb.org/manual/tutorial/install-mongodb-on-ubuntu/>`_

Beanstalkd
----------

Tsuru uses `Beanstalkd <http://kr.github.com/beanstalkd/>`_ as a work queue.
Install the latest version, by doing this:

.. highlight:: bash

::

    $ sudo apt-get install -y beanstalkd

Gandalf
-------

Tsuru uses `Gandalf <https://github.com/globocom/gandalf>`_ to manage git repositories, to get it installed `follow this steps <https://gandalf.readthedocs.org/en/latest/install.html>`_

Tsuru api and collector
-----------------------

You can download pre-built binaries of tsuru and collector. There are binaries
available only for Linux 64 bits, so make sure that ``uname -m`` prints
``x86_64``:

.. highlight:: bash

::

    $ uname -m
    x86_64

Then download and install the binaries. First, collector:

.. highlight:: bash

::

    $ curl -sL https://s3.amazonaws.com/tsuru/dist-server/tsuru-collector.tar.gz | sudo tar -xz -C /usr/bin

Then the API server:

.. highlight:: bash

::

    $ curl -sL https://s3.amazonaws.com/tsuru/dist-server/tsuru-api.tar.gz | sudo tar -xz -C /usr/bin

These commands will install ``collector`` and ``api`` commands in ``/usr/bin``
(you will need to be a sudoer and provide your password). You may install these
commands somewhere else in your ``PATH``.

Configuring
===========

Before running tsuru, you must configure it. By default, tsuru will look for
the configuration file in the ``/etc/tsuru/tsuru.cnf`` path. You can check a
sample configuration file and documentation for each tsuru setting in the
:doc:`"Configuring tsuru" </config>` page.

You can download the sample configuration file from Github:

.. highlight:: bash

::

    $ [sudo] mkdir /etc/tsuru
    $ [sudo] curl -sL https://raw.github.com/globocom/tsuru/master/etc/tsuru.conf -o /etc/tsuru/tsuru.conf

Make sure you define the required settings (database connection, authentication
configuration, AWS credentials, etc.) before running tsuru.

Running tsuru
=============

Now that you have ``api`` and ``collector`` properly installed, and you
:doc:`configured tsuru </config>`, you're three steps away from running it.

1. Start mongodb

.. highlight:: bash

::

    $ sudo service mongodb  start

2. Start beanstalkd

.. highlight:: bash

::

    $ sudo service beanstalkd start

3. Start tsuru and collector

.. highlight:: bash

::

    $ api &
    $ collector &

One can see the logs in:

.. highlight:: bash

::

    $ tail -f /var/log/syslog

Using tsuru
===========

Congratulations! At this point you should have a working tsuru server running
on your machine, follow the :doc:`tsuru client usage guide
</apps/client/usage>` to start build your apps.
