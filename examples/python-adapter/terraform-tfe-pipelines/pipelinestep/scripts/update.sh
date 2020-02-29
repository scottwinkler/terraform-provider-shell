#!/bin/bash
state=$(cat)
echo $state > ${STATE_FILE}
pip3 install -r ./modules/project/terraform-tfe-pipelines/python/requirements.txt > /dev/null
python3 ./modules/project/terraform-tfe-pipelines/python/main.py  --name=PipelineStepResource --module=pipelinestep_resource --command=update --state=${STATE_FILE}
rm ${STATE_FILE}