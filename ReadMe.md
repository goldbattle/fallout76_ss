# Fallout 76 Server Status Website


This website was a small 2018 Thanksgiving side project created to try to track the uptime of the Fallout 76 servers.
We will see if the servers "Just Work" as claimed.
This site is powered by a Golang webserver that queries the Bethesda API for server status and serves this as a website.
This is the first ever Golang app I have written, so please take all code with a small grain of salt.

Note that the the status is only recorded from the "Bethesda API", so this will only be of official downtimes (i.e. we are not actually checking the actual status of game servers here..).
An unknown status means that we where unable to get the status of the servers, or Bethesda themselves are reporting unknown status.
The online/offline status is updated every minute, and we report the uptime history for the past 6 months.
A day is considered to have "downtime" if there are more than 15 minutes of cumulative downtime, which we consider to be enough to disrupt a normal gaming session.



## Install Guide
* Download golang binaries
    * https://golang.org/dl/
    * Add go bin to `PATH=C:\Go\bin`
    * Set `GOPATH=C:\Users\Patrick\Go Workspace`
* Install gcc toolchain
    * Need this to build sqlite3
    * http://tdm-gcc.tdragon.net/
    * Install TDM64 package
    * Add bin to `PATH=C:\TDM-GCC-64\bin`
* Install golang package dependence
    * `go get github.com/mattn/go-sqlite3`
    * `go get github.com/tcnksm/go-httpstat`
    * `github.com/ararog/timeago`
* Clone this repo into the `src` folder of go workspace



## Cross Compiling (THIS DOES NOT WORK!!@#@!#)
* Install needed linux dependencies
    * `sudo apt-get update`
    * `sudo apt-get install sqlite3 libsqlite3-dev`
* We will cross build to `linux`
* Use `go env` to see your settings
* Use the provided `script_cc2linux.bat` script
    * This sets and environmental flag to linux
    * It will then call the go build command
    * The `fallout76_ss` is the generated binary file
* Upload the needed files to the server
    * `fallout76_ss`
    * `web_static/`
    * `web_tpl/`
* Ensure permissions are set to allow execution
* FUTURE: this does not seem to work???
* FUTURE: we need a linux compiler?? not sure..

