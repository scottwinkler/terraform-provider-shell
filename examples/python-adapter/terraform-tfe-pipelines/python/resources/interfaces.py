from abc import ABCMeta, abstractmethod

class IDataResource():
    __metaclass__ = ABCMeta

    @abstractmethod
    def read(self): raise NotImplementedError

class IResource():
    __metaclass__ = ABCMeta

    @abstractmethod
    def create(self): raise NotImplementedError
    
    @abstractmethod
    def read(self): raise NotImplementedError

    @abstractmethod
    def update(self): raise NotImplementedError

    @abstractmethod
    def delete(self): raise NotImplementedError