package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Das Beiwerk eines Sets — alles, was neben der Hauptdatei liegt.
//
// **Eine eigene Datei**, weil `setschreiber.go` an der Zeilengrenze steht (§2)
// — und weil der Schnitt einer Frage folgt: dort steht, wie ein Set
// hinausgeht, hier, was dabei mitkommt.

// Beiwerk kopiert alles, was neben der Hauptdatei zu einem Set gehört.
//
// # Warum das nötig ist
//
// [0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)
// führt die **Hauptdatei** im Bestand mit, nicht die Bilder — die wögen
// Megabyte und gehören nicht in eine Datenbank. Wer ein Set anderswohin
// speichert und nur die Hauptdatei bekommt, hat ein Set ohne Bilder: es öffnet,
// und die Karten sind leer.
//
// **Ein fehlendes Herkunftspaket ist kein Fehler, sondern ein Befund** (§8): die
// Hauptdatei kommt trotzdem heraus, und dass die Bilder fehlen, steht daneben.
//
// # Und die Ablage schlägt das Paket
//
// Ein Bild, das jemand **in dieser Anwendung** eingesetzt hat, liegt neben dem
// Set ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
// Käme das Paket zuletzt, überschriebe es die neue Datei mit der alten — und
// das Speichern nähme die Änderung zurück, die es sichern soll. Deshalb geht
// das Paket **zuerst** hinaus und die Ablage danach.
//
// **Ausgeführt (`Beiwerk`), weil es sich sonst nicht prüfen lässt**, ohne
// einen ganzen Schreiblauf aufzubauen: es beantwortet eine eigene Frage — was
// kommt mit — und die verdient einen eigenen Test.
func Beiwerk(herkunft, ablage, haupt, ziel string) []domain.Befund {
	befunde := ausPaketkopieren(herkunft, haupt, ziel)

	return append(befunde, ausAblageKopieren(ablage, ziel)...)
}

// ausPaketkopieren holt, was im Herkunftspaket liegt.
func ausPaketkopieren(herkunft, haupt, ziel string) []domain.Befund {
	if herkunft == "" {
		// Ein Set, das in der Anwendung entstanden ist, hat kein Herkunftspaket.
		return nil
	}

	paket, err := msepaket.Oeffne(herkunft)
	if err != nil {
		return []domain.Befund{{
			Datei: herkunft,
			Meldung: "das Herkunftspaket ließ sich nicht öffnen; " +
				"Bilder und Symbole sind deshalb nicht mitgekommen",
		}}
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	// **Eine Teilliste ist besser als keine** (§8): wer ein Verzeichnis nicht
	// lesen darf, bekommt die anderen Bilder trotzdem — und erfährt, dass eines
	// fehlt.
	namen, err := paket.Dateien()

	var befunde []domain.Befund

	if err != nil {
		befunde = append(befunde, domain.Befund{
			Datei:   herkunft,
			Meldung: "das Herkunftspaket ließ sich nicht ganz auflisten: " + err.Error(),
		})
	}

	for _, name := range namen {
		if name == haupt {
			continue
		}

		if fehler := kopiere(paket, name, ziel); fehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei:   herkunft,
				Ort:     name,
				Meldung: fehler.Error(),
			})
		}
	}

	return befunde
}

// ausAblageKopieren holt, was jemand in dieser Anwendung eingesetzt hat.
//
// **Ein gewöhnliches Verzeichnis und kein Paket.** Es ist keines: `msepaket`
// liest MSE-Pakete, und die Ablage ist unsere Zutat.
//
// **Ohne Unterverzeichnisse.** [Bildablage] legt flach ab; ein Verzeichnis
// darin käme von außen, und wir kopierten dann etwas, das niemand hier
// hingelegt hat.
func ausAblageKopieren(ablage, ziel string) []domain.Befund {
	if ablage == "" {
		return nil
	}

	eintraege, err := os.ReadDir(ablage)
	if err != nil {
		// **Kein Befund, wenn es das Verzeichnis nicht gibt.** Das ist der
		// Normalfall: in den allermeisten Sets hat niemand ein Bild eingesetzt.
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return []domain.Befund{{
			Datei:   ablage,
			Meldung: "die Bildablage ließ sich nicht lesen: " + err.Error(),
		}}
	}

	var befunde []domain.Befund

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			continue
		}

		if fehler := KopiereDatei(
			filepath.Join(ablage, eintrag.Name()),
			filepath.Join(ziel, eintrag.Name()),
		); fehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei:   ablage,
				Ort:     eintrag.Name(),
				Meldung: fehler.Error(),
			})
		}
	}

	return befunde
}

