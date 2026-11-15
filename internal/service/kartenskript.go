package service

// **Ein Kartentext, der im Spiel wirkt** — Werkstück M8 von Phase 5.5
// (#1610).
//
// # Die Naht, um derentwillen es dieses Werkstück gibt
//
// **Zerteiler und Auswerter liegen seit Phase 3 in `internal/msescript`** —
// hier bekommen sie die **Spielwelt** als Umgebung. Das Issue sagt es so,
// und es ist der Punkt: die Sprache gab es, den Gegenstand nicht.
//
// # Der Sandkasten, und wo seine Wände stehen
//
// **Ein Skript ändert den Zustand über benannte Wege und über keinen
// anderen.** Es bekommt dreizehn Funktionen, und jede schreibt eine
// [domain.Spielwirkung] **auf**; anwenden kann sie nur die Domäne
// ([domain.Partie.Anwenden]). Es gibt keinen Weg von einem Kartentext zu
// einem Feld der Partie, der an dieser Liste vorbeiführte.
//
// **Und es bekommt keine Datei, kein Netz und keine Uhr.** Die
// Standardbibliothek aus `msescript` trägt nichts davon — `write_image_file`
// steht ausdrücklich nicht darin, und `internal/schichten` hält es so.
//
// **Ein Skript ohne Ende wird abgebrochen.** Der Auswerter prüft den Kontext
// an jedem Knoten (§7), und dieser hier gibt ihm eine **Frist**. Dazu kommt
// [msescript.Tiefengrenze] gegen `a := {a()}`. Was abbricht, wird ein
// **Befund** und hält die Partie nicht an (§8).
//
// # Wann ausgewertet wird
//
// **Beim Ankündigen und nicht beim Verrechnen.** Der Grund und die
// Abweichung, die darin liegt, stehen bei [domain.Spielobjekt.Wirkungen].

import (
	"context"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Skriptfrist ist, wie lange ein Kartenskript rechnen darf.
//
// **Zweihundert Millisekunden**, und die Zahl ist eine Entscheidung: ein
// Kartentext rechnet nichts, wofür ein Rechner länger braucht, und ein Tisch,
// der eine Sekunde steht, ist kaputt. Wer sie ändert, ändert sie hier — und
// nicht an fünf Rufstellen.
const Skriptfrist = 200 * time.Millisecond

// Skriptumfeld sagt, wer die Fähigkeit beherrscht und worauf sie zielt.
//
// **Drei Namen und keine Partie**: ein Skript soll den Zustand nicht lesen
// können, den es nicht ändern darf. Was es sieht, steht hier.
type Skriptumfeld struct {
	// Beherrscher ist, wem die Fähigkeit gehört — CR 113.8.
	Beherrscher string

	// Objekt ist die Quelle — CR 113.7.
	Objekt string

	// Ziele sind die genannten Ziele — CR 115.1.
	Ziele []string
}

// Kartenskript wertet das Spielskript einer Karte aus.
type Kartenskript struct{ frist time.Duration }

// NeuesKartenskript gibt einen Auswerter mit der Vorgabefrist.
func NeuesKartenskript() *Kartenskript {
	return &Kartenskript{frist: Skriptfrist}
}

// MitFrist setzt eine andere Frist — für Läufe, die den Abbruch messen.
func (k *Kartenskript) MitFrist(frist time.Duration) *Kartenskript {
	k.frist = frist

	return k
}

// Deuten gibt die Wirkungen eines Skripts.
//
// **Ein leeres Skript gibt nichts und ist kein Fehler**: die allermeisten
// Karten tragen keines (§8).
//
// **Ein Fehler ist ein Befund und kein Abbruch der Partie.** Der Aufrufer
// legt ihn in den Strom; was das Skript bis dahin aufgeschrieben hat, fällt
// weg — eine halb ausgeführte Fähigkeit wäre schlimmer als keine (§10).
func (k *Kartenskript) Deuten(
	ctx context.Context, quelle string, umfeld Skriptumfeld,
) ([]domain.Spielwirkung, error) {
	if quelle == "" {
		return nil, nil
	}

	befristet, schluss := context.WithTimeout(ctx, k.frist)
	defer schluss()

	sammler := &wirkungssammler{umfeld: umfeld}

	if _, fehler := msescript.AuswertenQuelle(
		befristet, quelle, sammler.umgebung()); fehler != nil {
		// **Der Satz für Menschen, der Grund fürs Protokoll** —
		// [domain.MitGrund] ([0075](../../docs/entscheidungen/0075-der-satz-und-sein-grund.md)).
		// Ein Skriptfehler nennt Zeile und Spalte in einem Text, den
		// jemand vor Wochen geschrieben hat; am Tisch hilft davon nichts.
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"das Kartenskript von %q ist nicht durchgelaufen",
			umfeld.Objekt)
	}

	return sammler.wirkungen, nil
}

// wirkungssammler nimmt auf, was ein Skript bewirken will.
type wirkungssammler struct {
	umfeld    Skriptumfeld
	wirkungen []domain.Spielwirkung
}

