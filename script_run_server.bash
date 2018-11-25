#!/bin/sh

ps auxw | grep fallout76_ss | grep -v grep > /dev/null

if [ $? = 0 ]
then
	~/go/src/fallout76_ss/fallout76_ss -listen-addr :80 > /dev/null ~/go/src/fallout76_ss/logs/http_log.txt
fi