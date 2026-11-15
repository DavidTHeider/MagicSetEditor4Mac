// **Eine Sicherung ersetzt den laufenden Bestand** (#940).
//
// Der Gegenweg zu `bestandsicherung.go`, und er ist der schwierigere: sichern
// heißt lesen, zurückspielen heißt, die Datei unter der eigenen Anwendung
// auszutauschen. Deshalb hat er drei Schritte statt einem — **prüfen**,
// **tauschen**, **wieder öffnen** — und einen Rückweg für jeden davon.

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
)

// Sicherungspruefer liest eine Sicherungsdatei, ohne sie zu verändern.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2) — erfüllt von [storage.PruefeSicherung].
type Sicherungspruefer func(ctx context.Context, pfad string) (Sicherungsbefund, error)

// Sicherungsbefund ist das, was eine Sicherung über sich hergibt.
//
// **Er steht hier und nicht in `storage`**, weil er der Vertrag dieser Naht
// ist: `internal/app` darf `storage` außerhalb der Composition Roots nicht
// kennen (§2), und ein Rückgabetyp aus einem Paket, das man nicht kennen darf,
// wäre derselbe Import unter anderem Namen. Erfüllt wird die Naht von
// `storage.PruefeSicherung`; verdrahtet in `app.go`.
type Sicherungsbefund struct {
	// AngelegtAm ist der Zeitpunkt, zu dem der Bestand entstanden ist —
	// **nicht** der Zeitpunkt der Sicherung.
	AngelegtAm string

	// Schemastand ist der letzte Schritt, der in der Datei verbucht ist.
	Schemastand string

	// Sets und Karten sagen, wie viel darin liegt. **−1 heißt „diese Fassung
	// kennt die Tabelle nicht"** und ist etwas anderes als 0.
	Sets   int
	Karten int
}

// MitSicherungspruefer setzt die Naht — für Tests und den Composition Root.
func (b *Bestandsweg) MitSicherungspruefer(pruefer Sicherungspruefer) { b.pruefer = pruefer }

// Alleinpruefung sagt, ob niemand sonst die Bestandsdatei geöffnet hat.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2) — erfüllt von `storage.AlleinAmBestand`.
// **Sie gibt einen Anspruch zurück, nicht nur ein Urteil** (#1163, Befund
// Z6-02; #1174).
//
// Bis Phase 3.2 sagte sie nur ja oder nein — und ließ die Sperre, mit der sie
// geantwortet hat, gleich wieder los. Zwischen ihrer Antwort und dem
// Umbenennen lag ein Fenster; wer es traf, arbeitete auf einem Inode, den es
// gleich nicht mehr gab.
//
// **Die Rückgabe gibt den Anspruch zurück** und muss laufen, bevor der Bestand
// neu geöffnet wird: ein gehaltener Alleinanspruch stünde dem eigenen Öffnen
// im Weg.
type Alleinpruefung func(ctx context.Context, pfad string) (zurueck func() error, fehler error)

