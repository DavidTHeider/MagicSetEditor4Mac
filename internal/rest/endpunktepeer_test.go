// **Die Wege der eigenen Peeridentität** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort stehen die Wege
// **neben** dem Bestand, hier die drei, die sagen, wer dieses Programm ist.
package rest_test

import "net/http"

// endpunktePeer sind die drei Wege zur eigenen Identität.
func endpunktePeer() []endpunkt {
	return []endpunkt{
		{
			// **Die eigene Peeridentität** (N1, #1454) — `GET`, ohne
			// Nebenwirkung.
			name: "Peeridentität lesen", methode: http.MethodGet,
			weg: "/peer", erfolg: http.StatusOK,
		},
		{
			// **Die erste Identität anlegen** — `POST`, und nur, wenn es
			// keine gibt.
			name: "Peeridentität anlegen", methode: http.MethodPost,
			weg: "/peer", erfolg: http.StatusOK,
		},
		{
			// **Und sie ablösen** — `POST`, jedes Mal eine neue.
			name: "Peeridentität rotieren", methode: http.MethodPost,
			weg: "/peer/rotation", erfolg: http.StatusOK,
		},
	}
}
