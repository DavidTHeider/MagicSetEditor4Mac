// Die **Sammlung** im Bestand
// (#1165).
//
// # Warum eigene Tabellen und nicht Kartenfelder
//
// Weil ein Exemplar kein Entwurf ist. Die Begründung steht in
// [0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md) und
// in der Wanderung `0018_sammlung.sql`, seit 0189 Teil der Grundlinie. Kurz:
// Sammlungsdaten in `card_values` landeten in jeder `.mse-set`, die man teilt.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungsspeicher liest und schreibt die Exemplare.
type Sammlungsspeicher struct {
	db *sql.DB
}

// NeuerSammlungsspeicher baut ihn über einer Datenbank.
func NeuerSammlungsspeicher(db *sql.DB) *Sammlungsspeicher {
	return &Sammlungsspeicher{db: db}
}

// sammlungsspalten sind die Spalten in der Reihenfolge, in der gelesen wird.
//
// **Einmal aufgeschrieben und zweimal benutzt.** Zwei Aufzählungen, die
// auseinanderlaufen, ergeben ein `Scan` gegen die falsche Spalte — und das
// fällt erst auf, wenn zwei Textfelder vertauscht sind.
const sammlungsspalten = `s.id, s.fassung, s.printing, s.druck_id, s.menge, s.zustand,
	s.sprache, s.finish, s.erwerbsdatum, s.erwerbspreis, s.erwerbswaehrung,
	s.notiz, s.name, s.setcode, s.nummer, s.lagerort, s.wunsch,
	s.wunschrang, s.wunschart,
	coalesce(h.art, ''), coalesce(h.quelle, ''), coalesce(h.quell_profil, 0),
	coalesce(h.uebernommen, '')`

// **Die Herkunft kommt über einen `LEFT JOIN` mit** — Werkstück 12 von
// Phase 4.12 (#1351). Sie fehlt bei fast jeder Zeile; sie einzeln
// nachzufragen wäre ein Ruf je Exemplar, und ein Aufrufer, der es vergisst,
// zeigte eine Sammlung ohne Herkunft, ohne dass jemand es merkt.
//
// **`coalesce` und kein `sql.NullString`**: eine fehlende Herkunft ist die
// leere, und `domain.Herkunft{}` ist genau das — „hier entstanden".
//
// **Und der Tabellenname steht ausgeschrieben in jeder Abfrage**, nicht in
// einer Konstante daneben. Beim ersten Anlauf stand er in einer, und
// `TestJedeWurzelabfrageKlemmtAufEinProfil` sah diese beiden Abfragen nicht
// mehr — 48 statt 50. Der Wächter hat es selbst gemeldet, weil er eine
// Untergrenze zählt; ohne die wäre er grün geblieben und hätte zwei Abfragen
// weniger geprüft (§10: die Reichweite einer Messung).

// Anlegen legt ein Exemplar an und gibt seine Kennung.
//
// **Ein zweites gleiches ist ein [domain.ErrConflict]** — die Tabelle hat eine
// Eindeutigkeit über Druck, Zustand, Sprache und Ausführung. Warum das kein
// stilles Addieren ist, steht in der Migration.
func (s *Sammlungsspeicher) Anlegen(
	ctx context.Context, exemplar *domain.Exemplar,
) (int64, error) {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return 0, profilfehler
	}

	// **In einem Zug mit der Zeile des Drucks** (A2, #1441): der Alias legt
	// den Druck an oder findet ihn, und beides steht oder fällt zusammen.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, uebersetze(err, "ein Exemplar anlegen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ist sie wirkungslos

	id, fehler := exemplarEinfuegen(ctx, tx, exemplar, profilID)
	if fehler != nil {
		return 0, fehler
	}

	// **Ein Satz und kein Zweig** — dieselbe Form wie in `decks.go`.
	return id, uebersetze(tx.Commit(), "ein Exemplar anlegen")
}

