#!/bin/bash

export DALINK_GO_CONFIG_PATH=/data/manage_portal/server/prod.yml

PROCESS_NAME="./walletus-api"
LOG_FILE="output.log"

PID=$(pgrep -f $PROCESS_NAME)

if [ -n "$PID" ]; then
  echo "Killing existing process $PROCESS_NAME with PID $PID"
  kill -9 $PID
  sleep 2
else
  echo "No existing process found for $PROCESS_NAME"
fi

echo "Starting new process $PROCESS_NAME"
nohup ./$PROCESS_NAME > $LOG_FILE 2>&1 &

NEW_PID=$(pgrep -f $PROCESS_NAME)
if [ -n "$NEW_PID" ]; then
  echo "New process $PROCESS_NAME started with PID $NEW_PID"
else
  echo "Failed to start new process $PROCESS_NAME"
fi
