package crypto

import (
	"reflect"
	"sync"

	"squad-maker/utils/env"

	"github.com/matthewhartstonge/argon2"
)

var (
	lock                  = &sync.Mutex{}
	argon2ConfigSingleton argon2.Config
)

func GenerateHash(password string) (string, error) {
	argon := getDefaultArgon2Config()
	pwd, err := argon.HashEncoded([]byte(password))
	if err != nil {
		return "", err
	}
	return string(pwd), nil
}

func VerifyHash(password string, hash string) bool {
	ok, _ := argon2.VerifyEncoded([]byte(password), []byte(hash))
	return ok
}

func getDefaultArgon2Config() argon2.Config {
	lock.Lock()
	defer lock.Unlock()

	if reflect.ValueOf(argon2ConfigSingleton).IsZero() {
		cfg := argon2.DefaultConfig()

		c, err := env.GetUInt32("ARGON2_HASH_LENGTH")
		if err == nil {
			cfg.HashLength = c
		}

		c, err = env.GetUInt32("ARGON2_SALT_LENGTH")
		if err == nil {
			cfg.SaltLength = c
		}

		c, err = env.GetUInt32("ARGON2_TIME_COST")
		if err == nil {
			cfg.TimeCost = c
		}

		c, err = env.GetUInt32("ARGON2_MEMORY_COST")
		if err == nil {
			cfg.MemoryCost = c
		}

		c2, err := env.GetUInt8("ARGON2_PARALLELISM")
		if err == nil {
			cfg.Parallelism = c2
		}

		c, err = env.GetUInt32("ARGON2_MODE")
		if err == nil {
			cfg.Mode = getArgon2ModeFromUInt32(c, cfg.Mode)
		}

		c, err = env.GetUInt32("ARGON2_VERSION")
		if err == nil {
			cfg.Version = getArgon2VersionFromUInt32(c, cfg.Version)
		}

		argon2ConfigSingleton = cfg
	}
	return argon2ConfigSingleton
}

func getArgon2ModeFromUInt32(i uint32, defaultValue argon2.Mode) argon2.Mode {
	switch i {
	case argon2.ModeArgon2id:
	case argon2.ModeArgon2i:
		return argon2.Mode(i)
	}
	return defaultValue
}

func getArgon2VersionFromUInt32(i uint32, defaultValue argon2.Version) argon2.Version {
	switch i {
	case argon2.Version10:
	case argon2.Version13:
		return argon2.Version(i)
	}
	return defaultValue
}