// exemplarEinfuegen schreibt eine Zeile samt der Zeile ihres Drucks in einen
// Zug — für `Anlegen` und für `AlleAnlegen`, damit die Anweisung **eine** ist.
func exemplarEinfuegen(
	ctx context.Context, tx *sql.Tx, exemplar *domain.Exemplar, profilID int64,
) (int64, error) {
	druck, druckfehler := druckZeile(ctx, tx, exemplarbezug(exemplar))
	if druckfehler != nil {
		return 0, druckfehler
	}

	var id int64

	// **`RETURNING` statt `LastInsertId`** — ein Weg weniger, den kein echter
	// Treiber hergibt (§3).
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO sammlung
		   (printing, druck_id, menge, zustand, sprache, finish,
		    erwerbsdatum, erwerbspreis, erwerbswaehrung, notiz,
		    name, setcode, nummer, lagerort, wunsch, wunschrang, wunschart,
		    profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 RETURNING id`,
		exemplar.Printing, druck, exemplar.Menge, string(exemplar.Zustand),
		exemplar.Sprache, string(exemplar.Ausfuehrung()), exemplar.Erwerbsdatum,
		exemplar.Erwerbspreis, exemplar.Erwerbswaehrung, exemplar.Notiz,
		exemplar.Name, exemplar.Setcode, exemplar.Nummer,
		exemplar.Lagerort, exemplar.Wunsch, string(exemplar.Wunschrang),
		string(exemplar.Wunschart), profilID).Scan(&id); err != nil {
		return 0, uebersetze(err, "ein Exemplar anlegen")
	}

	return id, nil
}

// exemplarbezug sagt der Ablage, welchen Druck ein Exemplar meint.
func exemplarbezug(exemplar *domain.Exemplar) druckbezug {
	return druckbezug{
		druckID: exemplar.DruckID, printing: exemplar.Printing,
		name: exemplar.Name, setcode: exemplar.Setcode, nummer: exemplar.Nummer,
	}
}

// Aendern schreibt ein Exemplar zurück — wenn seine Fassung noch stimmt.
//
// # Warum die Fassung in der Bedingung steht
//
// Ohne sie ersetzte jedes Schreiben alle Spalten allein nach `id`. Zwei
// Fenster konnten dieselbe Zeile bearbeiten, und wer später speicherte,
// überschrieb die frühere Änderung — unbemerkt (§7, Regel 2). Angestrichen vom
// Zweitblick über Phase 3.4.
//
// **Eine Zeile, die es nicht gibt, und eine mit alter Fassung sind hier
// dasselbe `UPDATE`** — beide treffen keine Zeile. Unterschieden werden sie
// danach über eine zweite Frage; ein `WHERE`, das beides gleich behandelte,
// meldete „gibt es nicht", wo jemand anderes geschrieben hat.
func (s *Sammlungsspeicher) Aendern(
	ctx context.Context, exemplar *domain.Exemplar,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "ein Exemplar ändern")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ist sie wirkungslos

	druck, druckfehler := druckZeile(ctx, tx, exemplarbezug(exemplar))
	if druckfehler != nil {
		return druckfehler
	}

	// **Die Zusagen im selben Zug** — Werkstück `T2` (#1467). Die **Regel**
	// steht in der Domäne, der **Augenblick** hier: wer erst läse und dann
	// schriebe, ließe zwischen beidem eine zweite Zusage zu, die gegen einen
	// Besitz rechnet, den dieser Zug gerade wegschreibt.
	if fehler := zusagenTragen(ctx, tx, profilID, exemplar); fehler != nil {
		return fehler
	}

	ergebnis, err := tx.ExecContext(ctx,
		`UPDATE sammlung
		    SET printing = ?, druck_id = ?, menge = ?, zustand = ?, sprache = ?, finish = ?,
		        erwerbsdatum = ?, erwerbspreis = ?, erwerbswaehrung = ?,
		        notiz = ?, name = ?, setcode = ?, nummer = ?,
		        lagerort = ?, wunsch = ?, wunschrang = ?, wunschart = ?,
		        fassung = fassung + 1
		  WHERE id = ? AND fassung = ? AND profil_id = ?`,
		exemplar.Printing, druck, exemplar.Menge, string(exemplar.Zustand),
		exemplar.Sprache, string(exemplar.Ausfuehrung()), exemplar.Erwerbsdatum,
		exemplar.Erwerbspreis, exemplar.Erwerbswaehrung, exemplar.Notiz,
		exemplar.Name, exemplar.Setcode, exemplar.Nummer,
		exemplar.Lagerort, exemplar.Wunsch, string(exemplar.Wunschrang),
		string(exemplar.Wunschart),
		exemplar.ID, exemplar.Fassung, profilID)
	if err != nil {
		return uebersetze(err, "ein Exemplar ändern")
	}

	// **Erst abschließen, dann deuten**: `geaendert` fragt über `s.db` nach,
	// und die eine Verbindung (`SetMaxOpenConns(1)`) gehört bis hierher dem Zug.
	if fehler := tx.Commit(); fehler != nil {
		return uebersetze(fehler, "ein Exemplar ändern")
	}

	return s.geaendert(ctx, ergebnis, exemplar.ID, profilID)
}

// geaendert deutet das Ergebnis einer Änderung.
//
// **Ein eigenes Verfahren, damit sein Zählfehler zu betreten ist.** `modernc`
// zählt immer; über [Sammlungsspeicher.Aendern] wäre der Zweig einer, den kein
// Lauf erreicht — und ein Fehlerpfad, den kein Test betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3).
func (s *Sammlungsspeicher) geaendert(
	ctx context.Context, ergebnis sql.Result, id, profilID int64,
) error {
	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return s.warumNicht(ctx, id, profilID)
	}

	return nil
}

// warumNicht sagt, ob eine Zeile fehlt oder ihre Fassung veraltet ist.
//
// **Zwei Antworten auf denselben Fehlschlag**, und der Unterschied ist der
// halbe Zweck: „gibt es nicht" heißt aufräumen, „jemand anderes hat
// geschrieben" heißt nachladen (§8).
//
// **Das Profil kommt vom Rufer.** Es hier ein zweites Mal aus dem Kontext zu
// holen wäre ein Zweig, den kein Aufruf betreten kann: der Rufer hat es schon
// gelesen und wäre sonst gar nicht bis hierher gekommen (§3).
func (s *Sammlungsspeicher) warumNicht(
	ctx context.Context, id, profilID int64,
) error {
	var vorhanden int64

	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM sammlung WHERE id = ? AND profil_id = ?`,
		id, profilID).Scan(&vorhanden)
	if err != nil {
		return uebersetze(err, "das Exemplar nachschlagen")
	}

	if vorhanden == 0 {
		return fmt.Errorf("%w: Exemplar %d gibt es nicht", domain.ErrNotFound, id)
	}

	return fmt.Errorf(
		"%w — das Exemplar %d hat sich seit dem Lesen geändert", domain.ErrUeberholt, id)
}

