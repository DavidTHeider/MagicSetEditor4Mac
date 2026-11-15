package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Bilder in der Zwischenablage** —
// #1237.
//
// # Der Befund
//
// Die Ablage überträgt **Feldwerte**. Ein Bildfeld trägt aber keinen
// Bildinhalt, sondern einen **paketlokalen Namen** — `image: image1` —, und
// dieser Name ist im Bestand je Set aufgelöst: `<wurzel>/<setID>/image1`.
//
// Wer eine Karte aus Set A nach Set B kopierte, bekam dort einen Verweis auf
// `<wurzel>/B/image1`: eine Datei, die **nicht existiert** oder, schlimmer,
// ein **fremdes Bild**, das zufällig gleich heißt.
//
// # Was MSE tut
//
//	WITH_DYNAMIC_ARG(clipboard_package, &package);
//
// `serialize_for_clipboard`, `src/data/format/clipboard.cpp`, Zeilen 23–30,
// Fassung f9d9356 vom 2021-07-06: das **Quellpaket** geht mit, und beim Lesen
// werden paketlokale Namen ins Zielpaket **kopiert**
// (`src/util/io/package.cpp`, Zeile 311).
//
// # Warum hier Base64 und kein Paket
//
// MSEs Ablage ist ein Paketkontext; unsere ist **Text** — sie geht durch die
// Zwischenablage des Systems, und sie soll dort für MSE2 lesbar bleiben. Ein
// `image_data:`-Block ist ein Schlüssel, den MSE2 nicht kennt und deshalb
// überspringt (`Reader::unknownKey`, `src/util/io/reader.cpp`, Zeilen
// 231–246).
//
// **Der Preis steht dazu**: ein Kunstbild wiegt Hunderte Kilobyte, und Base64
// macht daraus ein Drittel mehr. Wer zwanzig Karten mit Bildern kopiert, legt
// einige Megabyte in die Zwischenablage.

// bilddatenschluessel ist der Block, unter dem ein Bild im Ablagetext steht.
//
// **Ein eigener Block je Datei und nicht je Feld**: zwei Felder derselben
// Karte können auf dieselbe Datei zeigen, und zweimal dieselben Bytes wären
// zweimal die Größe.
const bilddatenschluessel = "image_data"

// MitBildern setzt die Bildablage — ohne sie geht nur der Verweis mit.
//
// **Eine Naht und kein Pflichtfeld**: die Flächen (`ablagenahtflaechen.go`)
// haben kein Bilderverzeichnis, und dort ist ein Verweis ohne Bild richtig —
// besser als gar kein Kopieren.
func (k *Kartenablage) MitBildern(bilder *Bildablage) *Kartenablage {
	k.bilder = bilder

	return k
}

// MitFeldern sagt, welche Felder ein Zielset **speichert**.
//
// **Ohne sie schlägt jedes Einfügen einer echten Karte fehl** — und zwar so,
// dass eine leere Karte zurückbleibt. Im laufenden Fenster nachgestellt am
// 2026-08-28: eine Karte aus `Vanguard.mse-set` in ein anderes Set eingefügt
// gab „das Feld \"time created\" wird von der Vorlage berechnet und nicht
// gespeichert" — und `karten: ["336"]`, also eine angelegte Karte ohne Werte.
//
// # Warum das überhaupt passiert
//
// Ein `.mse-set` **speichert** `time created` und `time modified`; sie stehen
// im Kartenblock, und `Kopieren` nimmt sie mit. Die Vorlage **rechnet** sie
// aber, und `FelderVonAussen` weist deshalb jeden Schreibversuch ab — zu Recht:
// ein von außen gesetzter Zeitstempel wäre eine Behauptung über eine Karte,
// die es gerade erst gibt.
//
// **MSE2 behält sie beim Einfügen** (`CardsDataObject` spiegelt die ganze
// `Card`, samt `time_created`). Das ist hier eine bewusste Abweichung (§15,
// Regel 4): eine eingefügte Karte ist **neu**, und ihr Entstehungszeitpunkt
// ist jetzt. Was nicht mitgeht, steht als Befund daneben.
//
// **Wahlfrei**, wie die Bilder: ohne diese Naht bleibt es beim alten
// Verhalten, und die Läufe, die keine Vorlage haben, brauchen keine.
func (k *Kartenablage) MitFeldern(felder Kartenfeldquelle) *Kartenablage {
	k.felder = felder

	return k
}

