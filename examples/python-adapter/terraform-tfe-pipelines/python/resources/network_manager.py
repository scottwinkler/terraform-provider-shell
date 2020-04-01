import jwt, requests,sys, json
from urllib.parse import urlencode

class NetworkManager:
    def __init__(self,secret,fqdn,resource):
        self.fqdn = fqdn
        self.resource = resource
        jwt = self.make_jwt(secret,fqdn)
        self.jwt = jwt
        self.session = self.initialize_session(jwt)

    def make_jwt(self,secret,fqdn):
        #could claim anything here, doesn't matter
        claims = {'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name':['deploy'],'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress':['deploy-dev@elliemae.com'],'sub':'deploy','iss':fqdn}
        encoded = jwt.encode(claims,secret,algorithm='HS256')
        return encoded.decode('utf-8')

    def initialize_session(self,jwt):
        s = requests.Session()
        cookiejar = requests.cookies.cookiejar_from_dict({'jwt': jwt})
        s.cookies.update(cookiejar)
        s.headers.update({'Content-Type':'application/json'})
        return s

    def get_url(self):
        return 'https://'+self.fqdn+'/api'+self.resource

    def post(self,payload):
        url = self.get_url()
        return self.session.post(url,json=payload)

    def get(self,path="",params={}):
        url = self.get_url()
        url+=path
        if len(params)>0:
            querystring = urlencode(params)
            url += '?'+querystring
        return self.session.get(url)
    
    def patch(self,payload):
        url = self.get_url()
        return self.session.patch(url,json=payload)

    def delete(self,payload):
        url = self.get_url()
        return self.session.delete(url,json=payload)
