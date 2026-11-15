package werkzeuge

// **Die Rückfrage „Lokales Netzwerk" sagt, wofür** —
// #1996.
//
// Gemessen am 2026-10-01 in der Abnahme-VM (macOS 26.4): *Become reachable*
// löste die Rückfrage aus, und sie trug nur den Systemsatz — das Bündel
// nannte keinen Zweck. Apple verlangt ihn: „If your app accesses the local
// network, add the NSLocalNetworkUsageDescription property to its
// Info.plist" (TN3179, nachgelesen am 2026-10-01).
//
// # Was dieser Wächter nicht prüft
//
// - **Ob macOS den Satz zeigt** — das sieht nur, wer die Rückfrage auslöst
//   (`docs/HANDPRUEFUNGEN.md`).
// - **`NSBonjourServices`**: TN3179 verlangt es für Bonjour-APIs (dnssd,
//   Network framework, Multipeer Connectivity). `internal/entdeckung` spricht
//   mDNS über eigene Sockets (hashicorp/mdns) und erreicht keine davon.

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

const zweckschluessel = "NSLocalNetworkUsageDescription"

// zweckInPlist liest den Zweckstext aus der Vorlage des Bündels.
var zweckInPlist = regexp.MustCompile(`<key>` + zweckschluessel + `</key>\s*<string>([^<]+)</string>`)

// zweckInStrings liest ihn aus einer `InfoPlist.strings`.
var zweckInStrings = regexp.MustCompile(`"` + zweckschluessel + `"\s*=\s*"([^"]+)";`)

func TestDasBuendelNenntDenZweckDesLokalenNetzes(t *testing.T) {
	plist := zweckInPlist.FindSubmatch(projektdatei(t, "build/darwin/Info.plist"))
	if len(plist) < 2 || strings.TrimSpace(string(plist[1])) == "" {
		t.Fatalf("Info.plist trägt keinen %s", zweckschluessel)
	}

	// **Englisch eigens**: die Entwicklungsregion ist Deutsch, und ein
	// englisches System sähe sonst den deutschen Satz.
	englisch := zweckInStrings.FindSubmatch(projektdatei(t, "build/darwin/en.lproj/InfoPlist.strings"))
	if len(englisch) < 2 || bytes.Equal(englisch[1], plist[1]) {
		t.Fatalf("en.lproj trägt keinen eigenen %s", zweckschluessel)
	}

	if !strings.Contains(string(projektdatei(t, "scripts/bauen.sh")), "build/darwin/en.lproj") {
		t.Error("scripts/bauen.sh legt en.lproj nicht ins Bündel")
	}
}