// Zurueckspielen ersetzt den Bestand durch die Sicherung.
//
// # Die Reihenfolge ist die Zusage
//
//  1. **Zuerst wird kopiert** ([eingehendeKopie]). Ab hier ist der Gegenstand
//     eine Datei dieser Anwendung, und niemand von außen kann sie noch
//     austauschen.
//  2. **Geprüft wird die Kopie**, obwohl [Bestandsweg.Beschauen] die Vorlage
//     schon prüfte. Zwischen Auswahl und Bestätigung liegt eine Entscheidung
//     eines Menschen, und in dieser Zeit kann die Datei eine andere geworden
//     sein. Eine Prüfung, deren Ergebnis man aufhebt und später verwendet,
//     prüft die Vergangenheit — und eine, die einen **anderen** Gegenstand
//     prüft als den, der ankommt, prüft gar nichts
//     (#964).
//  3. **Erst dann wird geschlossen.** Wer zuerst schließt und dann prüft, hat
//     bei jeder Abweisung eine Anwendung ohne Datenbank.
//  4. **Getauscht wird mit Rückweg** ([bestandErsetzen]).
//  5. **Lässt sich der neue Stand nicht öffnen, geht der Rückweg zurück** — und
//     die Anwendung läuft mit dem alten weiter. Ein Zurückspielen, das
//     scheitert, darf nicht schlimmer enden als eines, das gar nicht stattfand.
//
// # Was sie nicht kann
//
// **Sie fragt nicht nach.** Die Rückfrage steht im Fenster, wo sie hingehört
// (§6) — hier anzuhalten hieße, einen Dialog aus dem Dienst zu öffnen.
//
// **Und sie hält die Vorlage nicht fest.** Wird die fremde Datei getauscht,
// *nachdem* kopiert wurde, bleibt das unbemerkt — dann ist es aber die Kopie,
// die ankommt, und die ist geprüft.
func (b *Bestandsweg) Zurueckspielen(ctx context.Context, pfad string) (string, error) {
	// **Ausgepackt wird vor allem anderen** (#1162): eine Sicherung ist seit
	// 0.3.0 ein Archiv aus Bestand **und** Bildern. Eine nackte Datenbank aus
	// 0.1.0 oder 0.2.0 geht denselben Weg — sie trägt nur keine Bilder.
	ausgepackt, kopierfehler := b.auspacken(ctx, pfad, b.orte.Datenbank)
	if kopierfehler != nil {
		return "", kopierfehler
	}

	defer ausgepackt.Wegraeumen()

	eingehend := ausgepackt.Bestand

	if _, fehler := b.pruefer(ctx, eingehend); fehler != nil {
		return "", fehler
	}

	// **Verglichen wird vor dem Tausch** (O7, #1431), weil es danach nur noch
	// **eine** Seite gibt: der laufende Bestand ist dann der von damals.
	// Geschrieben wird die Epoche unten, in den zurückgespielten hinein.
	//
	// **Ein gescheiterter Vergleich hält den Tausch nicht auf.** Die Rückkehr
	// ist der Zweck, der Vergleich das Sicherheitsnetz — wer deswegen
	// abbräche, ließe jemanden mit einem kaputten Bestand sitzen, weil eine
	// Auskunft darüber nicht zu bekommen war. **Er verschwindet aber auch
	// nicht**: dann steht die Epoche mit `HandpruefungNoetig` da und sagt,
	// dass niemand nachsehen konnte.
	lage := b.lageOderHandpruefung(ctx, eingehend)

	if fehler := b.schliessen(); fehler != nil {
		return "", fmt.Errorf("den Bestand schließen: %w", fehler)
	}

	// **Erst schließen, dann fragen, ob wir allein sind.**
	//
	// Die Reihenfolge ist der Kern: solange die eigene Verbindung offen ist,
	// wäre die Antwort immer „belegt" — von uns selbst.
	//
	// **Und wenn nicht, wird nichts angefasst** (#961):
	// eine Datei umzubenennen, die ein anderer Prozess geöffnet hält, nennt
	// SQLite undefiniert und korruptionsgefährlich. Die Anwendung bekommt
	// ihre Datenbank zurück und läuft weiter.
	anspruch, fehler := b.allein(ctx, b.orte.Datenbank)
	if fehler != nil {
		// **Ohne Abbruch**, wie in [Bestandsweg.wiederOeffnen]: geschlossen
		// ist geschlossen, und ein abgebrochener Kontext ließe die Anwendung
		// ohne Datenbank stehen.
		return "", errors.Join(fehler, b.oeffnen(context.WithoutCancel(ctx)))
	}

	// **Zurückgegeben wird er, bevor neu geöffnet wird** — und zwar
	// **ausdrücklich**, nicht über `defer`.
	//
	// # Warum das der Unterschied zwischen 0,2 und 10,5 Sekunden ist
	//
	// `flock` gilt für die **offene Datei** und nicht für den Prozess. Ein
	// gehaltener Alleinanspruch steht dem **eigenen** Öffnen im Weg: `Starten`
	// geht durch `storage.Oeffnen`, das die Nachbardatei geteilt nimmt — und
	// wartet dann die volle `Sperrfrist` ab, um sie am Ende ohne Sperre zu
	// bekommen.
	//
	// Der erste Anlauf gab ihn im `defer` zurück, also **nach** `Starten`.
	// Sichtbar wurde es allein an der Laufzeit eines Tests: er wurde grün und
	// dauerte zehneinhalb Sekunden.
	//
	// **Der `defer` bleibt als Rückfall** — für die Wege, die vorher
	// hinausgehen. `zurueckgeben` ist deshalb mehrfach aufrufbar.
	zurueckgeben := einmalig(anspruch)
	defer zurueckgeben()

	// **Die Marke steht vor dem ersten Schritt** (#1162, Werkstück 1).
	//
	// Ab hier sind zwei Umbenennungen in zwei Verzeichnissen unterwegs, und
	// kein Dateisystem führt sie zusammen aus. Was geht, ist Zuendeführen: die
	// Marke beschreibt den **zweiten** Schritt vollständig, und ein Start, der
	// sie vorfindet, holt ihn nach. Warum das keine Rückfrage braucht, steht
	// in `rueckspielmarke.go`.
	//
	// **Nur für den zweiteiligen Tausch.** Eine nackte `.db` hat einen
	// Schritt, und der ist für sich unteilbar.
	if ausgepackt.BilderErsetzen {
		if markenfehler := b.markeSchreiben(rueckspielstand{
			Quelle:  pfad,
			Beiwerk: ausgepackt.Beiwerk,
			Bilder:  ausgepackt.Bilderlage(),
		}); markenfehler != nil {
			return "", errors.Join(markenfehler, b.wiederOeffnen(ctx, zurueckgeben))
		}

		// **Weggeräumt wird sie auf jedem Weg hinaus.** Auch auf denen, die
		// zurückdrehen: dann ist der Stand von vorher wiederhergestellt, und
		// eine Marke darüber schickte den nächsten Start auf eine Arbeit, die
		// niemand mehr will.
		defer func() { b.markeAufraeumen(ctx) }()
	}

	// **Die Unsicherheit steht vor dem Tausch** — Befund 10 des Zweitblicks
	// über Phase 5.2. Vermerkt wird sie danach; scheitert das oder stirbt der
	// Prozess dazwischen, holt der nächste Start sie aus dieser Marke nach.
	// Ohne sie lief die Anwendung mit dem neuen Bestand weiter, und der
	// Nachweis für die unklaren Mengen war weg (0088 sagt das Gegenteil zu).
	//
	// **Auch ohne Bildertausch**: die Marke daneben gibt es nur, wenn Bilder
	// dabei sind; die Unsicherheit über Mengen hängt daran nicht.
	if markenfehler := b.epochenmarkeSchreiben(lage, filepath.Base(pfad), b.uhr()); markenfehler != nil {
		return "", errors.Join(markenfehler, b.wiederOeffnen(ctx, zurueckgeben))
	}

	zurueck, fehler := bestandErsetzen(eingehend, b.orte.Datenbank)
	if fehler != nil {
		// **Die Anwendung bekommt ihre Datenbank zurück**, auch wenn der
		// Tausch nicht stattfand — sonst steht sie ohne da, und der nächste
		// Aufruf meldet etwas ganz anderes als den eigentlichen Grund.
		//
		// **Und die Marke geht weg** (Befund 7 des Nachlaufs): ohne Tausch
		// gibt es keine Rückkehr, über die etwas unklar wäre.
		if markenfehler := b.epochenmarkeEntfernen(); markenfehler != nil {
			slog.Default().ErrorContext(ctx, "die Rückkehrmarke blieb liegen",
				"grund", markenfehler.Error())
		}

		return "", errors.Join(fehler, b.wiederOeffnen(ctx, zurueckgeben))
	}

	// **Und erst danach die Bilder.** Scheitert der Tausch der Datenbank,
	// sind sie ohne Bedeutung; scheitert er hier, steht die Datenbank schon —
	// und dann zählt der Rückweg beider Teile.
	//
	// **Und ein Absturz dazwischen wird beim nächsten Start nachgeholt.** Der
	// Rückweg hier läuft dann nicht — er ist Code in einem Prozess, den es
	// nicht mehr gibt. Was bleibt, ist die Marke von oben.
	//
	// **Eine alte Sicherung lässt das Beiwerk stehen.** Sie beschreibt keinen
	// Bilderstand; ihn zu leeren hieße, aus „weiß nichts darüber" ein „da war
	// nichts" zu machen.
	if ausgepackt.BilderErsetzen {
		bilderZurueck, bilderfehler := b.beiwerkErsetzen(ausgepackt.Beiwerk)
		if bilderfehler != nil {
			return "", b.zurueckDrehen(ctx, zurueck, zurueckgeben, bilderfehler)
		}

		// **Erst die Bilder zurück, dann der Bestand** — A05, Befund 3 des
		// Zweitblicks über Phase A: in der umgekehrten Folge ging der alte
		// Bestand auf, während noch das Beiwerk der Sicherung dalag.
		if fehler := b.wiederOeffnen(ctx, zurueckgeben); fehler != nil {
			bilderfehler := rueckweg(bilderZurueck)

			return "", errors.Join(b.zurueckDrehen(ctx, zurueck, zurueckgeben, fehler), bilderfehler)
		}

		b.paketeEinsetzen(ctx, ausgepackt.Paketplatz)

		return b.abschliessen(ctx, lage, pfad)
	}

	if fehler := b.wiederOeffnen(ctx, zurueckgeben); fehler != nil {
		return "", b.zurueckDrehen(ctx, zurueck, zurueckgeben, fehler)
	}

	b.paketeEinsetzen(ctx, ausgepackt.Paketplatz)

	return b.abschliessen(ctx, lage, pfad)
}

