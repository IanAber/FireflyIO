systemctl stop FireflyIO
cp bin/amd64/FireflyIO /usr/bin
cp -r web/* /FireflyIO/web
systemctl start FireflyIO
