package storage

// **Was eine Sicht zum Ordnen braucht — und was sie danach zeigt** — Werkstück
// #1500.
//
// # Zwei Abfragen statt einer, und warum das schneller ist
//
// Eine Sicht auf die Sammlung filtert und ordnet über **alle** Zeilen und
// zeigt davon zweihundert. Sie dafür alle vollständig zu lesen heißt: einund-
// zwanzig Spalten und einen Verbund auf die Herkunft, zehntausendmal — für
// Felder, die neunundneunzig Prozent der Zeilen nie zeigen werden.
//
// Gemessen am 2026-09-10 an zehntausend Exemplaren (Apple M5, dieselbe Datei):
// die volle Abfrage samt Verbund und `ORDER BY` **12,0 ms**, die schmale ohne
// beides **5,9 ms**. Das ist die Hälfte in SQLite — und der größere Teil liegt
// darüber, im Bauen der Objekte: mit `-gcflags "all=-N -l"`, wie `wails dev`
// übersetzt, kostete der ganze Ruf **100–110 ms** und mit den zwei Abfragen
// **60–70 ms**.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ordnungsspalten sind die Felder, aus denen eine Sicht filtert und ordnet.
//
// **Neun statt einundzwanzig**: `domain.Sammlungssicht` liest Bezeichnung,
// Druck, Setcode und Notiz (die Suche), Foil (die Ausführung) sowie Menge und
// Erwerbsdatum (die Ordnung). Alles andere steht erst in der Zeile, die
// jemand **sieht**.
const ordnungsspalten = `s.id, s.name, s.setcode, s.nummer, s.printing,
	s.notiz, s.menge, s.erwerbsdatum, s.finish`

// Ordnungsliste gibt je Exemplar die Felder, aus denen eine Sicht filtert und
// ordnet — und sonst keines.
//
// **Die Exemplare sind unvollständig, und das ist der Sinn.** Sie taugen zum
// Ordnen und nicht zum Zeigen: Zustand, Sprache, Preis, Lagerort und Herkunft
// bleiben leer. Wer sie zeigt, zeigt eine Karte ohne ihre Angaben — deshalb
// heißt der Weg dorthin [Sammlungsspeicher.NachKennungen] und nicht „Liste".
//
// **Ohne `ORDER BY`**: geordnet wird in `domain` nach Sprache, und eine
// Reihenfolge, die danach ohnehin verworfen wird, ist ein B-Baum umsonst.
func (s *Sammlungsspeicher) Ordnungsliste(ctx context.Context) ([]*domain.Exemplar, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return s.ordnungLesenBei(ctx, s.db, profilID)
}

// Fenster liest die Ordnung und die Zeilen des Ausschnitts in **einem** Stand.
//
// **`waehle` entscheidet dazwischen** — Werkstück O11 von Phase 5.2, Befund 13
// des Zweitblicks: welche Zeilen ein Fenster zeigt, weiß der Fachkern und
// nicht der Bestand; **dass beide Abfragen dieselbe Sammlung sehen**, weiß
// nur der Bestand. Deshalb geht die Wahl als Rückruf hinein, statt dass der
// Aufrufer zweimal fragt.
//
// **Die Reihenfolge kommt aus `waehle`**, wie bei [Sammlungsspeicher.NachKennungen].
func (s *Sammlungsspeicher) Fenster(
	ctx context.Context, waehle func(ordnung []*domain.Exemplar) []int64,
) ([]*domain.Exemplar, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	var fenster []*domain.Exemplar

	fehler := imLesegeschaeft(ctx, s.db, "das Fenster der Sammlung", func(geschaeft *sql.Tx) error {
		ordnung, lesefehler := s.ordnungLesenBei(ctx, geschaeft, profilID)
		if lesefehler != nil {
			return lesefehler
		}

		gewaehlt, holfehler := s.nachKennungenBei(ctx, geschaeft, profilID, waehle(ordnung))
		if holfehler != nil {
			return holfehler
		}

		fenster = gewaehlt

		return nil
	})
	if fehler != nil {
		return nil, fehler
	}

	return fenster, nil
}

