#!/bin/sh

SUCCEDED=$(curl -H "X-Gitlab-Token:$DEPLOY_TOKEN" -X POST $DEPLOY_STAGING_URL | grep "Done!" | wc -l)

if [ "$SUCCEDED" != 1 ] ; then
  echo "Deployment failed!"
  exit 1
fi

echo "Deployment succeeded"