// bilderAnhaengen legt die Dateien, auf die die Karten zeigen, in den Baum.
//
// **Gesammelt über alle Karten, nicht je Karte.** Wer zehn Karten mit
// demselben Rahmenbild kopiert, legt es einmal mit.
//
// **Und was hier nicht liegt, geht als bloßer Verweis mit** — mit einem
// Befund. Die Bilder eines geöffneten Sets liegen in **seinem Herkunftspaket**
// und nicht in dieser Ablage ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md));
// sie mitzukopieren wären bei `Multicolor.mse-set` 229 MB durch die
// Zwischenablage des Systems.
//
// Der Verweis allein ist dabei **gefährlich still**: `image0` heißt im
// Zielset etwas anderes, und dort stünde dann ein **fremdes** Bild. Deshalb
// sagt das Kopieren es, statt es geschehen zu lassen — im laufenden Fenster
// gesehen am 2026-08-28, nicht überlegt.
func (k *Kartenablage) bilderAnhaengen(
	setID int64, karten []*domain.CardRow, baum *msepaket.Baum,
) {
	if k.bilder == nil {
		return
	}

	ort := k.bilder.Verzeichnis(setID)
	if ort == "" {
		return
	}

	gesehen := map[string]bool{}

	for _, karte := range karten {
		for _, wert := range karte.Values {
			name := strings.TrimSpace(wert.Text)
			if name == "" || gesehen[name] || !bildname(name) {
				continue
			}

			gesehen[name] = true

			// **Ein fehlendes Bild hält das Kopieren nicht auf** (§8). Der
			// Verweis geht mit; im Zielset fehlt dann dieselbe Datei wie hier,
			// und das ist die ehrliche Übertragung des Zustands.
			roh, err := os.ReadFile(filepath.Join(ort, name)) // #nosec G304 -- Name aus dem eigenen Verzeichnis
			if err != nil {
				continue
			}

			baum.Wurzel = append(baum.Wurzel, &msepaket.Knoten{
				Schluessel: bilddatenschluessel,
				Wert:       name + " " + base64.StdEncoding.EncodeToString(roh),
			})
		}
	}
}

// verweiseOhneBild nennt die Bildwerte, zu denen kein Bild mitgekommen ist.
//
// # Warum das ein Befund ist und kein Schweigen
//
// Die Bilder eines geöffneten Sets liegen in **seinem Herkunftspaket** und
// nicht in dieser Anwendung
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md));
// sie mitzukopieren wären bei `Multicolor.mse-set` 229 MB durch die
// Zwischenablage des Systems. Der Verweis geht deshalb allein mit — und das
// ist **gefährlich still**: `image0` heißt im Zielset etwas anderes, und dort
// stünde dann ein **fremdes** Bild.
//
// Im laufenden Fenster gesehen am 2026-08-28, nicht überlegt: eine Karte aus
// `Vanguard.mse-set` kam mit `image0` im Zielset an, und dort zeigte der Name
// auf ein anderes Bild.
//
// **Über die Feldart und nicht über die Gestalt des Werts.** [bildname] ist
// eine Heuristik für das Kopieren, wo ein Fehlgriff nichts kostet: die Datei
// gibt es dann nicht, und der Wert geht unverändert mit. Hier kostet er einen
// **falschen Satz** — `2020-01-01` sieht aus wie ein Bildname, ist aber ein
// `time created`. Deshalb entscheidet die Vorlage.
//
// **Ohne Feldquelle sagt diese Stelle nichts.** Sie kann es dann nicht wissen,
// und ein Befund auf Verdacht wäre schlimmer als keiner.
func verweiseOhneBild(
	formen []domain.Field, bloecke []*msepaket.Knoten, umbenannt map[string]string,
) []domain.Befund {
	bildfeld := map[string]bool{}
	for _, form := range formen {
		bildfeld[msepaket.Kanonisch(form.Name)] = form.Kind == domain.KindImage
	}

	var (
		gesehen  = map[string]bool{}
		verwaist []string
	)

	for _, block := range bloecke {
		for _, kind := range block.Kinder {
			name := strings.TrimSpace(wertText(kind))
			if name == "" || gesehen[name] || !bildfeld[msepaket.Kanonisch(kind.Schluessel)] {
				continue
			}

			gesehen[name] = true

			if _, mitgekommen := umbenannt[name]; !mitgekommen {
				verwaist = append(verwaist, name)
			}
		}
	}

	if len(verwaist) == 0 {
		return nil
	}

	return []domain.Befund{{
		Meldung: "zu diesen Bildern ist keine Datei mitgekommen — sie liegen " +
			"im Herkunftspaket des Quellsets, und im Zielset kann derselbe " +
			"Name ein anderes Bild meinen: " + strings.Join(verwaist, ", "),
	}}
}

