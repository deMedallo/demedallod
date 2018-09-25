// Copyright 2014 The go-ethereum Authors
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

package pow

type PoW interface {
	Search(block Block, stop <-chan struct{}, index int) uint64
	Verify(block Block) bool
	GetHashrate() int64
	Turbo(bool)
}
