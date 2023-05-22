package tree

type Attrs map[string]interface{}

func (this *Attrs) Copy() Attrs {
	copy := Attrs{}
	for k, v := range *this {
		copy[k] = v
	}
	return copy
}