// zurueckDrehen holt den alten Bestand zurück, nachdem der neue nicht aufging.
//
// **Der ursprüngliche Grund bleibt vorn.** Wer eine Sicherung einspielt, die
// sich nicht öffnen lässt, will das lesen — und nicht, was beim Aufräumen
// danach noch geschah.
func (b *Bestandsweg) zurueckDrehen(
	ctx context.Context, zurueck func() error, zurueckgeben func(), grund error,
) error {
	// **Die Epochenmarke geht mit zurück** — Befund 7 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Sie steht **vor** dem Tausch, und wer sie
	// nach einem gelungenen Rückweg liegen ließe, erzeugte beim nächsten
	// Start eine Epoche über eine Rückkehr, die es nicht gibt: der Bestand
	// von vorher steht wieder da, und an ihm ist nichts unklar.
	if markenfehler := b.epochenmarkeEntfernen(); markenfehler != nil {
		slog.Default().ErrorContext(ctx, "die Rückkehrmarke blieb nach dem Rückweg liegen",
			"grund", markenfehler.Error())
	}

	if fehler := rueckweg(zurueck); fehler != nil {
		// **Hier endet der Rückweg.** Beide Stände liegen als Datei da; was
		// fehlt, ist der Weg zurück im Programm — und der gehört ins
		// Protokoll, damit ihn jemand von Hand gehen kann.
		slog.Default().ErrorContext(ctx, "der Bestand ließ sich nach einem gescheiterten Zurückspielen nicht wiederherstellen",
			"grund", fehler.Error())

		return errors.Join(grund, fehler)
	}

	return errors.Join(grund, b.wiederOeffnen(ctx, zurueckgeben))
}