// KopiereDatei legt eine Datei von der Platte an einen anderen Ort.
//
// **Ausgeführt, weil sich der Fehlerweg sonst nicht prüfen lässt**: eine Datei,
// die zwischen dem Auflisten und dem Lesen verschwindet, ist eine Zeitfrage —
// und ein Test, der auf sie wartet, ist ein Wackelkandidat.
func KopiereDatei(von, nach string) error {
	// #nosec G304 -- der Name kommt aus einem Verzeichniseintrag der Ablage
	inhalt, err := os.ReadFile(von)
	if err != nil {
		return fmt.Errorf("nicht lesbar: %w", err)
	}

	// **`filepath.Base` gegen den Pfaddurchgriff.** Der Name kommt aus einem
	// Verzeichniseintrag; ein Eintrag mit `..` kann so nicht entstehen, aber
	// ein Kopierweg soll sich nicht darauf verlassen, dass ein anderer ihn
	// schützt.
	nach = filepath.Join(filepath.Dir(nach), filepath.Base(nach))

	// **Über eine Nachbardatei** ([schreibeAtomar], Werkstück 7 von Phase
	// 4.13): ein Abbruch mitten im Kopieren ließe sonst ein halbes Bild unter
	// dem richtigen Namen stehen, und die Karte zeigte darauf.
	if err := schreibeAtomar(nach, inhalt); err != nil {
		return fmt.Errorf("nicht schreibbar: %w", err)
	}

	return nil
}

// kopiere legt eine Datei aus dem Paket ins Ziel.
func kopiere(paket *msepaket.Paket, name, ziel string) error {
	inhalt, err := paket.Datei(name)
	if err != nil {
		return fmt.Errorf("nicht lesbar: %w", err)
	}

	// **Der Name kommt aus der Dateiliste des Pakets**, nicht aus einer
	// Eingabe; ein Weg mit `..` kann so nicht entstehen.
	pfad := filepath.Join(ziel, filepath.FromSlash(name))

	if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
		return fmt.Errorf("das Verzeichnis anlegen: %w", err)
	}

	if err := schreibeAtomar(pfad, inhalt); err != nil {
		return fmt.Errorf("nicht schreibbar: %w", err)
	}

	return nil
}

// werteSetzen schreibt die geänderten Werte in ihre Zeilen.
func werteSetzen(knoten *msepaket.Knoten, karte *domain.CardRow) {
	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != msepaket.ArtEintrag {
			continue
		}

		wert, da := wertZurZeile(karte, kind.Schluessel)
		if !da || wieInDerQuelle(wert, kind) {
			continue
		}

		// **Verglichen wird gegen den *wirksamen* Wert, nicht gegen
		// `kind.Wert`** (#704).
		//
		// Bei einem mehrzeiligen Wert steht hinter dem Doppelpunkt nichts; der
		// Wert liegt in den eingerückten Kindern ([wertText]). Der Vergleich
		// gegen das leere `kind.Wert` hielt ihn deshalb bei **jedem** Lauf für
		// geändert — und weil die alten Kinder stehen blieben, stand der Wert
		// danach doppelt in der Datei. Über das kanonische `Vanguard` traf das
		// jede Karte: `notes` verdoppelte sich bei jedem Speichern.
		//
		// **Und der alte Block geht mit, wenn er der Wert war.** Er gehörte zu
		// dem Wert, den wir gerade ersetzen; ihn stehen zu lassen hieße, beide
		// zu behalten. Trug die Zeile ihren Wert dagegen selbst, bleiben ihre
		// Kinder, was sie sind — ein Kommentar unter einer Feldzeile ist kein
		// Feldwert.
		if kind.Wert == "" {
			kind.Kinder = nil
		}

		// **Die Rohzeile fällt weg, und das ist der Punkt.** `msepaket.Schreibe`
		// setzt genau dann zusammen, wenn keine dasteht — mit der Einrückung,
		// die die Tiefe vorgibt.
		kind.Wert = wert
		kind.Roh = ""
	}
}

// zeilenEntfernen wirft die Zeilen der Felder weg, die der Bestand nicht führt.
//
// # Was dabei mitgeht und was nicht
//
// **Der Unterblock geht mit.** Was unter einer Feldzeile eingerückt steht,
// gehört ihr syntaktisch; er bliebe sonst an der Karte hängen, ohne dass ein
// Feld ihn trägt.
//
// **Der Kommentar über einer Zeile bleibt**, und zwar ohne Zutun: der Zerteiler
// hängt ihn an die gerade offene Ebene, also an den Eintrag **darüber**
// (`docs/FORMATE.md`). Er ist damit gar kein Nachbar der entfernten Zeile,
// sondern ein Kind der vorigen.
//
// **Was darunter steht, geht dagegen mit** — auch ein Kommentar. Das ist die
// Kehrseite derselben Regel, und sie gehört ausgesprochen: wer `name` entfernt,
// verliert den Kommentar, der unter `name` steht, selbst wenn er von der Zeile
// darunter handelt. Eine andere Zuordnung zu erfinden hieße, eine Bindung zu
// behaupten, die das Format nicht kennt.
func zeilenEntfernen(kinder []*msepaket.Knoten, karte *domain.CardRow) []*msepaket.Knoten {
	bleibende := make([]*msepaket.Knoten, 0, len(kinder))

	for _, kind := range kinder {
		if kind.Zeilenart == msepaket.ArtEintrag {
			if _, da := wertZurZeile(karte, kind.Schluessel); !da {
				continue
			}
		}

		bleibende = append(bleibende, kind)
	}

	return bleibende
}

