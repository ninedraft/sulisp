package object

type Env struct {
	parent *Env
	values map[string]Object
}

func NewEnv() *Env {
	return &Env{
		values: map[string]Object{},
	}
}

func (env *Env) Names(yield func(string) bool) {
	for name := range env.values {
		if !yield(name) {
			return
		}
	}
}

func (env *Env) Child() *Env {
	return &Env{
		parent: env,
		values: map[string]Object{},
	}
}

func (env *Env) Assign(name string, value Object) *Env {
	if env.values == nil {
		env.values = map[string]Object{}
	}

	env.values[name] = value

	return env
}

func (env *Env) LookUp(name string) (_ Object, ok bool) {
	if env == nil {
		return Null{}, false
	}

	v, ok := env.values[name]
	if ok {
		return v, ok
	}

	if env.parent != nil {
		v, ok = env.parent.LookUp(name)
	}
	if ok {
		return v, ok
	}

	return Null{}, false
}