// einmalig macht aus einer Rückgabe eine, die mehrfach aufrufbar ist.
//
// **Weil sie an zwei Stellen steht**: ausdrücklich vor jedem Öffnen, und als
// `defer` für die Wege, die vorher hinausgehen. Ein zweiter Aufruf schlösse
// sonst einen Griff, den es nicht mehr gibt.
func einmalig(zurueck func() error) func() {
	var schon bool

	return func() {
		if schon {
			return
		}

		schon = true

		if fehler := zurueck(); fehler != nil {
			slog.Default().Error("die Bestandssperre ließ sich nicht zurückgeben",
				"grund", fehler.Error())
		}
	}
}

// wiederOeffnen gibt den Alleinanspruch zurück und öffnet den Bestand.
//
// **In dieser Reihenfolge und nicht umgekehrt.** Warum, steht oben bei
// `zurueckgeben`.
//
// # Der Kontext des Aufrufers gilt hier nicht mehr
//
// **Weil das Wiederöffnen keine Arbeit auf Verlangen ist, sondern Aufräumen.**
// Ab dem Schließen hat die Anwendung keine Datenbank; wer den Vorgang in
// diesem Augenblick abbricht, will nicht, dass sie **ohne** eine dasteht. Ein
// abgebrochener Kontext ließe `storage.Oeffnen` scheitern, und der Rückweg
// scheiterte gleich mit — beides mit derselben Ursache, und keine davon der
// Grund, aus dem jemand abgebrochen hat.
//
// **`context.WithoutCancel` und keine neue Frist**: der Wert des Aufrufers
// bleibt erhalten (Protokollmerkmale, Sprache), allein das Abbrechen fällt
// weg. Gefunden als Befund 1 des Zweitblicks über Phase 3.8.
func (b *Bestandsweg) wiederOeffnen(ctx context.Context, zurueckgeben func()) error {
	zurueckgeben()

	return b.oeffnen(context.WithoutCancel(ctx))
}
