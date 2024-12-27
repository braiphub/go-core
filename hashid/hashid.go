package hashid

//go:generate mockgen -destination=hashid_mock.go -package=hashid . Hasher

type Hasher interface {
	WithPrefix(prefix string) Hasher
	Generate(id uint) (string, error)
	Decode(hash string) (uint, error)
}
