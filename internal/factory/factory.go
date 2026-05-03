package factory

import (
	"sync/atomic"

	"github.com/brianvoe/gofakeit/v7"
	"golang.org/x/crypto/bcrypt"
)

var (
	fake                  *gofakeit.Faker
	DefaultHashedPassword string
	counter               uint64
)

func init() {
	fake = gofakeit.New(42)

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		panic("factory: hashing default password: " + err.Error())
	}
	DefaultHashedPassword = string(hash)
}

func nextID() uint64 {
	return atomic.AddUint64(&counter, 1)
}
