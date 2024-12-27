package hashids

import (
	"github.com/braiphub/go-core/hashid"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/speps/go-hashids/v2"
)

type HashIDsAdapter struct {
	prefix string
	hashID *hashids.HashID
}

func New(prefix, salt string, minLen int) (*HashIDsAdapter, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLen

	hashID, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, errors.Wrap(err, "init hashids")
	}

	return &HashIDsAdapter{
		prefix: prefix,
		hashID: hashID,
	}, nil
}

func (adapter *HashIDsAdapter) WithPrefix(prefix string) hashid.Hasher {
	newInstance := *adapter
	newInstance.prefix = prefix

	return &newInstance
}

func (adapter *HashIDsAdapter) Generate(id uint) (string, error) {
	hash, err := adapter.hashID.Encode([]int{int(id)})
	if err != nil {
		return "", errors.Wrap(err, "hash create for id["+strconv.Itoa(int(id))+"]")
	}

	return adapter.prefix + hash, nil
}

func (adapter *HashIDsAdapter) Decode(hash string) (uint, error) {
	hash = strings.TrimPrefix(hash, adapter.prefix)
	ids, err := adapter.hashID.DecodeWithError(hash)
	if err != nil {
		return 0, errors.Wrap(err, "id decode for hash "+hash)
	}
	if len(ids) != 1 {
		return 0, errors.Wrap(err, "len "+strconv.Itoa(len(ids))+"invalid; hash "+hash)
	}

	return uint(ids[0]), nil
}
