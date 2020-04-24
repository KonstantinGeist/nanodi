## NanoDI - a tiny dependency injection package

Just another DI for Go.

Advantages:
* Simplicity
* Flexibility
* Designed with DDD in mind
* No special DSL's, reflection or complex code generation

Drawbacks:
* Lacks features from other DI's
* Requires some handwritten boilerplate
* Runtime errors in case dependencies cannot be found or types don't match
* Really just hacked together in a night

### Installation

$ go get -u github.com/KonstantinGeist/nanodi

### Basic usage

The main concept in NanoDI is a `builder`. A builder describes how to create a dependency, given its name:

```
func MapQueryServiceBuilder() nanodi.Builder {
	return nanodi.NewBuilder(names.MapQueryService, func(provider nanodi.Provider) (interface{}, error) {
		worldAPI := provider.GetService(worldnames.API).(worldapi.API)
		mapAPI := provider.GetService(mapnames.API).(mapapi.API)

		return (app.MapQueryService)(&mapQueryService{worldAPI: worldAPI, mapAPI: mapAPI}), nil
	})
}
```

Here, the builder describes how to build a depedency named `names.MapQueryService`, which itself depends on two other dependencies, `worldnames.API` and `mapnames.API` (these are exported names, see below). A builder should be placed alongside with the implementation of the entity/service it builds (for example, in the infrastructure layer).

Each bounded context (or submodule etc.) of your project should export a list of all the builders defined in the context:

```
func Builders() []nanodi.Builder {
	return []nanodi.Builder{
		command.BusBuilder(),
		event.DispatcherBuilder(),
		ui.WindowBuilder(),
	}
}
```

Each bounded context should also export a list of dependency names, so that other contexts could refer to it:

```
const (
	CommandBus     = "command.Bus"
	CommandHandler = "command.Handler"

	EventDispatcher = "event.Dispatcher"
	EventHandler    = "event.Handler"
)
```

Builders and names should be placed in a special layer, for example I call it "build layer". It is allowed for bounded contexts to refer to names defined in build layers of other bounded contexts from inside anticorruption layers.

In the entry point of your application (main function, API entry point, etc.), you take builders from all known bounded contexts and combine them together:

```
func builders() []nanodi.Builder {
	builders := [][]nanodi.Builder{
		fxbuilders.Builders(),
		worldbuilders.Builders(), // builders from the bounded context "world"
		mapbuilders.Builders(),
		uibuilders.Builders(),
		{
			ConfigBuilder(),
		},
	}

	return nanodi.CombineBuilders(builders...)
}
```

and then all you have to do is assemble them all together and retrieve the root of the dependency tree:

```
func main() {
	assembly := nanodi.Assemble(builders())
	window := nanodi.GetService(fxnames.Window).(ui.Window)
	window.Show()
}
```

## Tips & tricks

* By default, services are cached, i.e. several requests to create a dependency by the same name will produce the same object. To make a builder generate a new object each time it is requested, use `NewBuilderWithOptions` with `BuilderOptions.IsShared=false`
* Several builders can implement the same interface (same name). If you call `provider.GetService`, it will panic, because of ambiguouity. However, you can call `provider.GetServices`, which will return a list of all builders with the given name. This is useful, for example, if we want a command bus to find and register all command handlers. However, the current implementation always returns non-cached objects for such a case.
* Detects dependency cycles (and panics).
* Dependency cycles can be resolved by lazy loading. Lazy loading is not implemented by default, but can easily be emulated by caching the provider instance in the builder and retrieving required dependencies later on first access:

```
func DispatcherBuilder() nanodi.Builder {
	return nanodi.NewBuilder(names.EventDispatcher, func(provider nanodi.Provider) (interface{}, error) {
		handlers := make(map[string][]Handler)
		return (Dispatcher)(&dispatcher{handlers: handlers, provider: provider}), nil
	})
}

func (d *dispatcher) Dispatch(event Event) error {
	d.lazyLoadIfRequired()
	// ...
}

func (d *dispatcher) lazyLoadIfRequired() {
	if d.isLazyLoaded {
		return
	}

	d.handlers = d.provider.GetServices(names.EventHandler)
	d.isLazyLoaded = true
}
```

* It's also often useful to configure a dependency externally in a file. It's also simple, you can create a `Config` object, describe a builder for it, and inject it in the builder of your entity/service.
* If you are being creative, you can combine different assemblies, for example, create a seperate global assembly for singleton services, and refer to them from inside newly generated request-scoped assemblies.
