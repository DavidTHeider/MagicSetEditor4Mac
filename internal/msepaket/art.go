package msepaket

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Art ist die Sorte eines Pakets, erkennbar an der Namensendung.
//
// Die sieben Endungen sind belegt (`docs/FORMATE.md`): 528 Pakete einer
// öffentlichen Sammlung tragen sämtlich eine davon.
type Art string

const (
	ArtUnbekannt     Art = ""
	ArtSet           Art = "mse-set"
	ArtSpiel         Art = "mse-game"
	ArtStil          Art = "mse-style"
	ArtSymbolschrift Art = "mse-symbol-font"
	ArtEinschub      Art = "mse-include"
	ArtSprache       Art = "mse-locale"
	ArtExportvorlage Art = "mse-export-template"
)

// arten ordnet jeder Endung ihre Art zu.
var arten = map[string]Art{
	"mse-set":             ArtSet,
	"mse-game":            ArtSpiel,
	"mse-style":           ArtStil,
	"mse-symbol-font":     ArtSymbolschrift,
	"mse-include":         ArtEinschub,
	"mse-locale":          ArtSprache,
	"mse-export-template": ArtExportvorlage,
}

// Arten nennt die sieben Paketarten — das Set zuerst, dann die Vorlagen.
//
// **Eine Liste und keine Zuordnung**, weil die Operationsmatrix in
// `docs/FAEHIGKEITEN.md` jede Art braucht (FMT-01,
// #1690):
// ein Wächter vergleicht beide, und eine achte Art erzwingt dort eine Zeile.
func Arten() []Art {
	return []Art{
		ArtSet, ArtSpiel, ArtStil, ArtSymbolschrift, ArtEinschub, ArtSprache,
		ArtExportvorlage,
	}
}

// ArtAusName liest die Art aus dem Paketnamen.
//
// Über `path.Ext` und nicht über `filepath.Ext`: Paketnamen stehen auch in
// Verweisen **innerhalb** von Paketen (`/magic-modules.mse-include/…`), und
// dort trennt immer der Schrägstrich — unter Windows würde `filepath` mit `\`
// arbeiten und daran vorbeigreifen (`docs/PORTIERUNG.md`).
func ArtAusName(name string) Art {
	endung := strings.TrimPrefix(path.Ext(path.Base(name)), ".")

	return arten[endung]
}

// Hauptdatei ist der Name der Datei, die den Kopf eines Pakets trägt.
//
// Die Regel lautet: Endung ohne das Präfix `mse-`. Belegt an zwei Paketen —
// `magic.mse-game/game` und `magic-m15.mse-style/style` (`docs/FORMATE.md`);
// für die übrigen fünf Arten ist sie **Vermutung**. Findet sich ein Paket, in
// dem sie nicht gilt, wird das Öffnen scheitern und die Meldung sagt, welche
// Datei gesucht wurde — das ist besser, als still eine falsche zu lesen.
func (a Art) Hauptdatei() string {
	if a == ArtUnbekannt {
		return ""
	}

	return strings.TrimPrefix(string(a), "mse-")
}

// Datum ist eine Paketversion.
//
// MSE-Versionen sind **Datumsangaben** (`2024-05-26`), keine Zahlenfolgen —
// belegt in `docs/FORMATE.md`. Als eigener Typ und nicht als Zeichenkette: ein
// Vergleich von ISO-Daten als Text geht nur deshalb gut, weil sie zufällig
// lexikographisch sortieren, und auf diese Zufälligkeit soll sich nichts stützen.
type Datum struct {
	Jahr  int
	Monat int
	Tag   int
}

func (d Datum) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Jahr, d.Monat, d.Tag)
}

// VorOder sagt, ob dieses Datum nicht nach dem anderen liegt.
func (d Datum) VorOder(anderes Datum) bool {
	if d.Jahr != anderes.Jahr {
		return d.Jahr < anderes.Jahr
	}
	if d.Monat != anderes.Monat {
		return d.Monat < anderes.Monat
	}

	return d.Tag <= anderes.Tag
}

// Leer sagt, ob gar kein Datum gesetzt ist.
func (d Datum) Leer() bool { return d == Datum{} }

// DatumLesen liest eine Version der Form JJJJ-MM-TT.
//
// Über `time.Parse` und nicht über drei `strconv.Atoi`: das prüft nebenbei, ob
// es den Tag überhaupt gibt. Ein „2026-02-30" käme sonst durch und fiele erst
// beim Vergleich auf, wo niemand ihn sucht.
func DatumLesen(text string) (Datum, error) {
	gestutzt := strings.TrimSpace(text)

	zeitpunkt, err := time.Parse(time.DateOnly, gestutzt)
	if err != nil {
		return Datum{}, fmt.Errorf("%q ist keine Version der Form JJJJ-MM-TT", gestutzt)
	}

	return Datum{
		Jahr:  zeitpunkt.Year(),
		Monat: int(zeitpunkt.Month()),
		Tag:   zeitpunkt.Day(),
	}, nil
}

// paketname gibt den Namen eines Pakets aus seinem Pfad.
//
// **Über `filepath` und nicht über `path`**, denn hier steht ein Pfad auf der
// Platte: unter Windows trennt der Gegenschrägstrich (`docs/PORTIERUNG.md`).
// Das ist die Gegenrichtung zu [ArtAusName], die auch Paketpfade sieht.
func paketname(pfad string) string { return filepath.Base(pfad) }
