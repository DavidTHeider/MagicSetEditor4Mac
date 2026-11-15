package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// mitZusatzwert setzt einen Wert in den Block einer Stilvorlage unter
// `extra_data:` und lässt alles andere, wie es dasteht.
//
// **Über den Baum und nicht über Zeichenketten**: der Zerteiler kennt die
// Einrückung, und [msepaket.Schreibe] setzt nur zusammen, was keine
// Rohzeile mehr trägt — die Blöcke anderer Vorlagen kommen Byte für Byte
// zurück, wie MSE2 sie ungelesen zurückschreibt
// (`Writer::handle(DelayedIndexMapsData)`, `src/util/delayed_index_maps.hpp`,
// Zeilen 59–71, Fassung f9d9356 vom 2021-07-06).
//
// **Ein Text, der sich nicht lesen lässt, wird nicht überschrieben** — wer
// daran setzte, verlöre, was darin steht.
func mitZusatzwert(text string, stil msepaket.Stil, schluessel, wert string) (string, error) {
	// **Jede Zeile mit Ende**: die letzte trägt im Bestand keines, und eine
	// neue dahinter klebte an ihr.
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	baum, err := msepaket.Zerteile(zusatzdatenfeld, []byte(text))
	if err != nil {
		// **Der Satz für Menschen, der Grund fürs Protokoll** (0075): die
		// Meldung des Zerteilers nennt Zeile und Spalte eines Texts, den
		// niemand als Datei vor sich hat.
		return "", domain.MitGrund(domain.NewFieldError(msepaket.Entkanonisch(zusatzdatenfeld),
			domain.SchluesselValidierung, "Die Werte der Zusatzfelder dieser Karte lassen sich nicht lesen."),
			err, "die Zusatzdaten zerteilen")
	}

	// **Die flache Altform kommt dabei in ihre Gestalt** — sonst stünde das
	// neue Feld unter dem Block und die alten daneben.
	altformEinordnen(baum, stil.Zusatzfelder)

	vorlage := stilvorlagenname(stil.Paket)

	block := eintragNamens(baum.Wurzel, vorlage)
	if block == nil {
		block = &msepaket.Knoten{Zeilenart: msepaket.ArtEintrag, Schluessel: vorlage}
		baum.Wurzel = append(baum.Wurzel, block)
	}

	zeile := eintragNamens(block.Kinder, schluessel)
	if zeile == nil {
		zeile = &msepaket.Knoten{Zeilenart: msepaket.ArtEintrag, Schluessel: schluessel}
		block.Kinder = append(block.Kinder, zeile)
	}

	// **Die Rohzeile fällt weg, und mit ihr ein alter Block darunter** —
	// dieselbe Regel wie beim Kartenwert (`werteSetzen`).
	zeile.Wert, zeile.Roh, zeile.Kinder = wert, "", nil

	return strings.TrimSuffix(string(msepaket.Schreibe(baum)), "\n"), nil
}

// eintragNamens sucht einen Eintrag über seinen kanonischen Schlüssel.
func eintragNamens(knoten []*msepaket.Knoten, schluessel string) *msepaket.Knoten {
	for _, kind := range knoten {
		if kind.Zeilenart == msepaket.ArtEintrag && msepaket.Kanonisch(kind.Schluessel) == schluessel {
			return kind
		}
	}

	return nil
}
