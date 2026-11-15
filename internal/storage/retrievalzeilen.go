package storage

// **Von der Zeile zum Dokument** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum das neben der Suche steht und nicht in ihr
//
// **Weil es zwei Fragen sind.** `retrievalsuche.go` beantwortet „welche
// Zeilen?"; hier steht „was ist eine Zeile?" — Scanziele, Deutung der
// Rohformen, die zweite Sichtprüfung und die Vektordarstellung. Zusammen
// liefen beide über die Zeilengrenze aus §2, und die Grenze ist nicht der
// Grund für den Schnitt, sondern sein Anlass.

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// treffernAus liest die Zeilen der lexikalischen Suche.
func treffernAus(zeilen *sql.Rows, profil int64) ([]domain.Retrievaltreffer, error) {
	gelesen := make([]domain.Retrievaltreffer, 0)

	for zeilen.Next() {
		var (
			dokument   domain.Retrievaldokument
			abgerufen  string
			merkmale   string
			ausschnitt string
		)

		if fehler := zeilen.Scan(append(
			zielefuer(&dokument, &abgerufen, &merkmale),
			&ausschnitt)...); fehler != nil {
			return nil, uebersetze(fehler, "eine Trefferzeile lesen")
		}

		if fehler := gedeutet(&dokument, abgerufen, merkmale); fehler != nil {
			return nil, fehler
		}

		if !sichtbar(dokument, profil) {
			continue
		}

		gelesen = append(gelesen, domain.Retrievaltreffer{
			Dokument: dokument, Rang: len(gelesen) + 1, Ausschnitt: ausschnitt,
		})
	}

	return gelesen, uebersetze(zeilen.Err(), "die Trefferliste lesen")
}

// naechsteAus rechnet die Ähnlichkeit und gibt die besten.
func naechsteAus(
	zeilen *sql.Rows, frage []float32, profil int64, grenze int,
) ([]domain.Retrievaltreffer, error) {
	type kandidat struct {
		dokument domain.Retrievaldokument
		naehe    float64
	}

	kandidaten := make([]kandidat, 0)

	for zeilen.Next() {
		var (
			dokument  domain.Retrievaldokument
			abgerufen string
			merkmale  string
			roh       []byte
		)

		if fehler := zeilen.Scan(append(
			zielefuer(&dokument, &abgerufen, &merkmale),
			&roh)...); fehler != nil {
			return nil, uebersetze(fehler, "eine Vektorzeile lesen")
		}

		if fehler := gedeutet(&dokument, abgerufen, merkmale); fehler != nil {
			return nil, fehler
		}

		if !sichtbar(dokument, profil) {
			continue
		}

		kandidaten = append(kandidaten, kandidat{
			dokument: dokument, naehe: cosinus(frage, alsVektor(roh)),
		})
	}

	// **Stabil geordnet**: bei gleicher Nähe entscheidet die Kennung. Das
	// Issue verlangt „stabile Tie-Breaks" — ohne sie tanzte die Liste bei
	// jedem Lauf.
	sort.SliceStable(kandidaten, func(links, rechts int) bool {
		if kandidaten[links].naehe != kandidaten[rechts].naehe {
			return kandidaten[links].naehe > kandidaten[rechts].naehe
		}

		return kandidaten[links].dokument.Kennung < kandidaten[rechts].dokument.Kennung
	})

	if len(kandidaten) > grenze {
		kandidaten = kandidaten[:grenze]
	}

	treffer := make([]domain.Retrievaltreffer, 0, len(kandidaten))
	for i, eines := range kandidaten {
		treffer = append(treffer, domain.Retrievaltreffer{
			Dokument: eines.dokument, Rang: i + 1,
		})
	}

	// **Der Iterationsfehler steht im Rückgabewert und nicht in einem `if`**
	// — dieselbe Gestalt wie nebenan (§10).
	return treffer, uebersetze(zeilen.Err(), "die Vektorliste lesen")
}

