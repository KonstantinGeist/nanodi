package nanodi

import (
	"testing"
)

const domainValue1 = 2
const domainValue2 = 3

func TestAssemble(t *testing.T) {
	builders := []Builder{
		NewBuilder(controllerName, func(provider Provider) (interface{}, error) {
			appService := provider.GetService(appServiceName).(appService)
			return (controller)(&controllerImpl{appService: appService}), nil
		}),
		NewBuilder(appServiceName, func(provider Provider) (interface{}, error) {
			domainService := provider.GetService(domainServiceName).(domainService)
			return (appService)(&appServiceImpl{domainService: domainService}), nil
		}),
		NewBuilder(domainServiceName, func(provider Provider) (interface{}, error) {
			domainObjectA := provider.GetService(domainObjectAName).(domainObjectA)
			domainObjectB := provider.GetService(domainObjectBName).(domainObjectB)
			return (domainService)(&domainServiceImpl{domainObjectA: domainObjectA, domainObjectB: domainObjectB}), nil
		}),
		NewBuilder(domainObjectAName, func(Provider) (interface{}, error) {
			return (domainObjectA)(&domainObjectAImpl{}), nil
		}),
		NewBuilder(domainObjectBName, func(Provider) (interface{}, error) {
			return (domainObjectB)(&domainObjectBImpl{}), nil
		}),
	}

	assembly := Assemble(builders)
	root := assembly.GetService(controllerName).(controller)

	want := 5
	got := root.DomainValue()
	if got != want {
		t.Errorf("got %d; want %d", got, want)
	}
}

func TestDependencyCollection(t *testing.T) {
	builders := []Builder{
		NewBuilder(collectionUserName, func(provider Provider) (interface{}, error) {
			untypedItems := provider.GetServices(collectionItemName)

			typedItems := make([]collectionItem, len(untypedItems))
			for i, untypedItem := range untypedItems {
				typedItems[i] = untypedItem.(collectionItem)
			}

			return (collectionUser)(&collectionUserImpl{items: typedItems}), nil
		}),
		NewBuilder(collectionItemName, func(Provider) (interface{}, error) {
			return &collectionItemImpl{domainValue: domainValue1}, nil
		}),
		NewBuilder(collectionItemName, func(Provider) (interface{}, error) {
			return &collectionItemImpl{domainValue: domainValue2}, nil
		}),
	}

	assembly := Assemble(builders)
	root := assembly.GetService(collectionUserName).(collectionUser)

	want := 5
	got := root.DomainValue()
	if got != want {
		t.Errorf("got %d; want %d", got, want)
	}
}

const (
	controllerName    = "controller"
	appServiceName    = "appService"
	domainServiceName = "domainService"
	domainObjectAName = "domainObjectA"
	domainObjectBName = "domainObjectB"

	collectionItemName = "collectionItem"
	collectionUserName = "collectionUser"
)

type domainObjectA interface {
	DomainValue() int
}

type domainObjectB interface {
	DomainValue() int
}

type domainService interface {
	DomainValue() int
}

type appService interface {
	DomainValue() int
}

type controller interface {
	DomainValue() int
}

type controllerImpl struct {
	appService appService
}

type appServiceImpl struct {
	domainService domainService
}

type domainServiceImpl struct {
	domainObjectA domainObjectA
	domainObjectB domainObjectB
}

type domainObjectAImpl struct{}
type domainObjectBImpl struct{}

type collectionItem interface {
	DomainValue() int
}

type collectionUser interface {
	DomainValue() int
}

type collectionItemImpl struct {
	domainValue int
}

type collectionUserImpl struct {
	items []collectionItem
}

func (i *controllerImpl) DomainValue() int {
	return i.appService.DomainValue()
}

func (i *appServiceImpl) DomainValue() int {
	return i.domainService.DomainValue()
}

func (i *domainServiceImpl) DomainValue() int {
	return i.domainObjectA.DomainValue() + i.domainObjectB.DomainValue()
}

func (i *domainObjectAImpl) DomainValue() int {
	return domainValue1
}

func (i *domainObjectBImpl) DomainValue() int {
	return domainValue2
}

func (i *collectionItemImpl) DomainValue() int {
	return i.domainValue
}

func (u *collectionUserImpl) DomainValue() int {
	sum := 0
	for _, item := range u.items {
		sum += item.DomainValue()
	}
	return sum
}
