package zkp

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

type SecureAuthCircuit struct {
	Secret       frontend.Variable `gnark:",private"`
	Address      frontend.Variable `gnark:",public"`
	Challenge    frontend.Variable `gnark:",public"`
	ExpectedHash frontend.Variable `gnark:",public"`
}

func (c *SecureAuthCircuit) Define(api frontend.API) error {
	mimcHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}
	
	mimcHasher.Write(c.Secret)
	mimcHasher.Write(c.Address)
	mimcHasher.Write(c.Challenge)
	computedHash := mimcHasher.Sum()
	
	api.AssertIsEqual(computedHash, c.ExpectedHash)
	return nil
}