// bildname sagt, ob ein Feldwert wie ein paketlokaler Bildname aussieht.
//
// **Über den Wert und nicht über die Feldart**, und der Grund ist die Naht:
// die Ablage kennt die Kartenfelder des Spiels nicht — sie bekommt Karten und
// Köpfe. Sie um eine Feldquelle zu erweitern, hieße jeden Aufrufer zu ändern,
// darunter die drei Flächen, die keine hat.
//
// **Was das kostet, gehört gesagt**: ein Textfeld, dessen Wert zufällig wie
// ein Bildname aussieht **und** für das eine gleichnamige Datei im
// Bilderverzeichnis liegt, würde mitkopiert. Der Wert wäre danach derselbe;
// es entstünde eine überzählige Datei im Zielset und sonst nichts.
func bildname(wert string) bool {
	if strings.ContainsAny(wert, " \t/\\") {
		return false
	}

	// **Wie `Bildablage.Ablegen` sie vergibt**: Feldname klein, dann eine Zahl.
	ziffern := 0
	for i := len(wert) - 1; i >= 0 && wert[i] >= '0' && wert[i] <= '9'; i-- {
		ziffern++
	}

	return ziffern > 0 && ziffern < len(wert)
}

// istEinfacherName sagt, ob ein Name eine Datei **in** einem Verzeichnis
// benennt — und nicht einen Weg dorthin.
//
// **Abgewiesen wird alles mit einem Trenner, `..` und der leere Name.**
// `filepath.Base` allein genügte nicht: es machte aus `../1/bild` ein `bild`
// und nähme damit stillschweigend etwas an, was jemand anders gemeint hat.
// Wer einen Weg schickt, bekommt einen Befund.
func istEinfacherName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	return !strings.ContainsAny(name, `/\`) && !strings.Contains(name, "..")
}

// bilderUebernehmen schreibt die mitgeschickten Bilder ins Zielset und gibt
// die Zuordnung „alter Name → neuer Name".
//
// **Ein neuer Name und nicht derselbe**: im Zielset kann `image1` schon belegt
// sein, und ihn zu überschreiben nähme einer anderen Karte ihr Bild.
func (k *Kartenablage) bilderUebernehmen(
	setID int64, baum *msepaket.Baum,
) (map[string]string, []domain.Befund) {
	bloecke := baum.Alle(bilddatenschluessel)
	if len(bloecke) == 0 || k.bilder == nil {
		return nil, nil
	}

	umbenannt := make(map[string]string, len(bloecke))

	var befunde []domain.Befund

	for _, block := range bloecke {
		name, roh, ist := strings.Cut(strings.TrimSpace(block.Wert), " ")
		if !ist {
			continue
		}

		// **Der Name kommt von außen** — aus der Zwischenablage, also aus
		// einem Text, den irgendjemand geschrieben hat. Ungeprüft
		// weitergereicht legte `image_data: ../1/bild <base64>` eine Datei im
		// Verzeichnis eines **anderen** Sets an: `filepath.Join` räumt `..`
		// auf, statt es abzuweisen.
		//
		// Angestrichen vom Nachlauf über die Nacharbeit
		// (#1282,
		// Befund 3). Nachgestellt hat es sich, sobald das Nachbarverzeichnis
		// existiert — im Betrieb also, sobald das erste Set ein Bild führt.
		if !istEinfacherName(name) {
			befunde = append(befunde, domain.Befund{
				Datei: name,
				Meldung: "der Name des mitgeschickten Bildes ist kein " +
					"einfacher Dateiname — er wird nicht übernommen",
			})

			continue
		}

		daten, err := base64.StdEncoding.DecodeString(roh)
		if err != nil {
			befunde = append(befunde, domain.Befund{
				Datei:   name,
				Meldung: "das mitgeschickte Bild ist nicht lesbar: " + err.Error(),
			})

			continue
		}

		neu, ablegefehler := k.bilder.Bytes(setID, name, daten)
		if ablegefehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei:   name,
				Meldung: "das Bild ließ sich nicht ablegen: " + ablegefehler.Error(),
			})

			continue
		}

		umbenannt[name] = neu
	}

	return umbenannt, befunde
}
