from .interfaces import IResource
from .network_manager import NetworkManager

class PipelineStepResource(IResource):
    def __init__(self,environment,state):
        self.environment = environment
        self.state = state
        self.enabled = self.environment["ENABLED"] == "true"
        secret = self.environment["ENCRYPTION_KEY"]
        fqdn = self.environment["DEPLOY_ENDPOINT"]
        self.network_manager = NetworkManager(secret,fqdn,'/pipelinestep')

    def make_payload(self,method):
        data = None
        if method == 'POST':
            name = self.environment['NAME']
            description = self.environment['DESCRIPTION']
            pipeline_id = self.environment['PIPELINE_ID']
            workspace_id = self.environment['WORKSPACE_ID']
            next_step = self.environment['NEXT'].split(',')
            if next_step[0] == "":
                next_step = []
            approvers = self.environment['APPROVERS'].split(',')
            if approvers[0] == "":
                approvers = []
            data = {'name':name,'description':description,'pipelineId':pipeline_id,'workspaceId':workspace_id,'next':next_step,'approvers':approvers}
        elif method == 'PATCH':
            name = self.environment['NAME']
            description = self.environment['DESCRIPTION']
            pipeline_id = self.environment['PIPELINE_ID']
            workspace_id = self.environment['WORKSPACE_ID']
            next_step = self.environment['NEXT'].split(',')
            approvers = self.environment['APPROVERS'].split(',')
            id = self.state['id']
            data = {'name':name,'description':description,'pipelineId':pipeline_id,'workspaceId':workspace_id,'next':next_step,'approvers':approvers, 'id':id}
        else:
            data = self.state
        return data

    def create(self):
        if not self.enabled:
            payload="{}"
            payload["id"]="empty"
            self.state=payload
            return
        print("python::create")
        payload = self.make_payload('POST')
        r = self.network_manager.post(payload)
        print(r)
        j = r.json()
        print(j)
        id = j['id']
        payload['id']=id
        self.state=payload
    
    def read(self):
        if not self.enabled:
            payload = "{}"
            payload["id"]="empty"
            self.state=payload
            return
        print("python::read")
        id = self.state['id']
        path = '/'+id
        r = self.network_manager.get(path=path)
        j = r.json()
        j.pop('creationTimestamp',None)
        print(j)
        self.state = j

    def update(self):
        if not self.enabled:
            payload = {}
            payload["id"]="empty"
            self.state=payload
            return
        print("python::update")
        payload = self.make_payload('PATCH')
        r = self.network_manager.patch(payload)
        j = r.json()
        j.pop('creationTimestamp',None)
        print(j)
        self.state = j
    
    def delete(self):
        if not self.enabled:
            payload = "{}"
            self.state=payload
            return
        print("python::delete")
        payload = self.make_payload('DELETE')
        r = self.network_manager.delete(payload)
        print(r)
        self.state={}
