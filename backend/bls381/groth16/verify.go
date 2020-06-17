// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark/internal/generators DO NOT EDIT

package groth16

import (
	curve "github.com/consensys/gurvy/bls381"
	"github.com/consensys/gurvy/bls381/fr"

	"github.com/consensys/gnark/backend"
)

// Verify verifies a proof
func Verify(proof *Proof, vk *VerifyingKey, inputs backend.Assignments) (bool, error) {

	c := curve.BLS381()

	var kSum curve.G1Jac
	var eKrsδ, eArBs, eKvkγ curve.PairingResult
	chan1 := make(chan bool, 1)
	chan2 := make(chan bool, 1)

	// e([Krs]1, -[δ]2)
	go func() {
		c.MillerLoop(proof.Krs, vk.G2.DeltaNeg, &eKrsδ)
		chan1 <- true
	}()

	// e([Ar]1, [Bs]2)
	go func() {
		c.MillerLoop(proof.Ar, proof.Bs, &eArBs)
		chan2 <- true
	}()

	kInputs, err := ParsePublicInput(vk.PublicInputs, inputs)
	if err != nil {
		return false, err
	}
	<-kSum.MultiExp(c, vk.G1.K, kInputs)

	// e(Σx.[Kvk(t)]1, -[γ]2)
	var kSumAff curve.G1Affine
	kSum.ToAffineFromJac(&kSumAff)

	c.MillerLoop(kSumAff, vk.G2.GammaNeg, &eKvkγ)

	<-chan1
	<-chan2
	right := c.FinalExponentiation(&eKrsδ, &eArBs, &eKvkγ)
	return vk.E.Equal(&right), nil
}

// parsePublicInput return the ordered public input values
// in regular form (used as scalars for multi exponentiation).
// The function is public because it's needed for the recursive snark.
func ParsePublicInput(expectedNames []string, input backend.Assignments) ([]fr.Element, error) {
	toReturn := make([]fr.Element, len(expectedNames))

	// ensure we don't assign private inputs
	publicInput := input.DiscardSecrets()

	for i := 0; i < len(expectedNames); i++ {
		if expectedNames[i] == backend.OneWire {
			// ONE_WIRE is a reserved name, it should not be set by the user
			toReturn[i].SetOne()
			toReturn[i].FromMont()
		} else {
			if val, ok := publicInput[expectedNames[i]]; ok {
				toReturn[i].SetBigInt(&val.Value).FromMont()
			} else {
				return nil, backend.ErrInputNotSet
			}
		}
	}

	return toReturn, nil
}
