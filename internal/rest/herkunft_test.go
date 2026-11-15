package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Wovor die Herkunftsprüfung schützt**
// (#997).
//
// Nicht vor einem fremden Rechner — dagegen steht die Bindung an
// `127.0.0.1`. Sondern vor einem **Browser auf diesem** Rechner, und zwar vor
// zwei Wegen:
//
//   - **Eine Seite, die einfach hinschreibt.** `fetch("http://127.0.0.1:…")`
//     aus einer beliebigen Seite. Sie kennt die Marke nicht, aber sie soll
//     nicht einmal bis zur Markenprüfung kommen — ein Browser schickt bei
//     einer solchen Anfrage `Origin` mit, ein Skript nicht.
//   - **DNS-Rebinding.** Der Angreifer lässt seinen Namen auf `127.0.0.1`
//     zeigen; der Browser hält die Seite dann für gleichherkünftig und schickt
//     **kein** `Origin`. Was er aber schickt, ist `Host: angreifer.example`.
//
// Beide Wege haben eine Kopfzeile, die ein Skript nicht hat oder anders setzt.
// Genau daran werden sie erkannt.
func TestEineFremdeHerkunftWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*http.Request){
		"eine Seite schickt Origin": func(r *http.Request) {
			r.Header.Set("Origin", "https://angreifer.example")
		},
		"eine Seite von localhost schickt auch Origin": func(r *http.Request) {
			r.Header.Set("Origin", "http://localhost:3000")
		},
		"ein leeres Origin ist auch eines": func(r *http.Request) {
			r.Header.Set("Origin", "null")
		},
		"Rebinding: fremder Host": func(r *http.Request) {
			r.Host = "angreifer.example:8080"
		},
		"Rebinding: Host ohne Port": func(r *http.Request) {
			r.Host = "angreifer.example"
		},
		"ein Verweis von einer Seite": func(r *http.Request) {
			r.Header.Set("Referer", "https://angreifer.example/seite")
		},
	}

	for name, setzen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
			anfrage.Host = "127.0.0.1:7777"
			setzen(anfrage)

			if rest.HerkunftTaugt(anfrage) {
				t.Error("die Anfrage kam durch")
			}
		})
	}
}

// **Und ein Skript kommt durch.**
//
// Ohne diesen Fall stünde die Prüfung auch dann grün da, wenn sie alles
// abwiese (§10). Die drei Gestalten sind die, die wirklich vorkommen: `curl`
// setzt `Host` aus der Adresse, und niemand setzt `Origin`.
func TestEinSkriptKommtDurch(t *testing.T) {
	t.Parallel()

	for _, rechner := range []string{"127.0.0.1:7777", "localhost:7777", "127.0.0.1"} {
		t.Run(rechner, func(t *testing.T) {
			t.Parallel()

			anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody)
			anfrage.Host = rechner

			if !rest.HerkunftTaugt(anfrage) {
				t.Errorf("eine Anfrage an %s wurde abgewiesen", rechner)
			}
		})
	}
}
