#!/bin/bash

cd `dirname $0`

cdir=`pwd`

#execSQL="/home/paas/paas/bin/execSQL.sh"
execSQL="/home/yourchanges/bin/execSQL.sh"

source /home/paas/paas/env.sh

#gc
PIDS=`ps aux | grep "$cdir/fetchError" | grep -v grep | awk '{print $2}'`

for PID in $PIDS ; do
     kill -9 $PID > /dev/null 2>&1
done


echo "`date "+%Y/%m/%d %H:%M:%S"` Starting fetch errors ... "
$cdir/fetchError

echo "`date "+%Y/%m/%d %H:%M:%S"` Starting load into DB ... "
$execSQL mysql 127.0.0.1:3306 platform root 123456 $cdir/load.sql

echo "`date "+%Y/%m/%d %H:%M:%S"` All finished."
cd -