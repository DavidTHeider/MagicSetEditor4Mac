package storage

// **Das Buch der Rückkehren** — Werkstück O7 von Phase 5.2
// (#1431).
//
// Es hält, was `domain.Wiederherstellungsepoche` beschreibt: wann
// zurückgegangen wurde, auf welchen Stand, aus welcher Datei — und welche
// Mengen das unklar gelassen hat.
//
// # Zwei Reichweiten in einer Antwort
//
// **Die Epoche sieht jedes Profil, ihre Mengen nur ihr eigenes.** Der Bestand
// ist gemeinsam zurückgegangen; das zu verschweigen wäre die Unsicherheit, die
// niemand sieht. Was *wem* fehlt, ist dagegen eine Frage an eine Sammlung, und
// die gehört einem Profil.

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Wiederherstellungsbuch ist die Ablage der Epochen.
type Wiederherstellungsbuch struct{ db *sql.DB }

// NeuesWiederherstellungsbuch baut sie über einer Datenbank.
func NeuesWiederherstellungsbuch(db *sql.DB) *Wiederherstellungsbuch {
	return &Wiederherstellungsbuch{db: db}
}

// Vermerken schreibt eine Rückkehr samt ihren Mengen — **in einer Transaktion**.
//
// **Eine Epoche ohne ihre Mengen wäre die schlechtere Hälfte**: sie sagte, dass
// etwas unklar ist, und nicht was. Ein Abbruch dazwischen darf sie nicht
// hinterlassen.
//
// **Jede Menge nennt ihr Profil selbst.** Eine Rückkehr betrifft den **ganzen**
// Bestand; sie unter dem Profil dessen abzulegen, der sie ausgelöst hat, hieße,
// die Unsicherheit der anderen zu verschweigen.
//
// # Und deshalb verlangt sie kein Profil im Kontext
//
// **Anders als jeder andere Schreibweg dieses Pakets.** Der Grund ist kein
// Versehen: ein Zurückspielen geschieht auch **beim Start**, bevor jemand ein
// Profil gewählt hat — die Wiederherstellungsseite ist genau dafür da. Ein
// Profilzwang hier hieße, dass ausgerechnet die Rückkehr aus einem kaputten
// Bestand keinen Nachweis bekommt.
//
// **Gelesen wird trotzdem nur mit Profil**: was *wem* fehlt, ist eine Frage an
// eine Sammlung, und die gehört einem.
func (w *Wiederherstellungsbuch) Vermerken(
	ctx context.Context, epoche domain.Wiederherstellungsepoche,
) (int64, error) {
	geschaeft, fehler := w.db.BeginTx(ctx, nil)
	if fehler != nil {
		return 0, uebersetze(fehler, "die Rückkehr vermerken")
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach einem Commit ohne Wirkung

	ergebnis, schreibfehler := geschaeft.ExecContext(ctx,
		`INSERT INTO wiederherstellungen
		   (zurueckgespielt, datenstand, quelle, klaerung, geklaert_am)
		 VALUES (?, ?, ?, ?, '')`,
		Zeitstempel(epoche.Zurueckgespielt), Zeitstempel(epoche.Datenstand),
		epoche.Quelle, string(domain.KlaerungOffen))
	if schreibfehler != nil {
		return 0, uebersetze(schreibfehler, "die Rückkehr vermerken")
	}

	kennung, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		return 0, uebersetze(kennungsfehler, "die Kennung der Rückkehr lesen")
	}

	for _, menge := range epoche.Mengen {
		if _, mengenfehler := geschaeft.ExecContext(ctx,
			`INSERT INTO wiederherstellungsmengen
			   (wiederherstellung_id, profil_id, schluessel, zustand, sprache,
			    finish, menge, name, setcode, nummer)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			kennung, menge.Profil, menge.Schluessel, menge.Zustand, menge.Sprache,
			string(menge.Ausfuehrung()), menge.Menge, menge.Name, menge.Setcode, menge.Nummer,
		); mengenfehler != nil {
			return 0, uebersetze(mengenfehler, "eine unklare Menge vermerken")
		}
	}

	if abschlussfehler := geschaeft.Commit(); abschlussfehler != nil {
		return 0, uebersetze(abschlussfehler, "die Rückkehr abschließen")
	}

	return kennung, nil
}

// Die zwei Abfragen — **ganz und nicht zusammengesetzt**.
//
// Ein `... + wo + ...` wäre hier harmlos, weil `wo` aus dieser Datei kommt und
// nie aus einer Eingabe. Es steht trotzdem nicht da: `gosec` rügt jede
// SQL-Verkettung, und eine Ausnahme mit der Begründung „diesmal nicht" ist
// genau die Gestalt, an der später jemand vorbeiliest (§10).
const (
	rueckkehrenAlle = `SELECT id, zurueckgespielt, datenstand, quelle, klaerung, geklaert_am
	     FROM wiederherstellungen
	    ORDER BY zurueckgespielt DESC, id DESC`

	rueckkehrenOffene = `SELECT id, zurueckgespielt, datenstand, quelle, klaerung, geklaert_am
	     FROM wiederherstellungen
	    WHERE klaerung = 'offen'
	    ORDER BY zurueckgespielt DESC, id DESC`
)

// Offene gibt die Rückkehren, die noch niemand quittiert hat.
func (w *Wiederherstellungsbuch) Offene(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return w.lesen(ctx, rueckkehrenOffene)
}

// Alle gibt jede Rückkehr — auch die quittierten.
//
// **Weil ein Nachweis, der verschwindet, keiner ist.** Das Issue verlangt
// dauerhafte Nachweise; eine quittierte Epoche wandert aus [Offene] heraus und
// bleibt hier stehen.
func (w *Wiederherstellungsbuch) Alle(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return w.lesen(ctx, rueckkehrenAlle)
}

// Quittieren erklärt eine Rückkehr für nachgesehen.
func (w *Wiederherstellungsbuch) Quittieren(
	ctx context.Context, kennung int64, wann time.Time,
) error {
	ergebnis, fehler := w.db.ExecContext(ctx,
		`UPDATE wiederherstellungen SET klaerung = ?, geklaert_am = ?
		  WHERE id = ? AND klaerung = 'offen'`,
		string(domain.KlaerungQuittiert), Zeitstempel(wann), kennung)
	if fehler != nil {
		return uebersetze(fehler, "die Rückkehr quittieren")
	}

	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return uebersetze(zaehlfehler, "die Quittung zählen")
	}

	// **Null heißt: es gibt sie nicht, oder sie ist längst quittiert.** Beides
	// ist dasselbe für den, der fragt — und beides ist etwas anderes als
	// „erledigt", das ein stiller Erfolg behaupten würde (§8).
	if betroffen == 0 {
		return fmt.Errorf("%w: diese Rückkehr steht nicht offen", domain.ErrNotFound)
	}

	return nil
}

// lesen holt die Epochen samt den Mengen des Profils.
func (w *Wiederherstellungsbuch) lesen(ctx context.Context, abfrage string) (
	[]domain.Wiederherstellungsepoche, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := w.db.QueryContext(ctx, abfrage)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Rückkehren lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	epochen := []domain.Wiederherstellungsepoche{}

	for zeilen.Next() {
		epoche, lesefehler := epocheAus(zeilen)
		if lesefehler != nil {
			return nil, lesefehler
		}

		epochen = append(epochen, epoche)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Rückkehren lesen")
	}

	for stelle := range epochen {
		mengen, mengenfehler := w.mengen(ctx, epochen[stelle].ID, profilID)
		if mengenfehler != nil {
			return nil, mengenfehler
		}

		epochen[stelle].Mengen = mengen
	}

	return epochen, nil
}

// epocheAus liest eine Zeile.
func epocheAus(zeilen *sql.Rows) (domain.Wiederherstellungsepoche, error) {
	var (
		epoche                               domain.Wiederherstellungsepoche
		zurueck, stand, klaerung, geklaertAm string
	)

	if fehler := zeilen.Scan(&epoche.ID, &zurueck, &stand, &epoche.Quelle,
		&klaerung, &geklaertAm); fehler != nil {
		return epoche, uebersetze(fehler, "eine Rückkehr lesen")
	}

	epoche.Klaerung = domain.Klaerungsstand(klaerung)

	for ziel, roh := range map[*time.Time]string{
		&epoche.Zurueckgespielt: zurueck,
		&epoche.Datenstand:      stand,
		&epoche.GeklaertAm:      geklaertAm,
	} {
		// **Leer bleibt leer.** `geklaert_am` einer offenen Epoche ist leer,
		// und eine Nullzeit daraus zu parsen wäre ein Fehler ohne Anlass.
		if roh == "" {
			continue
		}

		wann, zeitfehler := time.Parse(time.RFC3339Nano, roh)
		if zeitfehler != nil {
			return epoche, uebersetze(zeitfehler, "den Zeitpunkt %q lesen", roh)
		}

		*ziel = wann
	}

	return epoche, nil
}

// mengen holt die unklaren Mengen **dieses** Profils.
func (w *Wiederherstellungsbuch) mengen(ctx context.Context, kennung, profilID int64) (
	[]domain.Sammlungsmenge, error,
) {
	zeilen, fehler := w.db.QueryContext(ctx,
		`SELECT schluessel, zustand, sprache, finish, menge, name, setcode, nummer
		   FROM wiederherstellungsmengen
		  WHERE wiederherstellung_id = ? AND profil_id = ?
		  ORDER BY schluessel, zustand, sprache, finish`, kennung, profilID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die unklaren Mengen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	mengen := []domain.Sammlungsmenge{}

	for zeilen.Next() {
		var menge domain.Sammlungsmenge
		if lesefehler := zeilen.Scan(&menge.Schluessel, &menge.Zustand, &menge.Sprache,
			&menge.Finish, &menge.Menge, &menge.Name, &menge.Setcode,
			&menge.Nummer); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine unklare Menge lesen")
		}

		mengen = append(mengen, menge)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die unklaren Mengen lesen")
	}

	return mengen, nil
}
