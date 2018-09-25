// Copyright 2017 The go-ethereum Authors
// Copyright 2018 deMedallo project
// This file is part of deMedallo.
//
// deMedallo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// deMedallo is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with deMedallo. If not, see <http://www.gnu.org/licenses/>.

// This file contains configuration literals.

package core

import (
	"math/big"

	"github.com/demedallo/demedallod/core/vm"
)

var DefaultDiehardGasTable = &vm.GasTable{
	ExtcodeSize:     big.NewInt(700),
	ExtcodeCopy:     big.NewInt(700),
	Balance:         big.NewInt(400),
	SLoad:           big.NewInt(200),
	Calls:           big.NewInt(700),
	Suicide:         big.NewInt(5000),
	ExpByte:         big.NewInt(50),
	CreateBySuicide: big.NewInt(25000),
}
