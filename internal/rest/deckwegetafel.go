package rest

import "net/http"

// deckwege trägt die Wege der Decks und Formate ein — herausgelöst aus
// `wegetafel.go`, als die an der Zeilengrenze stand (§2).
func (d *Dienst) deckwege(weg *http.ServeMux) {
	// **Decks und Formate** (#1293, Vermerk 0061). `PUT` auf die Plätze, weil
	// derselbe Bezug im selben Bereich eine höhere Menge ist — idempotent.
	weg.HandleFunc("GET /decks", d.decks)
	weg.HandleFunc("POST /decks", d.deckAnlegen)
	weg.HandleFunc("GET /decks/{deckID}", d.deck)
	weg.HandleFunc("PUT /decks/{deckID}", d.deckUmbenennen)
	weg.HandleFunc("DELETE /decks/{deckID}", d.deckLoeschen)
	weg.HandleFunc("POST /decks/{deckID}/kopie", d.deckKopieren)
	// **Aus einem Set ein Deck** (FMT-04, #1693) — `POST`, es entsteht eines.
	weg.HandleFunc("POST /decks/aus-set", d.deckAusSet)
	weg.HandleFunc("POST /decks/{deckID}/fehlkarten/wunschliste", d.deckFehlkartenWuenschen)
	weg.HandleFunc("PUT /decks/{deckID}/plaetze", d.deckplatzLegen)
	weg.HandleFunc("DELETE /decks/{deckID}/plaetze/{platzID}", d.deckplatzNehmen)
	// **Der Bereichswechsel ist ein Weg** (Befund 7 des Zweitblicks über
	// Phase 5.1): legen und nehmen ließ eine Karte in beiden Bereichen.
	weg.HandleFunc("PUT /decks/{deckID}/plaetze/{platzID}/bereich", d.deckplatzVerschieben)
	weg.HandleFunc("PUT /decks/{deckID}/format", d.deckFormat)
	weg.HandleFunc("GET /decks/{deckID}/pruefung", d.deckPruefung)
	weg.HandleFunc("GET /decks/{deckID}/bestand", d.deckBestand)

	// **Die Decklisteneinfuhr** — Werkstück T8 von Phase 5.6 (#1473).
	//
	// **`POST` für beides**: die Liste steht im Rumpf und nicht in der
	// Adresse, und das Übernehmen ist ausdrücklich **nicht** idempotent —
	// es ist an den Stand gebunden, und ein zweiter Ruf mit demselben Stand
	// wird abgewiesen.
	weg.HandleFunc("POST /decks/{deckID}/liste/vorschau", d.decklisteVorschau)
	weg.HandleFunc("POST /decks/{deckID}/liste", d.decklisteUebernehmen)

	// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
	weg.HandleFunc("GET /decks/{deckID}/zuordnungen", d.deckZuordnungen)
	weg.HandleFunc("PUT /decks/{deckID}/zuordnungen", d.deckZuordnen)
	weg.HandleFunc("DELETE /decks/{deckID}/zuordnungen/{zuordnungID}", d.deckZuordnungLoesen)
	weg.HandleFunc("GET /decks/{deckID}/fehlkarten", d.deckFehlkarten)
	weg.HandleFunc("GET /decks/{deckID}/liste", d.deckliste)
	weg.HandleFunc("GET /sammlung/exemplare/{exemplarID}/decks", d.exemplarDecks)
	weg.HandleFunc("GET /formate", d.formate)
	weg.HandleFunc("POST /formate", d.formatAnlegen)
	weg.HandleFunc("GET /formatbauarten", d.formatbauarten)
	weg.HandleFunc("PUT /formate/{formatID}", d.formatSetzen)
	weg.HandleFunc("DELETE /formate/{formatID}", d.formatLoeschen)
}
