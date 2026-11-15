//go:build darwin

package app

import "testing"

// **Der Platz eines Ortes ist gemessen** — frei nicht mehr als gesamt, und
// ein Ort, den es nicht gibt, hat keinen.
func TestDerPlatzEinesOrtesIstGemessen(t *testing.T) {
	t.Parallel()

	frei, gesamt, err := platzAuf(t.TempDir())
	if err != nil || frei == 0 || gesamt < frei {
		t.Errorf("frei %d, gesamt %d, %v", frei, gesamt, err)
	}

	if _, _, err := platzAuf("/gibt/es/nicht/mse4mac"); err == nil {
		t.Error("Platz an einem Ort, den es nicht gibt")
	}
}