// ordnungLesenBei liest die Ordnungszeilen eines Profils.
//
// **Als eigener Weg neben [Sammlungsspeicher.lesen]**, weil die Spalten andere
// sind — und damit ein weißer Lauf den `Scan` gegen die falsche Spaltenzahl
// betreten kann. Das ist der Zweig, den niemand herbeiführen kann, solange
// Abfrage und `Scan` zusammenpassen (siehe `sammlungsfehler_intern_test.go`).
func (s *Sammlungsspeicher) ordnungLesenBei(
	ctx context.Context, bei abfrager, profilID int64,
) ([]*domain.Exemplar, error) {
	return s.ordnungLesen(ctx, bei,
		`SELECT `+ordnungsspalten+` FROM sammlung s WHERE s.profil_id = ?`, profilID)
}

// ordnungLesen führt eine schmale Abfrage aus und baut die Zeilen.
func (s *Sammlungsspeicher) ordnungLesen(
	ctx context.Context, bei abfrager, abfrage string, werte ...any,
) ([]*domain.Exemplar, error) {
	zeilen, err := bei.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "die Sammlung ordnen")
	}

	defer func() { _ = zeilen.Close() }()

	exemplare := []*domain.Exemplar{}

	for zeilen.Next() {
		exemplar := &domain.Exemplar{}

		if fehler := zeilen.Scan(&exemplar.ID, &exemplar.Name, &exemplar.Setcode,
			&exemplar.Nummer, &exemplar.Printing, &exemplar.Notiz,
			&exemplar.Menge, &exemplar.Erwerbsdatum, &exemplar.Finish); fehler != nil {
			return nil, uebersetze(fehler, "eine Ordnungszeile lesen")
		}

		exemplare = append(exemplare, exemplar)
	}

	return exemplare, uebersetze(zeilen.Err(), "die Sammlung ordnen")
}

// stapelgroesse ist, wie viele Kennungen eine Abfrage höchstens trägt.
//
// **Weil eine Abfrage nicht beliebig viele Platzhalter hat.** SQLite begrenzt
// sie (`SQLITE_MAX_VARIABLE_NUMBER`), und ein Fenster, das jemand über die
// Kommandozeile auf zehntausend Zeilen stellt, liefe sonst in eine Schranke,
// die von der Übersetzung des Treibers abhängt statt von diesem Programm.
const stapelgroesse = 500

// NachKennungen gibt die vollen Exemplare zu den genannten Kennungen — in der
// **übergebenen** Reihenfolge.
//
// **Die Reihenfolge kommt vom Rufenden**, nicht aus der Datenbank: sie ist das
// Ergebnis der Ordnung, die `domain` gerechnet hat, und ein `ORDER BY` hier
// machte sie wieder zunichte.
//
// **Was es nicht gibt, fehlt still.** Zwischen dem Ordnen und dem Zeigen kann
// jemand gelöscht haben; eine Zeile weniger ist die richtige Antwort, ein
// Fehler wäre es nicht.
func (s *Sammlungsspeicher) NachKennungen(ctx context.Context, kennungen []int64) (
	[]*domain.Exemplar, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return s.nachKennungenBei(ctx, s.db, profilID, kennungen)
}

// nachKennungenBei ist dasselbe über einer genannten Abfragestelle.
func (s *Sammlungsspeicher) nachKennungenBei(
	ctx context.Context, bei abfrager, profilID int64, kennungen []int64,
) ([]*domain.Exemplar, error) {
	gefunden := make(map[int64]*domain.Exemplar, len(kennungen))

	for anfang := 0; anfang < len(kennungen); anfang += stapelgroesse {
		stapel := kennungen[anfang:min(anfang+stapelgroesse, len(kennungen))]

		werte := make([]any, 0, len(stapel)+1)
		werte = append(werte, profilID)
		for _, kennung := range stapel {
			werte = append(werte, kennung)
		}

		exemplare, fehler := s.lesenBei(ctx, bei,
			`SELECT `+sammlungsspalten+`
			   FROM sammlung s
			   LEFT JOIN sammlung_herkunft h ON h.exemplar_id = s.id
			  WHERE s.profil_id = ? AND s.id IN (`+
				strings.TrimPrefix(strings.Repeat(",?", len(stapel)), ",")+`)`, werte...)
		if fehler != nil {
			return nil, fehler
		}

		for _, eines := range exemplare {
			gefunden[eines.ID] = eines
		}
	}

	geordnet := make([]*domain.Exemplar, 0, len(kennungen))

	for _, kennung := range kennungen {
		if eines, da := gefunden[kennung]; da {
			geordnet = append(geordnet, eines)
		}
	}

	return geordnet, nil
}
