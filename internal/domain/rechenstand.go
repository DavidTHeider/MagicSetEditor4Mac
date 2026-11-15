package domain

// **Der Stand einer Rechnung** — #2533 (Nachlauf über das Gesamtreview über
// 0.9.0, Befund 4).
//
// # Warum der Kartenstand nicht genügt
//
// **Ein Wertskript liest auch die Setangaben.** Seit #2396 speichert der
// Feldweg, was gerechnet ist, und hält beim Schreiben den Kartenstand, aus dem
// gerechnet wurde. Eine Änderung an `set.*` dazwischen ließ er durch: die
// Karte war dieselbe, gespeichert wurde ein Wert aus Setangaben, die es nicht
// mehr gab.
//
// # Warum zusammengesetzt und nicht eine Summe über beides
//
// **Weil der Aufrufer die Teile braucht.** Ein KI-Auftrag nennt einen reinen
// Kartenstand, und wer ihn gegen den gerechneten hält, vergleicht den
// Kartenteil; der Bestand prüft den Setteil in derselben Transaktion. Eine
// Summe über beides beantwortete nur „gleich oder nicht" — und ein Auftrag
// ohne Setteil wäre nie gleich.

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

// standtrenner steht zwischen Karten- und Setteil — ein Zeichen, das in
// keiner hexadezimalen Prüfsumme vorkommt.
const standtrenner = "+"

// Rechenstand setzt den Stand einer Karte und den ihrer Setangaben zusammen.
func Rechenstand(karte string, setwerte []Value) string {
	return karte + standtrenner + Setstand(setwerte)
}

// StandTeile zerlegt einen Stand in Karten- und Setteil. **Ein reiner
// Kartenstand hat keinen Setteil** — dann ist der zweite leer.
func StandTeile(stand string) (karte, set string) {
	karte, set, _ = strings.Cut(stand, standtrenner)

	return karte, set
}

// Setstand ist die Prüfsumme über die Setangaben.
//
// **Auch ohne Angaben gibt es einen** — sonst fiele eine erste, die zwischen
// Rechnen und Schreiben entsteht, nicht auf. **Die Reihenfolge zählt nicht**,
// aus demselben Grund wie bei [CardRow.Stand].
func Setstand(werte []Value) string {
	teile := make([]string, 0, len(werte))
	for _, wert := range werte {
		teile = append(teile, wert.Field+"\x1f"+wert.Text)
	}

	slices.Sort(teile)

	summe := sha256.Sum256([]byte("set\x1e" + strings.Join(teile, "\x1e")))

	return hex.EncodeToString(summe[:])
}
