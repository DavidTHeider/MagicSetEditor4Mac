package domain

// **Wo ein Dienst erreicht wird** — Werkstück 3 von A18
// (#1533).
//
// # Warum es das gibt
//
// **Die Adressen der Cloudanbieter waren fest verdrahtet.** In `app` rief
// jede Naht ihren Klienten mit `""`, und der nahm dann seine
// `Vorgabeadresse`. Ein Gateway — LiteLLM, OpenRouter, Azure, ein
// Firmengateway — war damit nicht erreichbar, und ein billiger Testendpunkt
// ebenso wenig.
//
// # Und warum die Vorgabe nicht hier steht
//
// **Weil `domain` die Blätter nicht kennt** (§2). `openai.Vorgabeadresse`
// gehört `internal/openai`; eine zweite Aufzählung hier liefe beim ersten
// Wechsel auseinander — derselbe Fehler, den der Kopf von `einstellung.go`
// zweimal beschreibt.
//
// **Deshalb sagt `domain` nur, wer eine Adresse tragen darf**, und „leer"
// heißt: die Vorgabe des Anbieters. Dieselbe Gestalt wie [KIAdresse].

import "strings"

// DienstAdresse ist der Einstellungsschlüssel für die Adresse eines Dienstes.
//
// **`dienst.<name>.adresse` und nicht `adresse.<name>`**: der Schalter heißt
// schon `dienst.<name>`, und was zu einem Dienst gehört, steht bei ihm.
func DienstAdresse(dienst Dienst) string {
	return DienstEinstellung(dienst) + ".adresse"
}

// AdresseEinstellbar sagt, ob dieser Dienst hinter einem Gateway stehen darf.
//
// **Nicht jeder darf es.** Scryfall, MTGJSON, Scrydex und die übrigen
// Kartendienste liefern **Daten über Karten**; eine umgelenkte Adresse dort
// hieße, Kartendaten aus einer Quelle zu beziehen, die sich als Scryfall
// ausgibt — und die Herkunftsangabe an jedem Treffer (A2, A15) wäre eine
// Behauptung über jemand anderen.
//
// **Bei den KI-Anbietern ist es umgekehrt**: dort ist das Gateway der
// Normalfall der Branche, und wer es verböte, zwänge jeden Schlüssel direkt
// zum Anbieter.
func (d Dienst) AdresseEinstellbar() bool {
	for _, zeile := range dienstzeilen {
		if zeile.Dienst == d {
			return zeile.Adresse
		}
	}

	return false
}

// OrtFuerDienstadresse stuft ein, wo ein Dienst unter dieser Adresse rechnet.
//
// # Warum die leere Adresse hier **Cloud** heißt
//
// **Und bei [KIAdresse] „lokal"** — das ist kein Widerspruch, sondern
// derselbe Satz von zwei Seiten. Dort heißt leer „die Schleifenadresse",
// hier „die Vorgabe des Anbieters", und die liegt bei Anthropic, OpenAI und
// xAI im Netz. Wer beides gleich behandelte, meldete für einen Cloudanbieter
// ohne eingestellte Adresse „gerechnet wird auf diesem Rechner".
//
// **Gefunden am 2026-09-12 in der Sichtprüfung an A16**: dieselbe
// Verwechslung stand schon einmal in `kianbieterliste.go` und ist dort
// benannt.
//
// # Warum ohne den Dienst
//
// **Die erste Fassung nahm ihn entgegen und benutzte ihn nicht** — ein
// Versprechen, das der Rumpf nicht hält (§10). Die Einstufung hängt an der
// **Adresse**: wer eine nennt, wird nach ihr eingestuft, und wer keine nennt,
// bekommt die Vorgabe seines Anbieters, und die liegt im Netz. Welcher
// Anbieter es ist, ändert daran nichts.
func OrtFuerDienstadresse(adresse string) KIOrt {
	if strings.TrimSpace(adresse) == "" {
		return KIOrtCloud
	}

	return KIOrtFuerAdresse(adresse)
}

// ZugangKontoFuer gibt den Ort im Schlüsselbund — **je Adresse**.
//
// # Warum der Schlüssel zur Adresse gehört
//
// **Ein Schlüssel für ein Gateway ist kein Schlüssel für den Anbieter.** Wer
// die Adresse auf ein Firmengateway stellt, dort seinen Schlüssel einträgt
// und die Adresse später leert, schickte denselben Schlüssel an
// `api.openai.com` — eine Weitergabe an jemanden, für den er nie ausgestellt
// wurde. Das Issue sagt es als eigene Zeile.
//
// **Die Vorgabe behält ihren alten Ort.** `ZugangKonto(dienst, angabe)` ohne
// Zusatz — sonst wäre jeder eingetragene Schlüssel nach diesem Umbau
// verschwunden, und niemand wüsste, warum.
func ZugangKontoFuer(dienst Dienst, angabe, adresse string) string {
	gekuerzt := strings.TrimSpace(adresse)
	if gekuerzt == "" {
		return ZugangKonto(dienst, angabe)
	}

	// **Ein Abdruck und nicht die Adresse selbst.** Ein Kontoname trägt keine
	// Schrägstriche, und eine Adresse im Klartext im Schlüsselbund verriete
	// jedem, der die Liste sieht, wohin dieser Rechner spricht.
	//
	// **Getrennt mit einem Punkt und nicht mit `@`.** Der Schlüsselbund
	// erlaubt Buchstaben, Ziffern, Punkt, Strich und Unterstrich — sonst
	// nichts. Die erste Fassung nahm `@`, und `Schreibe` wies den Ort ab:
	// der Gateway-Schlüssel ließ sich **gar nicht ablegen**. Gefunden hat es
	// die Sichtprüfung; mein Lauf prüfte auf Leerzeichen und Schrägstriche —
	// einen Zeichensatz, den ich mir ausgedacht hatte.
	return ZugangKonto(dienst, angabe) + ".an." + kurzabdruck(gekuerzt)
}

// kurzabdruck gibt die ersten zwölf Stellen des Abdrucks einer Adresse.
//
// **Zwölf Stellen reichen**, um zwei Gateways zu unterscheiden, und sie machen
// den Kontonamen lesbar. **Es ist kein Geheimnisschutz** — derselbe Satz wie
// beim Inhaltshash in A15: er sagt, ob zwei Adressen dieselbe sind, und sonst
// nichts.
func kurzabdruck(adresse string) string {
	return abdruckVon([]string{adresse})[:12]
}
