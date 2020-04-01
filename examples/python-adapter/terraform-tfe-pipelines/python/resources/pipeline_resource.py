from .interfaces import IResource
from .network_manager import NetworkManager

class PipelineResource(IResource):
    def __init__(self,environment,state):
        self.environment = environment
        self.state = state
        secret = self.environment["ENCRYPTION_KEY"]
        fqdn = self.environment["DEPLOY_ENDPOINT"]
        self.network_manager = NetworkManager(secret,fqdn,'/pipeline')

    def make_payload(self,method):
        data = None
        if method == 'POST':
            name = self.environment['NAME']
            description = self.environment['DESCRIPTION']
            organization = self.environment['ORGANIZATION']
            data = {'name':name,'description':description,'organization':organization}
        elif method == 'PATCH':
            name = self.environment['NAME']
            description = self.environment['DESCRIPTION']
            organization = self.environment['ORGANIZATION']
            id = self.state['id']
            data = {'name':name,'description':description,'organization':organization,'id':id}
        elif method == 'DELETE':
            data = self.state
        return data

    def create(self):
        print("python::create")
        payload = self.make_payload('POST')
        print(payload)
        r = self.network_manager.post(payload)
        print(r)
        j = r.json()
        print(j)
        id = j['id']
        payload['id']=id
        self.state=payload
    
    def read(self):
        print("python::read")
        id = self.state['id']
        path = '/'+id
        r = self.network_manager.get(path=path)
        j = r.json()
        j.pop('creationTimestamp',None)
        print(j)
        self.state = j

    def update(self):
        print("python::update")
        payload = self.make_payload('PATCH')
        r = self.network_manager.patch(payload)
        j = r.json()
        j.pop('creationTimestamp',None)
        print(j)
        self.state = j
    
    def delete(self):
        print("python::delete")
        payload = self.make_payload('DELETE')
        r = self.network_manager.delete(payload)
        print(r)
        self.state={}
