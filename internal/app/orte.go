package app

// **Wo die Anwendung ihre Sachen ablegt** — und wie sie es erfährt.
//
// Eine eigene Datei, weil `app.go` mit ihr über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, wie die Anwendung
// zusammengesetzt wird, hier, wo sie hinschreibt.
//
// # Zwei Hälften, und sie gehören zusammen
//
// **Oben stehen die echten Orte** — `Datenbankpfad`, `Bilderverzeichnis`. Sie
// fragen diesen Rechner.
//
// **Unten stehen die Nähte dorthin** ([ortsnaehte], seit dem 2026-09-16).
// Sie sind der Grund, warum ein Lauf die echten **nicht** fragen muss: ein
// Test, der gegen `os.UserHomeDir` oder gegen den Ort dieses `.app` liefe,
// wäre auf einer zweiten Maschine grün oder rot, je nachdem, wer dort was
// bezogen hat — genau die Sorte Ergebnis, vor der
// [0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md) warnt.
//
// **Der Satz stand siebenmal in `anwendung.go`**, einmal an jedem Feld. Jetzt
// steht er hier.

import (
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
)

// Datenbankpfad gibt den Ort der Datenbank im Benutzerverzeichnis.
//
// Unter `~/Library/Application Support/`, weil dort hingehört, was die Anwendung
// selbst führt — nicht in `~/Documents`, wo Nutzende ihre eigenen Sets ablegen.
func Datenbankpfad() (string, error) {
	verzeichnis, err := konfigurationsort()
	if err != nil {
		return "", fmt.Errorf("Benutzerverzeichnis ermitteln: %w", err)
	}

	return filepath.Join(verzeichnis, "MagicSetEditor4Mac", "bestand.db"), nil
}

// Bilderverzeichnis gibt den Ort, an dem eingesetzte Bilder liegen.
//
// Neben der Datenbank und den Paketen unter `~/Library/Application Support/`:
// es ist Bestand der Anwendung, kein Dokument der Nutzenden
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
//
// **Ohne `MkdirAll`.** Das Verzeichnis entsteht beim ersten Bild
// ([service.Bildablage.Ablegen]); es bei jedem Start anzulegen hinterließe auf
// jedem Rechner einen leeren Ordner, den die allermeisten nie füllen.
//
// **Neben dem Bestand, der wirklich offen ist** — QA-10.NAT-01 (#1741). Hier
// stand der Vorgabeort: zwei Instanzen mit eigenem Bestand teilten sich die
// Bilder, und die Bilder eines Bestands lagen neben einem anderen.
func Bilderverzeichnis() (string, error) {
	pfad, err := BestandsortFuerFlaechen()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(pfad), "beiwerk"), nil
}

// ortsnaehte ist in [App] eingebettet — die Nähte zu den Orten oben.
//
// # Der Anlass für diesen Schnitt
//
// **`anwendung.go` stand nach G5 exakt auf 400 Zeilen** — und zwar, weil zwei
// erklärende Kommentarzeilen gestrichen worden waren, um darunter zu bleiben.
// Das ist die falsche Reihenfolge: die Grenze gibt es, damit eine Datei
// **eine** Frage beantwortet, nicht damit Erklärungen verschwinden (§2).
//
// # Und warum `oeffnen` nicht dazugehört
//
// **Weil es keinen Ort nennt, sondern eine Datenbank aufmacht.** Es steht
// weiter in `anwendung.go` neben `db`; was mit der **Datei darunter**
// geschieht, steht in `anwendungbestand.go`.
type ortsnaehte struct {
	// paketort ist die Naht zum Paketverzeichnis.
	paketort func() (string, error)

	// bilderort ist die Naht zu [Bilderverzeichnis].
	//
	// **Als Feld, seit sein Fehler protokolliert wird** (#857): ein Test kann
	// kein Benutzerverzeichnis wegnehmen. Ohne die Naht wäre der Zweig, der
	// ihn meldet, einer, den niemand betritt — und damit eine Behauptung
	// darüber, was im Ernstfall geschieht (§3).
	bilderort func() (string, error)

	// saatort ist die Naht zur mitgelieferten Saat — ein Test hat kein
	// `.app`, und der Ort hängt davon ab, wo dieses Programm liegt.
	saatort func() string

	// vorlagenort und beispielsetort sind die Nähte zu dem, was der
	// Freigabeweg ins Bündel legt
	// (#702).
	vorlagenort    func() string
	beispielsetort func() string

	// schnittstellen sind die Orte, an denen OPS-09 einrichtet — Zuhause,
	// Werkzeugort, PATH der Anmeldeshell ([App.MitSchnittstellenorten]).
	schnittstellen schnittstellenfelder

	// mse2orte ist die Naht zu den Vorschlägen aus #55.
	mse2orte func() []string

	// downloadort ist die Naht zum Downloadverzeichnis — aus demselben Grund
	// als Feld wie `bilderort`: er entsteht aus dem Zuhause des Rechners.
	downloadort func() (string, error)

	// netzliste ist die Naht zu den Schnittstellen dieses Rechners (G08,
	// #2272) — welche es gibt, hängt an WLAN, VPN und einer VM (C06).
	netzliste func() ([]entdeckung.Schnittstelleneintrag, error)
}
