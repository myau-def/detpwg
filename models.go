package detpwg

import (
	"fmt"
	"slices"
)

var (
	Digits   = NewCharSet("digits", "0123456789", nil)
	Lowers   = NewCharSet("lowercase", "abcdefghijklmnopqrstuvwxyz", nil)
	Uppers   = NewCharSet("uppercase", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", nil)
	Specials = NewCharSet("specials", "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", nil)

	DefaultAlphabet = Alphabet{Digits, Lowers, Uppers, Specials}
)

type MasterKey struct {
	Key []byte
}

type Login struct {
	Service string
	Login   string
	Info    string
	Counter int
}

type AlphabetConfig struct {
	Exclude          string
	IncludeLowercase bool
	IncludeUppercase bool
	IncludeDigits    bool
	IncludeSpecial   bool
}

type Alphabet []CharSet

type CharSet struct {
	Name string
	RuneSet
}

type RuneSet map[rune]struct{}

func NewMasterKey(master, password string) (MasterKey, error) {
	key, err := GenerateMasterKey(master, password)
	if err != nil {
		return MasterKey{}, err
	}
	return MasterKey{Key: key}, nil
}

func (mk *MasterKey) GeneratePassword(login *Login, alphabet Alphabet, length int) (string, error) {
	return GeneratePassword(&mk.Key, login, alphabet, length)
}

func (l *Login) Salt() []byte {
	return []byte(fmt.Sprintf("%s\n%s\n%s\n%d", l.Service, l.Info, l.Login, l.Counter))
}

func NewLogin(service, login string) Login {
	return Login{
		Service: service,
		Login:   login,
		Info:    "account on service",
		Counter: 1,
	}
}

func NewRuneSet(chars string, exclude RuneSet) RuneSet {
	set := make(RuneSet, len(chars))
	for _, char := range chars {
		if _, ok := exclude[char]; !ok {
			set[char] = struct{}{}
		}
	}
	return set
}

func (set *RuneSet) Exclude(rs *RuneSet) {
	for char := range *rs {
		if _, ok := (*set)[char]; ok {
			delete(*set, char)
		}
	}
}

func (set *RuneSet) Clone() RuneSet {
	clone := make(RuneSet, len(*set))
	for char := range *set {
		clone[char] = struct{}{}
	}
	return clone
}

func (set *RuneSet) Pool() []rune {
	runes := make([]rune, 0, len(*set))
	for char := range *set {
		runes = append(runes, char)
	}
	slices.Sort(runes)
	return runes
}

func (set *RuneSet) String() string {
	return string(set.Pool())
}

func NewCharSet(name, chars string, exclude RuneSet) CharSet {
	return CharSet{
		Name:    name,
		RuneSet: NewRuneSet(chars, exclude),
	}
}

func (set *CharSet) Clone() CharSet {
	return CharSet{
		Name:    set.Name,
		RuneSet: set.RuneSet.Clone(),
	}
}

func (ac *AlphabetConfig) BuildAlphabet() (Alphabet, error) {
	alphabet := make(Alphabet, 0)
	excluded := NewRuneSet(ac.Exclude, nil)

	if ac.IncludeDigits {
		digits := Digits.Clone()
		digits.Exclude(&excluded)
		alphabet = append(alphabet, digits)
	}

	if ac.IncludeLowercase {
		lowers := Lowers.Clone()
		lowers.Exclude(&excluded)
		alphabet = append(alphabet, lowers)
	}

	if ac.IncludeUppercase {
		uppers := Uppers.Clone()
		uppers.Exclude(&excluded)
		alphabet = append(alphabet, uppers)
	}

	if ac.IncludeSpecial {
		specials := Specials.Clone()
		specials.Exclude(&excluded)
		alphabet = append(alphabet, specials)
	}

	return alphabet, nil
}

func NewAlphabet(
	exclude string,
	includeLowercase bool,
	includeUppercase bool,
	includeDigits bool,
	includeSpecial bool,
) (Alphabet, error) {
	cfg := AlphabetConfig{
		Exclude:          exclude,
		IncludeLowercase: includeLowercase,
		IncludeUppercase: includeUppercase,
		IncludeDigits:    includeDigits,
		IncludeSpecial:   includeSpecial,
	}
	return cfg.BuildAlphabet()
}

func (a *Alphabet) Clone() Alphabet {
	clone := make(Alphabet, 0, len(*a))
	for _, set := range *a {
		clone = append(clone, set.Clone())
	}
	return clone
}

func (a *Alphabet) PoolSize() int {
	length := 0
	for _, set := range *a {
		length += len(set.RuneSet)
	}
	return length
}

func (a *Alphabet) ValidateAndPool() (RuneSet, error) {
	if len(*a) == 0 {
		return nil, fmt.Errorf("no charsets in alphabet")
	}

	pool := make(RuneSet, a.PoolSize())
	for _, set := range *a {
		if len(set.RuneSet) == 0 {
			return nil, fmt.Errorf("no chars in charset %s", set.Name)
		}

		for r := range set.RuneSet {
			if _, ok := pool[r]; ok {
				return nil, fmt.Errorf("intersects char %c in charset %s", r, set.Name)
			}
			pool[r] = struct{}{}
		}
	}

	return pool, nil
}

func SwapAndPop[S interface{ ~[]T }, T any](s S, i int) (S, T) {
	l := len(s) - 1
	if l < 0 {
		panic("pop from empty slice")
	}

	if i != l {
		s[i], s[l] = s[l], s[i]
	}
	return s[:l], s[l]
}

func SwapToBack[S interface{ ~[]T }, T any](s S, i int, e T) S {
	if i < 0 || i > len(s) {
		panic("index out of range")
	}
	if i == len(s) {
		return append(s, e)
	}
	s = append(s, s[i])
	s[i] = e
	return s
}