// zeilenErgaenzen setzt die Felder ein, für die keine Zeile dasteht.
//
// # Wohin, und woher wir das wissen
//
// **Hinter die Zeile des Vorgängers** — des Feldes, das im Bestand davor
// steht. Der Bestand gibt seine Werte in der Reihenfolge heraus, die
// `card_values.reihenfolge` führt, und seit
// #160
// überlebt sie eine Rücknahme. Das ist die einzige Stelle, die sich aus dem
// ableiten lässt, was dasteht, statt sie zu erfinden.
//
// Hat der Vorgänger selbst keine Zeile, wird weiter zurückgesehen; hat keiner
// eine, kommt sie an den **Anfang** des Kartenblocks.
func zeilenErgaenzen(kinder []*msepaket.Knoten, karte *domain.CardRow) []*msepaket.Knoten {
	for stelle, wert := range karte.Values {
		if zeileVon(kinder, wert.Field) >= 0 {
			continue
		}

		// **Rückwärts, nicht vorwärts.** Der nächste Vorgänger mit einer Zeile
		// ist der, hinter den es gehört; ein Nachfolger sagte nur, wovor es
		// nicht stehen darf.
		hinter := -1

		for zurueck := stelle - 1; zurueck >= 0 && hinter < 0; zurueck-- {
			hinter = zeileVon(kinder, karte.Values[zurueck].Field)
		}

		neue := &msepaket.Knoten{
			Zeilenart: msepaket.ArtEintrag,
			// **Kanonisch geschrieben** — so schreibt MSE selbst
			// (`canonical_name_form`, #396).
			Schluessel: msepaket.Kanonisch(wert.Field), Wert: wert.Text,
			Zeilenende: zeilenendeVon(kinder),
		}

		// **Eine Zeile, die einen Nachfolger bekommt, braucht ein Ende.**
		//
		// Die letzte Zeile einer Datei darf ohne Umbruch enden
		// (`docs/FORMATE.md`). Wer hinter sie einfügt, ohne ihr eines zu geben,
		// klebt beide zusammen: aus `name: Feuerball` und `farbe: rot` wurde
		// `name: Feuerball\tfarbe: rot`. Die neue Zeile übernimmt dafür das
		// Ende, das die alte hatte — auch das leere, denn jetzt ist sie die
		// letzte.
		if hinter >= 0 {
			neue.Zeilenende = kinder[hinter].Zeilenende
			kinder[hinter].Zeilenende = zeilenendeVon(kinder)
		}

		kinder = slices.Insert(kinder, hinter+1, neue)
	}

	return kinder
}

// zeileVon sucht die Zeile eines Feldes; −1, wenn es keine gibt.
// **Kanonisch verglichen** (#396): in der Datei steht `rule_text`, im Modell
// `rule text`. Wer wörtlich vergliche, fände die Zeile nicht und schriebe eine
// zweite daneben — dieselbe Karte mit zwei Regeltexten.
// wertZurZeile sucht den Bestandswert zu einer Zeile der Quelldatei — **über
// die kanonische Form**, wie [zeileVon] in der Gegenrichtung.
//
// Die Datei schreibt `rule_text:`, der Bestand führt `rule text`
// ([msepaket.Entkanonisch]). Ein Vergleich der rohen Schlüssel fand für jedes
// Feld mit Leerzeichen im Namen keinen Wert: die Zeile fiel weg und kam aus
// dem Bestand neu — mit anderen Abständen, ohne ihren Kommentar und bei
// `extra_data:` in der flachen Fassung eines älteren Bestands (#2298).
func wertZurZeile(karte *domain.CardRow, schluessel string) (string, bool) {
	gesucht := msepaket.Kanonisch(schluessel)

	for _, wert := range karte.Values {
		if msepaket.Kanonisch(wert.Field) == gesucht {
			return wert.Text, true
		}
	}

	return "", false
}

func zeileVon(kinder []*msepaket.Knoten, feld string) int {
	gesucht := msepaket.Kanonisch(feld)

	for stelle, kind := range kinder {
		if kind.Zeilenart == msepaket.ArtEintrag &&
			msepaket.Kanonisch(kind.Schluessel) == gesucht {
			return stelle
		}
	}

	return -1
}

// zeilenendeVon nimmt das Zeilenende, das im Kartenblock schon benutzt wird.
//
// **Eine Datei trägt beide in derselben Datei** (`docs/FORMATE.md`), und eine
// neue Zeile mit dem falschen wäre ein Unterschied, den niemand angeordnet hat.
// Ohne Anhaltspunkt gilt `\n`.
func zeilenendeVon(kinder []*msepaket.Knoten) string {
	for _, kind := range kinder {
		if kind.Zeilenende != "" {
			return kind.Zeilenende
		}
	}

	return "\n"
}
