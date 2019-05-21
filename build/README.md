# Running With Vagrant

You can easily get started with a development environment running in a VM using [Vagrant](https://www.vagrantup.com/) and [Virtual Box](https://www.virtualbox.org/wiki/Downloads).

Once you have Vagrant and Virtual Box installed you can change into either the `ubuntu` or `centos` directories and run `vagrant up`.

This will provision a VM that installs the necessary dev dependencies and runs the Flipt test suite.

Once the provisioning process is complete, run:

```shell
$ vagrant ssh
$ cd ~/app/flipt
$ make dev
```

This will run Flipt in development mode inside your VM.
