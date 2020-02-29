package resources

var (
	DataResourceMap = map[string]DataResourceConstructor{
		"organizations": NewTFEOrganizations,
	}
	ResourceMap = map[string]ResourceConstructor{
		"registry-module": NewTFERegistryModule,
	}
)

type DataResourceConstructor func(map[string]string) IDataResource

type ResourceConstructor func(map[string]string, map[string]string) IResource