// Loeschen nimmt ein Exemplar aus der Sammlung.
//
// # Warum die Fassung auch hier steht
//
// Sie stand seit #1196 nur an [Sammlungsspeicher.Aendern], und damit war die
// **Klasse** offen, die dort geschlossen werden sollte: ein schreibender Zug
// gegen einen veralteten Stand entfernt fremde Arbeit. Jemand liest „ein
// Stück", ein zweites Fenster trägt drei nach, das erste löscht seinen alten
// Stand — weg sind vier, und niemand erfährt es.
//
// Angestrichen vom Nachlauf über Phase 3.4 als Befund 10.
func (s *Sammlungsspeicher) Loeschen(ctx context.Context, id, fassung int64) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "ein Exemplar löschen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit wirkungslos

	// **Eine zugesagte Zeile geht nicht fort** (T2, #1467). Ohne diese Frage
	// nähme `ON DELETE CASCADE` die Zusage mit, und der Handel wüsste nichts
	// mehr davon.
	zusagen, zusagenfehler := zusagenInTx(ctx, tx, profilID, id)
	if zusagenfehler != nil {
		return zusagenfehler
	}

	if fehler := domain.LoeschenPasst(id, zusagen, time.Now()); fehler != nil {
		return fehler
	}

	ergebnis, err := tx.ExecContext(ctx,
		`DELETE FROM sammlung WHERE id = ? AND fassung = ? AND profil_id = ?`,
		id, fassung, profilID)
	if err != nil {
		return uebersetze(err, "ein Exemplar löschen")
	}

	if fehler := tx.Commit(); fehler != nil {
		return uebersetze(fehler, "ein Exemplar löschen")
	}

	return s.geaendert(ctx, ergebnis, id, profilID)
}

