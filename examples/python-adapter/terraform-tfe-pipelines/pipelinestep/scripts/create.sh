#!/bin/bash
pip3 install -r ./modules/project/terraform-tfe-pipelines/python/requirements.txt > /dev/null
python3 ./modules/project/terraform-tfe-pipelines/python/main.py --name=PipelineStepResource --module=pipelinestep_resource --command=create --state=${STATE_FILE}
cat ${STATE_FILE}
rm ${STATE_FILE}
