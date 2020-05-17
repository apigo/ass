package ass

import "testing"

func TestEventValidate(t *testing.T) {
	cases := []struct {
		input Event
		valid bool
	}{}

	for _, c := range cases {
		err := c.input.validate()
		if c.valid && err != nil {
			t.Errorf("Expect validate success, got: %v", err)
			continue
		}
		if !c.valid && err == nil {
			t.Errorf("Expect invalid event, but passed")
		}
	}
}
