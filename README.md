gothru
=======
gothru is a socks 5 server which will help you

* GO THROUGH a firewall
* conveniently manage the access to your server by adding white-listed users to the config file.


Developed by Qing Cheng to gothru the GREAT FIRE WALL. Contact me at qingchengnus@gmail.com.

## How To Get It

Find the executable in dist folder for you OS and Architechture and simply get it!

If you cannot find the executable for your system, you have to build it yourself.

Install Golang on your system and set you GOPATH environment variable.

then open your terminal and run

```go
go get github.com/qingchengnus/gothru
```
Now, if you have added $GOPATH/bin to your $PATH variable, you already can run gothru in your terminal!
Also, you can find the excutable in $GOPATH/bin and run it.

You got an error? Of course you will get an error, because you haven't set up your configuration file!

## Server Usage

First, have your config.xml file ready! Here is an example:
```xml
<config>
    <server_port>18888</server_port>
    <user>
    	<username>babyfatdragon</username>
    	<password>1352463570</password>
    </user>
    <user>
    	<username>queenofpain</username>
    	<password>987654321</password>
    </user>
</config>
```
You can have as many user as you want. Once you decide prevent a user from being able to access your server, just remove him from  the config file and restart the server!

After you set up the config.xml, run
```go
gothru -s //if your the config.xml is in your current directory and its name is exactly config.xml!
gothru -s -c ~/myconfig.xml //use -c to enter the file path to your configuration file.
```

## Client Usage
First, have your config.xml file ready! Here is an example:
```xml
<config>
    <server_address>123.456.789.123</server_address>
    <server_port>18888</server_port>
    <local_port>16666</local_port>
    <username>babyfatdragon</username>
    <password>1352463570</password>
</config>
```
After you set up the config.xml, run
```go
gothru //if your the config.xml is in your current directory and its name is exactly config.xml!
gothru -c ~/myconfig.xml //use -c to enter the file path to your configuration file.
```

Now you can set your browser or any other application like Dropbox to use gothru! Set the proxy type to socks version 5, server address to localhost, server port to 16666(as you defined in your config.xml). You don't need to set any username or password in the proxy setting panel, as you have set them in your config.xml.

You can also run
```go
gothru -h
```
for help.
## Future plan
* Allow server to block requests to certain hosts.(Black list)
* Allow server to only accept requests to certain hosts.(White list)
* Some statistics data, for example, amount of data transfered by each user.
* Set cap to data usage for each user.
* Contact me via qingchengnus@gmail.com if you have interesting ideas.
