# Patrol
### Process and Service Management


## Building Patrol
```bash
go get -t -u sabey.co/patrol
cd ~/go/src/sabey.co/patrol/patrol
go build -a -v
./patrol
# HTML GUI: http://localhost:8421
```


## [Installing Patrol - systemd](https://github.com/sabey/patrol/tree/master/patrol)
```bash
# see: `patrol.service`
# edit `WorkingDirectory` to your Patrol working directory
# edit `ExecStart` to the location of your Patrol binary
sudo nano /lib/systemd/system/patrol.service
sudo systemctl enable patrol.service
sudo systemctl start patrol
# Verify Patrol is running!
sudo journalctl -f -u patrol
```


