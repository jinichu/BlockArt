package blockartlib

import "testing"

func TestShapeSvgString(t *testing.T) {
	cases := []struct {
		in   Shape
		want string
	}{
		{
			Shape{
				Type:   PATH,
				Svg:    "M 0 0 L 20 20",
				Stroke: "red",
				Fill:   "transparent",
			},
			`<path d="M 0 0 L 20 20" stroke="red" fill="transparent"/>`,
		},
	}

	for i, c := range cases {
		out := c.in.SvgString()
		if out != c.want {
			t.Errorf("%d. %+v.SvgString() = %q; not %q", i, c.in, out, c.want)
		}
	}
}
