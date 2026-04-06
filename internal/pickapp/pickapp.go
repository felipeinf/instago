package pickapp

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"math/rand"

	"github.com/felipeinf/instago/config"
)

func PickBySeed(seed string, rng *rand.Rand) config.AppProfile {
	list := config.AppProfileList()
	if len(list) == 0 {
		panic("pickapp: no app profiles")
	}
	if seed != "" {
		h := sha256.Sum256([]byte(seed))
		digest := hex.EncodeToString(h[:])
		bi := new(big.Int)
		bi.SetString(digest, 16)
		m := big.NewInt(int64(len(list)))
		idx := int(new(big.Int).Mod(bi, m).Int64())
		return list[idx]
	}
	if rng != nil {
		return list[rng.Intn(len(list))]
	}
	return list[rand.Intn(len(list))]
}

func MatchStored(appVersion string) (config.AppProfile, bool) {
	p, ok := config.AppProfiles[appVersion]
	return p, ok
}
