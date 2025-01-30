[![main](https://github.com/flowerinthenight/spindle-cb/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/spindle-cb/actions/workflows/main.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/flowerinthenight/spindle-cb.svg)](https://pkg.go.dev/github.com/flowerinthenight/spindle-cb)

## spindle-cb

A port of [spindle](https://github.com/flowerinthenight/spindle) using [aws/clock-bound](https://github.com/aws/clock-bound), and PostreSQL for storage.

Using this library requires [CGO](https://pkg.go.dev/cmd/cgo) due to its usage of [clockbound-ffi-go](https://github.com/flowerinthenight/clockbound-ffi-go).

To create the database and table (`spindle` and `locktable` are just examples):

```sql
-- create the database
CREATE DATABASE spindle;

-- create the table
CREATE TABLE locktable (
  name TEXT PRIMARY KEY,
	heartbeat TIMESTAMP,
	token TIMESTAMP,
	writer TEXT
);
```

A sample cloud-init [startup script](./startup-aws-asg.sh) is provided for spinning up an [ASG](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html) with the ClockBound daemon already setup.

```sh
# Create a launch template. ImageId here is Amazon Linux, default VPC.
# (Added newlines for readability. Might not run when copied as is.)
$ aws ec2 create-launch-template \
  --launch-template-name spindle-lt \
  --version-description version1 \
  --launch-template-data '
  {
    "UserData":"'"$(cat startup-aws-asg.sh | base64 -w 0)"'",
    "ImageId":"ami-0fb04413c9de69305",
    "InstanceType":"t2.micro",
  }'

# Create the ASG; update {target-zone} with actual value:
$ aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name spindle-asg \
  --launch-template LaunchTemplateName=spindle-lt,Version='1' \
  --min-size 1 \
  --max-size 1 \
  --tags Key=Name,Value=spindle-asg \
  --availability-zones {target-zone}

# You can now SSH to the instance. Note that it might take some time before
# ClockBound is running due to the need to build it in Rust. You can wait
# for the `clockbound` process, or tail the startup script output, like so:
$ tail -f /var/log/cloud-init-output.log

# Run the sample code:
# Download the latest release sample from GitHub.
$ tar xvzf spindle-{version}-x86_64-linux.tar.gz
$ ./example -db postgres://postgres:pass@loc.rds.amazonaws.com:5432/spindle
```
