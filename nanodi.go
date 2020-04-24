package nanodi

import "strings"

type Provider interface {
	GetService(name string) interface{}
	GetServices(name string) []interface{}
}

type Assembly interface {
	GetService(name string) interface{}
}

type Builder interface {
	CommandName() string
	Options() BuilderOptions
	Build(provider Provider) (interface{}, error)
}

type BuilderOptions struct {
	IsShared bool
}

type BuilderFn func(provider Provider) (interface{}, error)

func NewBuilder(name string, fn BuilderFn) Builder {
	defaultOptions := BuilderOptions{
		IsShared: true,
	}
	return &builder{name: name, fn: fn, options: defaultOptions}
}

func NewBuilderWithOptions(name string, fn BuilderFn, options BuilderOptions) Builder {
	return &builder{name: name, fn: fn, options: options}
}

func Assemble(builders []Builder) Assembly {
	return newBuildContext(builders)
}

func CombineBuilders(builders ...[]Builder) []Builder {
	count := 0
	for _, outer := range builders {
		count += len(outer)
	}

	combined := make([]Builder, count)
	index := 0
	for _, outer := range builders {
		for _, inner := range outer {
			combined[index] = inner
			index++
		}
	}
	return combined
}

type builder struct {
	name    string
	fn      BuilderFn
	options BuilderOptions
}

type buildContext struct {
	builders map[string][]Builder
	built    map[string]interface{}

	// Lists the current dependency stack to detect dependency cycles.
	buildStack []string
}

func (b *builder) CommandName() string {
	return b.name
}

func (b *builder) Options() BuilderOptions {
	return b.options
}

func (b *builder) Build(provider Provider) (interface{}, error) {
	return b.fn(provider)
}

func newBuildContext(builders []Builder) *buildContext {
	builderMap := make(map[string][]Builder)
	for _, builder := range builders {
		name := builder.CommandName()
		builderMap[name] = append(builderMap[name], builder)
	}

	built := make(map[string]interface{})

	return &buildContext{
		builders: builderMap,
		built:    built,
	}
}

func (c *buildContext) GetService(name string) interface{} {
	if c.isInBuildStack(name) {
		c.panicCircular(name)
	}

	c.buildStack = append(c.buildStack, name)
	defer func() {
		c.buildStack = c.buildStack[:len(c.buildStack)-1]
	}()

	built, ok := c.built[name]
	if ok {
		return built
	}

	builders, ok := c.builders[name]
	if !ok {
		panic("builder does not exist for the type")
	}
	if len(builders) > 1 {
		panic("ambiguous: several builders for the type")
	}
	builder := builders[0]

	built, err := builder.Build(c)
	if err != nil {
		panic("failed to build service")
	}

	if builder.Options().IsShared {
		c.built[name] = built
	}
	return built
}

func (c *buildContext) GetServices(name string) []interface{} {
	if c.isInBuildStack(name) {
		c.panicCircular(name)
	}

	c.buildStack = append(c.buildStack, name)
	defer func() {
		c.buildStack = c.buildStack[:len(c.buildStack)-1]
	}()

	built, ok := c.built[name]
	if ok {
		return built.([]interface{})
	}

	builders, ok := c.builders[name]
	if !ok {
		return nil
	}

	result := make([]interface{}, len(builders))

	for i, builder := range builders {
		built, err := builder.Build(c)
		if err != nil {
			panic("failed to build service")
		}

		result[i] = built
	}

	c.built[name] = result
	return result
}

func (c *buildContext) isInBuildStack(name string) bool {
	for _, dep := range c.buildStack {
		if dep == name {
			return true
		}
	}
	return false
}

func (c *buildContext) panicCircular(offendingDep string) {
	strBuilder := strings.Builder{}
	for _, dep := range c.buildStack {
		strBuilder.WriteString(dep)
		strBuilder.WriteString(" > ")
	}
	strBuilder.WriteString(offendingDep)

	panic("circular dependency detected: " + strBuilder.String())
}
