[![main](https://github.com/flowerinthenight/spindle-cb/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/spindle-cb/actions/workflows/main.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/flowerinthenight/spindle-cb.svg)](https://pkg.go.dev/github.com/flowerinthenight/spindle-cb)

## spindle-cb

A distributed locking library (port of [spindle](https://github.com/flowerinthenight/spindle)) built on [aws/clock-bound](https://github.com/aws/clock-bound), and PostgreSQL (storage).

Using this library requires the [ClockBound daemon](https://github.com/aws/clock-bound/tree/main/clock-bound-d) (making it AWS-only), and [CGO](https://pkg.go.dev/cmd/cgo), due to its dependency on the [clockbound-ffi-go](https://github.com/flowerinthenight/clockbound-ffi-go) library.

## Usage

At the moment, the table needs to be created beforehand(`spindle` and `locktable` are just examples):

```sql
-- create the database:
CREATE DATABASE spindle;

-- create the table:
CREATE TABLE locktable (
    name TEXT PRIMARY KEY,
    heartbeat TIMESTAMP,
    token TIMESTAMP,
    writer TEXT
);
```

After creating the lock object, you will call the `Run(...)` function which will attempt to acquire a named lock at a regular interval (lease duration) until cancelled. A `HasLock()` function is provided which returns true (along with the lock token) if the lock is successfully acquired. Something like:

```go
import (
    ...
    "github.com/flowerinthenight/spindle-cb"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    // error checks redacted
    db, _ := sql.Open("pgx", *dbstr)
    defer db.Close()

    done := make(chan error, 1) // notify me when done (optional)
    quit, cancel := context.WithCancel(context.Background()) // for cancel
    
    // Create the lock object using a 5s lease duration using 'locktable' above.
    lock := spindle.New(db, "locktable", "mylock", spindle.WithDuration(5000))
    
    lock.Run(quit, done) // start the main loop, async

    time.Sleep(time.Second * 20)
    locked, token := lock.HasLock()
    log.Println("HasLock:", locked, token)
    time.Sleep(time.Second * 20)
    
    cancel()
    <-done
}
```

A sample cloud-init [startup script](./startup-aws-asg.sh) is provided for spinning up an [Auto Scaling Group](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html) with the ClockBound daemon already setup and running.

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

# Run multiple instances of `example` to see locking taking place:
$ ./example -db postgres://postgres:pass@loc.rds.amazonaws.com:5432/spindle
```

<p align="left">
  <img src="./assets/spindle-cb.png" title="spindle">
</p>

## License

This library is licensed under the [Apache 2.0 License](./LICENSE).
