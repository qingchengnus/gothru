gothru
=======
gothru is a socks 5 server which will help you

* GO THROUGH a firewall
* conveniently manage the access to your server by adding white-listed users to the config file.
* set the maximum amount of data a user can transfer through the server.


Developed by Qing Cheng to gothru the GREAT FIRE WALL. Contact me at qingchengnus@gmail.com.

## How To Get It

Find the executable in the dist folder for you OS and Architechture and simply get it!

If you cannot find the executable for your system, you have to build it yourself.

Install Golang on your system and set you GOPATH environment variable. If you are new to Go, follow http://golang.org/doc/code.html

Then open your terminal and run

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
        <datacap>10000</datacap>
    </user>
    <user>
    	<username>queenofpain</username>
    	<password>987654321</password>
    </user>
    <user>
        <username>queenofpain</username>
        <password>987654321</password>
        <datacap>0</datacap>
    </user>
</config>
```
You can have as many user as you want. Once you decide prevent a user from being able to access your server, just remove him from  the config file and restart the server!
Datacap is in MB. 10000 means 10000MB, after the user exceeded  the amount, he will not be able to use the server.
Not setting, or setting datacap to 0 means there is no data limit on this user.

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

## Mac Usage
Download the config.xml in dist/ and gothru executable file in dist/darwin-amd64/ and put them in the same directory. Here is an example of the config.xml:
```xml
<config>
    <server_address>123.456.789.123</server_address>
    <server_port>18888</server_port>
    <local_port>16666</local_port>
    <username>public4</username>
    <password>12345</password>
</config>
```
After you change the server_address to 128.199.143.166, which is my own instance, then run
```go
gothru //if your the config.xml is in your current directory and its name is exactly config.xml!
```

Now Open Network Preferences, click Advanced, choose Proxies, tick SOCKS Proxy, set SOCKS Proxy Server to 127.0.0.1:16666 and do not tick Proxy server requires password. Then click OK and also remember to click APPLY!

Now you can go through the GFW! Use browsers that use your system proxy settings, in my case, chrome and safari. You can also direct other software's traffic to 127.0.0.1:16666 to go through the GFW.

## TO DO
* Implement bind and udpassciate method
* Contact me via qingchengnus@gmail.com if you have interesting ideas.
