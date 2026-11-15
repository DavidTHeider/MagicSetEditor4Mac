// **Das Profil eines Aufrufs** — Werkstück O1 von Phase 5.2
// (#1425).
//
// # Warum die Ablage das beantwortet und nicht der Dienst
//
// **Weil der Kontextschlüssel hier liegt** ([MitProfil], [ProfilAus]) und
// `internal/service` dieses Paket nicht kennen darf (§2). Der Dienst fragt
// „in welchem Profil arbeitet dieser Aufruf?" — und die Antwort hat zwei
// Quellen: was der Kontext trägt (ein Token sagt es, ein Wechsel im Fenster
// sagt es), sonst der Zeiger im Bestand.

package storage

import "context"

// Arbeitsprofil sagt, in welchem Profil dieser Aufruf arbeitet.
func (a *Nutzerablage) Arbeitsprofil(ctx context.Context) (int64, error) {
	if profilID, fehler := ProfilAus(ctx); fehler == nil {
		return profilID, nil
	}

	return a.Profilzeiger(ctx)
}
