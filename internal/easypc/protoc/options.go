package protoc

// Option configures protoc compilation.
type Option func(*compileConfig)

type compileConfig struct {
	resolvers         []Resolver
	importPaths       []string
	includeImports    bool
	includeSourceInfo bool
	standardImports   bool
	concurrency       int
	errorHandler      func(Diagnostic)
	variant           string
	strategy          string
	syntaxTrees       bool
}

// WithSyntaxTrees makes Link keep every file's syntax tree and tokens
// (LinkedFile.Syntax). It has no effect on Compile.
func WithSyntaxTrees(keep bool) Option {
	return func(c *compileConfig) {
		c.syntaxTrees = keep
	}
}

// WithResolver appends a resolver to the compiler's resolver chain.
func WithResolver(r Resolver) Option {
	return func(c *compileConfig) {
		if r != nil {
			c.resolvers = append(c.resolvers, r)
		}
	}
}

// WithImportPaths adds directory search paths for imports.
func WithImportPaths(paths ...string) Option {
	return func(c *compileConfig) {
		c.importPaths = append(c.importPaths, paths...)
	}
}

// WithIncludeImports sets whether dependencies are included in the serialized output.
func WithIncludeImports(include bool) Option {
	return func(c *compileConfig) {
		c.includeImports = include
	}
}

// WithIncludeSourceInfo sets whether source code info locations and comments are generated.
func WithIncludeSourceInfo(include bool) Option {
	return func(c *compileConfig) {
		c.includeSourceInfo = include
	}
}

// WithStandardImports controls whether standard protobuf Well-Known Types
// (e.g., google/protobuf/timestamp.proto, descriptor.proto) are automatically resolvable.
func WithStandardImports(enable bool) Option {
	return func(c *compileConfig) {
		c.standardImports = enable
	}
}

// WithConcurrency sets the maximum concurrency (number of worker threads).
func WithConcurrency(n int) Option {
	return func(c *compileConfig) {
		c.concurrency = n
	}
}

// WithErrorHandler provides a callback invoked for each error or warning diagnostic.
func WithErrorHandler(handler func(Diagnostic)) Option {
	return func(c *compileConfig) {
		c.errorHandler = handler
	}
}

// WithVariant specifies the lexer variant (auto, scalar, swar, neon, sve2, avx2, avx512).
func WithVariant(variant string) Option {
	return func(c *compileConfig) {
		c.variant = variant
	}
}

// WithStrategy specifies the compilation strategy (auto, a, b, c).
func WithStrategy(strategy string) Option {
	return func(c *compileConfig) {
		c.strategy = strategy
	}
}