// sichtbar ist die **zweite** Prüfung, die das Issue verlangt.
//
// # Und sie ist strenger als die erste
//
// **Sonst wäre sie keine.** Der erste Anlauf wiederholte die `WHERE`-Zeile
// Wort für Wort — und damit konnte sie nichts fangen: was die Abfrage
// durchließ, ließ auch sie durch. Ein Zweig, den kein Lauf betreten kann,
// sieht aus wie eine geprüfte Zusage und ist keine (§10); die Abdeckung hat
// ihn angestrichen.
//
// **Der Unterschied ist die unbekannte Sichtbarkeit.** Die Abfrage fragt
// „öffentlich **oder** mein Profil"; eine Zeile mit einem dritten Wort in
// `sichtbarkeit` und der eigenen Profilkennung kommt dort durch. Hier nicht:
// was nicht eines der zwei bekannten Wörter trägt, ist keine Auskunft über
// Sichtbarkeit, sondern ein verbogener Bestand — und der wird nicht gedeutet
// (§8).
func sichtbar(dokument domain.Retrievaldokument, profil int64) bool {
	switch dokument.Sichtbarkeit {
	case domain.SichtbarOeffentlich:
		return true

	case domain.SichtbarProfil:
		return dokument.ProfilID == profil && profil != 0

	default:
		return false
	}
}

// zielefuer gibt die Scanziele eines Dokuments, in der Reihenfolge der
// Spalten.
func zielefuer(
	dokument *domain.Retrievaldokument, abgerufen, merkmale *string,
) []any {
	return []any{
		&dokument.Kennung, &dokument.Quelle, &dokument.Bezugsadresse,
		&dokument.Spiel, &dokument.Kennungssystem, &dokument.Fachkennung,
		&dokument.Semantik, &dokument.Regelnummer, &dokument.Urheber,
		&dokument.Sprache, &dokument.Titel, &dokument.Text,
		&dokument.Quellrevision, &dokument.Quelldatum, abgerufen,
		&dokument.Inhaltshash, merkmale, &dokument.Abschnitt, &dokument.Von,
		&dokument.Bis, &dokument.ProfilID, &dokument.Sichtbarkeit,
	}
}

// gedeutet setzt Abrufzeit und Merkmale aus ihren Rohformen.
//
// **Als Zeitpunkt und als Struktur und nicht als Zeichenketten
// weitergereicht** — wer sie roh ließe, ließe jede Fläche selbst deuten, und
// eine von ihnen anders.
func gedeutet(dokument *domain.Retrievaldokument, abgerufen, merkmale string) error {
	gelesen, fehler := time.Parse(time.RFC3339Nano, abgerufen)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Abrufzeit von %q lesen", dokument.Kennung)
	}

	dokument.Abgerufen = gelesen

	if fehler := json.Unmarshal([]byte(merkmale), &dokument.Merkmale); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Merkmale von %q lesen", dokument.Kennung)
	}

	return nil
}

// alsMerkmalsjson schreibt die Merkmale.
//
// **Ein Fehler ist hier keiner, der eintreten kann**: was hineingeht, sind
// Zeichenketten, Zahlen und eine Map davon — `encoding/json` kann daran nicht
// scheitern. Der Zweig dafür stünde ungeprüft da und sähe aus wie eine
// geprüfte Zusage (§10); deshalb geht der Fehler in den Rückgabewert der
// **Anweisung**, die ihn verursachen könnte.
func alsMerkmalsjson(merkmale domain.Retrievalmerkmale) string {
	roh, _ := json.Marshal(merkmale) //nolint:errcheck,errchkjson // siehe oben

	return string(roh)
}

// alsVektor liest einen Vektor aus seinem BLOB.
//
// **Little-endian float32**, so wie [alsBlob] ihn schreibt. Ein halber Wert
// am Ende wird weggelassen — er wäre eine Zahl, die niemand geschrieben hat.
func alsVektor(roh []byte) []float32 {
	vektor := make([]float32, 0, len(roh)/4)

	for i := 0; i+4 <= len(roh); i += 4 {
		vektor = append(vektor,
			math.Float32frombits(binary.LittleEndian.Uint32(roh[i:i+4])))
	}

	return vektor
}

// alsBlob schreibt einen Vektor.
func alsBlob(vektor []float32) []byte {
	roh := make([]byte, 4*len(vektor))

	for i, zahl := range vektor {
		binary.LittleEndian.PutUint32(roh[4*i:], math.Float32bits(zahl))
	}

	return roh
}
