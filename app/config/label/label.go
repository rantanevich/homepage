package label

import "github.com/traefik/paerser/parser"

func Decode(labels map[string]string, element any, filters ...string) error {
	return parser.Decode(labels, element, "HOMEPAGE_", filters...)
}
