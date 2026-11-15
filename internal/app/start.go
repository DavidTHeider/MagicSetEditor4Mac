// **Der Start der Anwendung** — öffnen, migrieren, auffrischen.
//
// Aus `app.go` herausgelöst, als es über 400 Zeilen ging (§2). Der Schnitt
// folgt der Frage: dort steht, **woraus** die Anwendung besteht und wie sie
// zusammengesetzt wird, hier, **was beim Aufwachen geschieht** — und in
// welcher Reihenfolge.

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Starten öffnet die Datenbank und migriert sie.
//
// **Dass hier migriert wird, ist der Punkt.** Der Migrator sagt von sich „jeder
// Start migriert" — bis zum Zweitblick am Ende von Phase 1 stimmte das nicht:
// die einzige vollständige Folge aus Öffnen und Migrieren stand in der
// Testvorrichtung, und `main.go` startete Wails, ohne sie je zu rufen. Eine
// Zusage, die nur im Test eingelöst wird, ist keine.
func (a *App) Starten(ctx context.Context) error {
	ctx = a.imProfil(ctx)

	// **Ein abgebrochenes Zurückspielen wird zuerst zu Ende geführt**
	// (#1162).
	//
	// **Vor dem Öffnen**, weil es das Bilderverzeichnis betrifft und nicht die
	// Datenbank — und **genau einmal je Prozess**: `Starten` läuft auch
	// mitten in einem Zurückspielen, unmittelbar nach dem Tausch. Warum das
	// zählt, steht in `rueckspielmarke.go`.
	a.rueckspielGeprueft.Do(a.rueckspielVollenden)

	a.startfehler = vollStattEA(a.starten(ctx), filepath.Dir(a.datenbankpfad), freierPlatz)

	// **Der Weg zurück wird gesucht, wenn der Start scheitert** — und nur
	// dann (Werkstück 5 von Phase 4.13, #1352). Er kostet ein Öffnen je
	// Sicherung; im geglückten Fall fragt niemand danach.
	if a.startfehler != nil {
		a.rueckweg = a.wegZurueck(ctx)
	}

	return a.startfehler
}

// starten ist der Start selbst; [App.Starten] hält sein Ergebnis fest.
func (a *App) starten(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(a.datenbankpfad), 0o700); err != nil {
		return fmt.Errorf("Datenverzeichnis anlegen: %w", err)
	}

	// **Erst die liegengebliebenen Reste, dann der Bestand** (OPS04, #2350) —
	// geräumt wird nur, solange niemand sonst ihn hält.
	a.restePflegen(ctx)

	db, err := a.oeffnen(ctx, a.datenbankpfad)
	if err != nil {
		return err
	}

	a.db = db
	a.verdrahten(db)

	if fehler := a.aufDenHeutigenStand(ctx, db); fehler != nil {
		return fehler
	}

	// **Und die Sitzung, wie sie zuletzt war** — ID-02 (#1701): gesperrt
	// bleibt über ⌘Q hinweg gesperrt.
	a.sitzungBeginnen(ctx, a.sitzungsablage)

	// **Das Protokoll nach seinen Einstellungen** (OPS-04, #1721) — erst
	// jetzt, denn sie liegen im Bestand, und erst nach dem Heben kann er sie
	// lesen. **Mit dem gezeigten Profil**, das es erst ab hier gibt: der
	// Kontext von `Starten` entstand vor dem Öffnen und kennt keines, und die
	// Einstellungen liegen je Profil (`docs/EINSTELLUNGEN.md`).
	a.protokollEinrichten(a.imProfil(ctx))

	// **Und die Zwischenspeicher** — OPS-08 (#1725): die Grenze auf der
	// Platte gilt ab dem Start, und die Vorschauen gehören nicht über einen
	// Neubeginn hinweg — nach einem Zurückspielen wäre es ein anderer
	// Bestand.
	a.zwischenspeicherPflegen(a.imProfil(ctx))

	// **Und die Spur, wie sie eingestellt ist** — Befund 8 des Zweitblicks
	// über Phase 6.4: das Fenster sagte immer „aus". Aufgenommen wird hier
	// nichts, denn die Bindungen des Fensters misst niemand; die Auskunft
	// sagt es unter „nicht gemessen".
	a.messungEinrichten(a.imProfil(ctx))
	a.bildfach.Leere()

	// **Und erst jetzt die zweite Verbindung** — Werkstück 16 von Phase 4.12.
	// Warum nicht früher, steht in `nurlesen.go`: `Migriere` beginnt unbedingt
	// eine Transaktion, und die scheitert nur lesend.
	a.leserverbindungRichten(ctx)

	// **Der Paketkatalog wird beim Start aufgefrischt**
	// (#692).
	//
	// Er liegt im Bestand, und gefüllt hat ihn bis dahin nur die **Seite**
	// *Pakete*. Wer die Anwendung zum ersten Mal öffnete, ein Set aufmachte
	// und den Stil wechseln wollte, fand eine leere Auswahl — die Stile lagen
	// die ganze Zeit auf der Platte, nur wusste der Bestand nichts von ihnen.
	//
	// **Beim Start und nicht beim ersten Lesen.** Beim Lesen kostete es dort,
	// wo jemand wartet, und zwei gleichzeitige Leser bräuchten eine Sperre.
	// Hier ist es einmal, an einer Stelle, an der ohnehin noch nichts zu sehen
	// ist.
	//
	// **Und ein Fehlschlag hält den Start nicht auf** (§8).
	//
	// Der erste Anlauf brach hier ab — und stand damit im Widerspruch zu
	// seinem eigenen Kommentar. Wer den Katalog nicht auffrischen kann, kann
	// trotzdem Sets öffnen, Karten ändern und sichern; nur die Stilauswahl
	// stünde auf einem älteren Stand. Eine Anwendung, die deswegen gar nicht
	// erst startet, tauscht einen kleinen Mangel gegen einen großen.
	//
	// **Gemeldet wird er, sobald jemand die Paketseite aufschlägt**: sie
	// frischt selbst auf und zeigt Befunde **und** Fehler. Hier gibt es dafür
	// keinen Weg — `Starten` hat keinen, und einen zu bauen hieße, eine
	// Meldung zu erfinden, die niemand liest.
	// **Der Rückgabewert wird ausdrücklich weggeworfen** — der Grund steht
	// darüber. Ein benannter Empfänger sagt es deutlicher als ein Unterstrich.
	_, katalogfehler := a.pakete.Auffrischen(ctx)
	_ = katalogfehler

	// **Und zuletzt der freiwillige Plan** (OPS01, #2347) — über dem Bestand,
	// der jetzt offen ist; ohne Ordner tut er nichts.
	a.sicherungsplanStarten()

	return nil
}
