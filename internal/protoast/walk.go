package protoast

// Visitor receives declarations at every nesting level. Nil callbacks are
// skipped. Fields include map and group fields, oneof members and extension
// fields declared in messages or at file level; synthetic map-entry messages
// and their key and value fields are not visited.
type Visitor struct {
	Message   func(*Message)
	Enum      func(*Enum)
	EnumValue func(*Enum, *EnumValue)
	Field     func(*Field)
	Oneof     func(*Oneof)
	Service   func(*Service)
	RPC       func(*Service, *RPC)
}

// Walk visits the file's declarations in source order within each scope.
func (f *File) Walk(v Visitor) {
	for _, m := range f.Messages {
		v.message(m)
	}
	for _, e := range f.Enums {
		v.enum(e)
	}
	for _, e := range f.Extends {
		v.fields(e.Fields)
	}
	for _, s := range f.Services {
		if v.Service != nil {
			v.Service(s)
		}
		if v.RPC != nil {
			for _, r := range s.RPCs {
				v.RPC(s, r)
			}
		}
	}
}

func (v Visitor) message(m *Message) {
	if v.Message != nil {
		v.Message(m)
	}
	v.fields(m.Fields)
	v.fields(m.Maps)
	v.fields(m.Groups)
	for _, o := range m.Oneofs {
		if v.Oneof != nil {
			v.Oneof(o)
		}
		v.fields(o.Fields)
	}
	for _, e := range m.Extends {
		v.fields(e.Fields)
	}
	for _, e := range m.Enums {
		v.enum(e)
	}
	for _, nested := range m.Messages {
		v.message(nested)
	}
}

func (v Visitor) enum(e *Enum) {
	if v.Enum != nil {
		v.Enum(e)
	}
	if v.EnumValue != nil {
		for _, value := range e.Values {
			v.EnumValue(e, value)
		}
	}
}

func (v Visitor) fields(fields []*Field) {
	if v.Field == nil {
		return
	}
	for _, f := range fields {
		v.Field(f)
	}
}
