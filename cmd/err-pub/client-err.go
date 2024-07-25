package main

type clientErr struct {
	s string
}

func (c *clientErr) Error() string {
	return c.s
}

func (c *clientErr) Is(err error) bool {
	_, is := err.(*clientErr)
	return is
}