// zusagenTragen weist ein Schreiben ab, das zugesagte Stücke wegnähme.
//
// **Ein eigenes Verfahren, damit sein Lesefehler zu betreten ist** —
// derselbe Grund wie bei [Sammlungsspeicher.geaendert].
func zusagenTragen(
	ctx context.Context, wer bestandsfrager, profilID int64,
	exemplar *domain.Exemplar,
) error {
	zusagen, fehler := zusagenInTx(ctx, wer, profilID, exemplar.ID)
	if fehler != nil {
		return fehler
	}

	return domain.AenderungPasst(*exemplar, zusagen, time.Now())
}

// Liste gibt die ganze Sammlung.
//
// **Nach Druck, Zustand, Sprache und Ausführung sortiert**, also genau nach
// dem, was eine Zeile ausmacht. Ohne `ORDER BY` gäbe SQLite zurück, was ihm
// gerade passt, und niemand fände etwas wieder.
func (s *Sammlungsspeicher) Liste(ctx context.Context) ([]*domain.Exemplar, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return s.lesen(ctx,
		`SELECT `+sammlungsspalten+`
		   FROM sammlung s
		   LEFT JOIN sammlung_herkunft h ON h.exemplar_id = s.id
		  WHERE s.profil_id = ?
		  ORDER BY s.printing, s.zustand, s.sprache, s.finish`, profilID)
}

// Fuer gibt die Exemplare eines Drucks.
func (s *Sammlungsspeicher) Fuer(ctx context.Context, printing string) (
	[]*domain.Exemplar, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return s.lesen(ctx,
		`SELECT `+sammlungsspalten+`
		   FROM sammlung s
		   LEFT JOIN sammlung_herkunft h ON h.exemplar_id = s.id
		  WHERE s.profil_id = ? AND s.printing = ?
		  ORDER BY s.zustand, s.sprache, s.finish`, profilID, printing)
}

// NachDruck gibt die Exemplare einer Druckzeile — auch die ohne Alias.
//
// **Neben [Sammlungsspeicher.Fuer] und nicht an seiner Stelle**: der Alias
// bleibt, was die Flächen kennen (0086); die Zeile ist, was ein vorläufiger
// Druck hat. Die Listeneinfuhr fragt hier nach, bevor sie einen Stapel erhöht.
func (s *Sammlungsspeicher) NachDruck(ctx context.Context, druckID int64) (
	[]*domain.Exemplar, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return s.lesen(ctx,
		`SELECT `+sammlungsspalten+`
		   FROM sammlung s
		   LEFT JOIN sammlung_herkunft h ON h.exemplar_id = s.id
		  WHERE s.profil_id = ? AND s.druck_id = ?
		  ORDER BY s.zustand, s.sprache, s.finish`, profilID, druckID)
}

// Einzeln gibt ein Exemplar nach seiner Kennung — im eigenen Profil.
//
// **`ErrNotFound` und kein `nil`** (U9, #1420): wer ein Exemplar zuordnen
// will, das es nicht gibt, soll das hören — anders als beim Verweis, wo
// „keiner" der Normalfall ist.
func (s *Sammlungsspeicher) Einzeln(ctx context.Context, id int64) (*domain.Exemplar, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	exemplare, err := s.lesen(ctx,
		`SELECT `+sammlungsspalten+`
		   FROM sammlung s
		   LEFT JOIN sammlung_herkunft h ON h.exemplar_id = s.id
		  WHERE s.profil_id = ? AND s.id = ?`, profilID, id)
	if err != nil {
		return nil, err
	}

	if len(exemplare) == 0 {
		return nil, fmt.Errorf("%w: Exemplar %d gibt es nicht", domain.ErrNotFound, id)
	}

	return exemplare[0], nil
}