// umgebung baut die Spielwelt, die ein Skript sieht.
//
// **Über der Standardbibliothek** (`Darueber`), damit `if`, `+` und
// `to_string` dastehen — und damit ein Skript sie **nicht** überschreiben
// kann: eine eigene Zuweisung legt einen Namen in der Kindebene an und lässt
// die Bibliothek unberührt.
func (w *wirkungssammler) umgebung() *msescript.Umgebung {
	umgebung := msescript.Bibliothek().Darueber()

	// **Drei Namen aus dem Umfeld** — CR 113.8 (wer), CR 113.7 (was) und
	// CR 115.1 (worauf).
	umgebung.Setze("you", msescript.Text(w.umfeld.Beherrscher))
	umgebung.Setze("this", msescript.Text(w.umfeld.Objekt))
	umgebung.Setze("target", msescript.Text(w.erstesZiel()))

	for name, art := range skriptfunktionen {
		umgebung.Setze(name, msescript.NeueEingebaute(w.baue(art)))
	}

	return umgebung
}

// erstesZiel gibt das erste genannte Ziel — leer, wo keines steht.
func (w *wirkungssammler) erstesZiel() string {
	if len(w.umfeld.Ziele) == 0 {
		return ""
	}

	return w.umfeld.Ziele[0]
}

// skriptfunktionen sind die dreizehn benannten Wege, je Wirkungsart einer.
//
// **Englisch**, weil MSE-Script eine fremde Sprache ist (§9, die Tabelle der
// Bezeichner): wer `to_string` schreibt, schreibt daneben `draw`.
var skriptfunktionen = map[string]domain.Wirkungsart{
	"draw":        domain.WirkungZiehen,
	"change_life": domain.WirkungLeben,
	"deal_damage": domain.WirkungSchaden,
	"destroy":     domain.WirkungZerstoeren,
	"exile":       domain.WirkungExilieren,
	"put_counter": domain.WirkungMarke,
	"mill":        domain.WirkungMahlen,
	"tap":         domain.WirkungTappen,
	"untap":       domain.WirkungEnttappen,

	// **Zwei aus `M10`** (#1612):
	// sie legen einen **Dauereffekt** an statt einer einmaligen Änderung —
	// der Unterschied ist CR 611.1 gegen CR 610.1.
	"boost": domain.WirkungWerte,
	"grant": domain.WirkungFaehigkeit,

	// **Zwei aus `M11`** (#1613):
	// sie kommen nicht auf den Stapel, sie **ändern, was geschieht**
	// (CR 614.1, CR 615.1).
	"exile_instead": domain.WirkungStattdessen,
	"prevent":       domain.WirkungVerhindern,
}

// Skriptname gibt die Funktion, unter der eine Wirkungsart zu rufen ist.
//
// **Exportiert für den Wächter**, der beide Listen gegeneinander hält: eine
// Wirkungsart ohne Funktion wäre eine Fähigkeit, die kein Skript rufen kann
// (§10). Leer heißt: es gibt keine.
func Skriptname(art domain.Wirkungsart) string {
	gefunden := ""

	for name, eine := range skriptfunktionen {
		if eine == art {
			gefunden = name
		}
	}

	return gefunden
}

// baue macht aus einer Wirkungsart eine eingebaute Funktion.
//
// **Eine Funktion für dreizehn Arten**, weil die Unterschiede in den Feldern
// liegen und nicht im Ablauf — dieselbe Überlegung wie bei der
// Handlungsleiste im Fenster.
func (w *wirkungssammler) baue(art domain.Wirkungsart) msescript.Eingebaut {
	return func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		wirkung := domain.Spielwirkung{Art: art, Menge: 1}

		wirkung.Wer = w.textAus(ruf, "player", w.umfeld.Beherrscher)
		wirkung.Objekt = w.textAus(ruf, "object", w.erstesZiel())
		wirkung.Marke = w.textAus(ruf, "counter", domain.MarkePlusEins)

		// **`keyword:` schreibt dasselbe Feld** — bei
		// [domain.WirkungFaehigkeit] steht dort das Schlüsselwort, und ein
		// zweites Feld daneben wäre eines, das zehn Arten leer mitführten.
		if wort, da := ruf.Benannter("keyword"); da {
			wirkung.Marke = wort.AlsText()
		}

		if menge, da := zahlAus(ruf, "amount"); da {
			wirkung.Menge = menge
		}

		w.wirkungen = append(w.wirkungen, wirkung)

		return msescript.Text(string(art)), nil
	}
}

// textAus liest einen benannten Text — oder gibt die Vorgabe.
func (w *wirkungssammler) textAus(
	ruf msescript.Ruf, name, vorgabe string,
) string {
	wert, da := ruf.Benannter(name)
	if !da {
		return vorgabe
	}

	return wert.AlsText()
}

// zahlAus liest eine benannte Zahl.
//
// **Und eine, die keine ist, ist keine Zahl und kein Fehler** (§8): ein
// Kartentext, der `amount: "viele"` schreibt, bekommt die Vorgabe. Ein
// Abbruch wäre eine Partie, die an einem Tippfehler stehen bleibt.
func zahlAus(ruf msescript.Ruf, name string) (int, bool) {
	wert, da := ruf.Benannter(name)
	if !da {
		return 0, false
	}

	if wert.Art != msescript.ArtZahlwert {
		return 0, false
	}

	return int(wert.Zahl), true
}

// SkriptBefund macht aus einem Fehler den Satz, der am Tisch steht.
//
// **Ein Befund und kein Fehler** (§8): die Partie läuft weiter, und was nicht
// geschehen ist, wird **benannt**.
func SkriptBefund(fehler error) string {
	if fehler == nil {
		return ""
	}

	if errors.Is(fehler, context.DeadlineExceeded) {
		return "das Kartenskript hat die Frist überschritten und wurde " +
			"abgebrochen"
	}

	return fehler.Error()
}
