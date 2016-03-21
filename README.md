# ec2_snapshot
[![Build Status](https://travis-ci.org/PermissionData/ec2_snapshot.svg?branch=master)](https://travis-ci.org/PermissionData/ec2_snapshot)
Used to take snapshots of EC2 instances and keep them for a user specified time period.  Can be used with chron to take periodic snapshots.   


## Usage
**get it**
```bash
$ go get github.com/PermissionData/ec2_snapshot
```
**set up aws credentials in ~/.aws/credentials**
```bash
$ cat ~/.aws/credentials
[default]
aws_access_key_id = <aws_access_key_id>
aws_secret_access_key = <aws_secret_access_key>
```
**run tool**
```bash
$ ./ec2_snapshot --image-name someimage.backup --instance-id i-1234abc --time-to-save 302400
```
The 'time-to-save' argument specifies the amount of time (in seconds) to keep backups for.  All images created before the time-to-save value will be deleted.  By default, if no CLI argument is passed, the value for 'time-to-save' is 604800 seconds.

There is also an optional 'log-location' CLI argument to specify where to log results to.  By default, logging is set to StdOut.  Logging is kept at a minimum, logging only results and/or errors.

Both 'image-name' and 'instance-id' are required.  Image name has to be a minimum of 4 characters.  AMIs are saved as '<imagename>.<timestamp>'

Filters for querying AWS is configured thru a yaml config file.  By default the location is './config.yml', but can be overwritten by using the 'config-location' CLI arg.  See config.yml.sample for an example.
