A port of [spindle](https://github.com/flowerinthenight/spindle) using [aws/clock-bound](https://github.com/aws/clock-bound). Using this library requires [CGO](https://pkg.go.dev/cmd/cgo) due to its usage of [clockbound-ffi-go](https://github.com/flowerinthenight/clockbound-ffi-go).

A cloud-init [startup script](./startup-aws-asg.sh) is also provided for spinning up an [ASG](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html) with the ClockBound daemon already setup.

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

# Create the ASG:
$ aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name spindle-asg \
  --launch-template LaunchTemplateName=spindle-lt,Version='1' \
  --min-size 1 \
  --max-size 1 \
  --tags Key=Name,Value=spindle-asg \
  --availability-zones {target-zone}

# You can view the logs through:
$ [sudo] journalctl -f
```
