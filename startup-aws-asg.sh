#!/bin/bash
yum install -y gcc git
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
git clone https://github.com/aws/clock-bound
cd /clock-bound/clock-bound-d/
/root/.cargo/bin/cargo build --release
echo '# Ref: https://github.com/aws/clock-bound/tree/main/clock-bound-d' >> /etc/chrony.d/clockbound.conf
echo 'maxclockerror 50' >> /etc/chrony.d/clockbound.conf
systemctl restart chronyd
systemctl status chronyd
cp -v /clock-bound/target/release/clockbound /usr/local/bin/clockbound
chown chrony:chrony /usr/local/bin/clockbound

cat >/usr/lib/systemd/system/clockbound.service <<EOL
[Unit]
Description=ClockBound

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/clockbound --max-drift-rate 50
RuntimeDirectory=clockbound
RuntimeDirectoryPreserve=yes
WorkingDirectory=/run/clockbound
User=chrony
Group=chrony

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable clockbound
systemctl start clockbound
systemctl status clockbound

cd /clock-bound/clock-bound-ffi/
/root/.cargo/bin/cargo build --release
cp -v /clock-bound/clock-bound-ffi/include/clockbound.h /usr/include/
cp -v /clock-bound/target/release/libclockbound.a /usr/lib/
cp -v /clock-bound/target/release/libclockbound.so /usr/lib/

METADATA_TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
INTERNAL_IP=$(curl -H "X-aws-ec2-metadata-token: $METADATA_TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4)
echo "internal-ip: $INTERNAL_IP"
