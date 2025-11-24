package srt

// Options represents SRT socket options.
type Options map[string]string

// Clone returns a shallow copy of the Options map.
func (o Options) Clone() Options {
	if o == nil {
		return nil
	}
	cp := make(Options, len(o))
	for k, v := range o {
		cp[k] = v
	}
	return cp
}
