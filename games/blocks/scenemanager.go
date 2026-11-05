// Copyright 2014 Hajime Hoshi
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

// This file is modified from Ebitengine v2.10.5 examples/blocks/blocks for ikaten.
// The changes are listed in docs/porting.md.

package blocks

import (
	ebiten "github.com/sat0ken/ikaten"
)

type Scene interface {
	Update(state *GameState) error
	Draw(screen *ebiten.Image)
}

const transitionMaxCount = 20

type SceneManager struct {
	current         Scene
	next            Scene
	transitionCount int
}

type GameState struct {
	SceneManager *SceneManager
	Input        *Input
}

func (s *SceneManager) Update(input *Input) error {
	if s.transitionCount == 0 {
		return s.current.Update(&GameState{
			SceneManager: s,
			Input:        input,
		})
	}

	s.transitionCount--
	if s.transitionCount > 0 {
		return nil
	}

	s.current = s.next
	s.next = nil
	return nil
}

func (s *SceneManager) Draw(r *ebiten.Image) {
	if s.transitionCount == 0 {
		s.current.Draw(r)
		return
	}

	// Fade through black instead of cross-fading. A cross-fade needs two
	// offscreen images of the screen size, which do not fit in the RAM.
	const half = transitionMaxCount / 2
	if s.transitionCount > half {
		s.current.Draw(r)
		fillTranslucent(r, 0, 0, ScreenWidth, ScreenHeight, float32(transitionMaxCount-s.transitionCount)/half)
		return
	}
	s.next.Draw(r)
	fillTranslucent(r, 0, 0, ScreenWidth, ScreenHeight, float32(s.transitionCount)/half)
}

func (s *SceneManager) GoTo(scene Scene) {
	if s.current == nil {
		s.current = scene
	} else {
		s.next = scene
		s.transitionCount = transitionMaxCount
	}
}
