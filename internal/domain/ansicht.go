package domain

// **Wie eine Liste aussieht, wird gemerkt** — UX-11.AK-04
// (#1683):
// *„Ansichtspräferenzen sind je Fachbereich und Profil speicherbar."*
//
// **Dieselbe Bauform wie der Arbeitsplatz** (`arbeitsplatz.go`): der Mensch
// hat eine Wahl, das Profil erbt sie oder sticht sie, und beides reist mit
// dem Arbeitsplatz — nur auf Wunsch. Eine Ansicht ist eine Frage an ein
// Fenster und nicht an ein Set.
//
// **`new`, und wo gesucht wurde**: MSE2 hat eine Kartenliste und keine
// umschaltbaren Ansichten über Sets oder Decks; gesucht am 2026-09-25 in
// `/Users/david/mse2-quelle/src/gui` (Fassung f9d9356 vom 2021-07-06) nach
// `view`, `grid` und `tile`.

const (
	// AnsichtSets ist, wie die Sets stehen — als Kacheln oder als Liste.
	AnsichtSets = "arbeitsplatz.ansicht.sets"

	// AnsichtDecks ist, wie die Decks stehen.
	AnsichtDecks = "arbeitsplatz.ansicht.decks"

	// AnsichtSammlung ist die Sicht der Sammlung — Tabelle, Galerie oder
	// Gruppen (U8, #1419), seit UX-11 gemerkt.
	AnsichtSammlung = "arbeitsplatz.ansicht.sammlung"
)

// Die Wörter der Ansichten.
const (
	AnsichtKacheln = "kacheln"
	AnsichtListe   = "liste"
	AnsichtTabelle = "tabelle"
	AnsichtGalerie = "galerie"
	AnsichtGruppen = "gruppen"
)

// ansichtszeilen sind die drei Bereiche mit ihren Ansichten — **die erste ist
// die Vorgabe**, und sie ist, was dort bisher stand.
var ansichtszeilen = []struct {
	schluessel string
	werte      []string
}{
	{AnsichtSets, []string{AnsichtKacheln, AnsichtListe}},
	{AnsichtDecks, []string{AnsichtListe, AnsichtKacheln}},
	{AnsichtSammlung, []string{AnsichtTabelle, AnsichtGalerie, AnsichtGruppen}},
}

// ansichtsschluessel sind die drei Schlüssel.
func ansichtsschluessel() []string {
	schluessel := make([]string, 0, len(ansichtszeilen))
	for _, zeile := range ansichtszeilen {
		schluessel = append(schluessel, zeile.schluessel)
	}

	return schluessel
}

// haengeAnsichtenAn hängt die drei an beide Reihen — hinter den Arbeitsplatz,
// aus demselben `init` und aus demselben Grund (`arbeitsplatz.go`).
func haengeAnsichtenAn() {
	for _, zeile := range ansichtszeilen {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: zeile.schluessel,
			Wirkung:    WirkungSofort,
			Bereich:    BereichArbeitsplatz,
			Werte:      append([]string{Erben}, zeile.werte...),
			Vorgabe:    Erben,
		})

		nutzereinstellungen = append(nutzereinstellungen, nutzereinstellung{
			Schluessel: zeile.schluessel,
			Werte:      zeile.werte,
			Vorgabe:    zeile.werte[0],
		})
	}
}
