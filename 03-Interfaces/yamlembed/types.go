package yamlembed

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type Foo struct {
	A string `yaml:"aa"`
	p int64
}

type Bar struct {
	I      int64    `yaml:"-"`
	B      string   `yaml:"b"`
	UpperB string   `yaml:"-"`
	OI     []string `yaml:"oi,omitempty"`
	F      []any    `yaml:"f,flow"`
}

type Baz struct {
	Foo `yaml:",inline"`
	Bar `yaml:",inline"`
}

func (b *Bar) UnmarshalYAML(value *yaml.Node) error {
	type plainBar Bar

	if err := value.Decode((*plainBar)(b)); err != nil {
		return err
	}

	b.UpperB = strings.ToUpper(b.B)

	return nil
}

func (b *Baz) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&b.Foo); err != nil {
		return err
	}

	return value.Decode(&b.Bar)
}

func (b Baz) MarshalYAML() (interface{}, error) {
	type plainBar Bar

	return struct {
		Foo `yaml:",inline"`
		Bar plainBar `yaml:",inline"`
	}{
		Foo: b.Foo,
		Bar: plainBar(b.Bar),
	}, nil
}
