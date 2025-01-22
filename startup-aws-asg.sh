#!/bin/bash
yum install -y gcc git
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
git clone https://github.com/aws/clock-bound
cd clock-bound/clock-bound-d/
/root/.cargo/bin/cargo build --release
echo '# Ref: https://github.com/aws/clock-bound/tree/main/clock-bound-d' >> /etc/chrony.d/clockbound.conf
echo 'maxclockerror 50' >> /etc/chrony.d/clockbound.conf
systemctl restart chronyd
systemctl status chronyd
cp /clock-bound/target/release/clockbound /usr/local/bin/clockbound
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

# wget https://github.com/flowerinthenight/zgroup/releases/download/v0.3.2/zgroup-v0.3.2-x86_64-linux.tar.gz
# tar -xzvf zgroup-v0.3.2-x86_64-linux.tar.gz
METADATA_TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
INTERNAL_IP=$(curl -H "X-aws-ec2-metadata-token: $METADATA_TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4)
# ZGROUP_JOIN_PREFIX=0b9303ad-1beb-483f-abb5-bc58e0214531 ./zgroup group1 ${INTERNAL_IP}:8080 2>&1 | logger &