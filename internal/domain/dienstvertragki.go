package domain

// **Der Vertrag des ersten KI-Anbieters** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum er in einer eigenen Datei steht
//
// **Je Dienst eine Datei, weil jede ihre eigene Quelle und ihr eigenes
// Abrufdatum hat** (siehe `dienstvertrag.go`). Bei den Kartendiensten ist die
// Quelle eine Dokumentationsseite; hier sind es **drei**, und sie sagen
// Verschiedenes: die Fehlerseite, die Grenzenseite und die Preisseite.
//
// # Und was er nicht sagt
//
// **Was ein einzelnes Modell kostet.** Der Tarif steht je Modell und ändert
// sich; hier steht die **Spanne** mit dem Tag, an dem sie gelesen wurde, und
// der Adresse, unter der die heutige nachzulesen ist. Eine Zahl im Quelltext,
// die niemand nachzählt, veraltet still (§9).

// kivertraege gibt die Verträge der KI-Anbieter.
//
// **Drei seit A10** (#1449), und sie unterscheiden sich in mehr als der
// Adresse: Anthropic **nennt** die Fähigkeiten seiner Modelle, OpenAI nennt
// sie **nicht** ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)),
// und xAI nennt **zwei von dreien**
// ([0101](../../docs/entscheidungen/0101-was-xai-belegt.md)).
func kivertraege() []Dienstvertrag {
	return []Dienstvertrag{
		{
			Dienst: DienstAnthropic, Frage: FrageKIVorschlag,
			Wirkung:      WirkungExternSenden,
			Basisadresse: "https://api.anthropic.com",

			// **Beide Wege der Positivliste** — `anthropic.Wege()`. Keiner
			// davon schreibt beim Anbieter etwas: kein Konto, keine Datei,
			// kein eingereihter Auftrag (die Stapel-API steht in A13).
			Endpunkte: []string{"/v1/models", "/v1/messages"},

			Zugang: "eine Kopfzeile, x-api-key — dazu anthropic-version als " +
				"Pflichtangabe; der Schlüssel liegt im Schlüsselbund von " +
				"macOS und nicht im Bestand",

			Tarif: "kostenpflichtig je Zeichen, und der Preis steht je " +
				"Modell: am 2026-09-12 zwischen 1 und 10 US-Dollar je " +
				"Million Eingabezeichen und zwischen 5 und 50 je Million " +
				"Ausgabezeichen. Ein kostenloses Kontingent gibt es nicht — " +
				"neue Konten bekommen ein kleines Startguthaben",

			Grenze: "je Tarifstufe verschieden; in der untersten 1.000 " +
				"Anfragen je Minute je Modellklasse. Gedrosselt wird auf " +
				"eine je Sekunde, weil die Dokumentation vor einem " +
				"Minutenbudget ausdrücklich warnt. Jede Antwort nennt ihr " +
				"Kontingent in anthropic-ratelimit-*; ein 429 ohne " +
				"retry-after ist der Monatsdeckel und keine Drossel",

			Frist: "keine Zwischenspeicherung — ein Vorschlag ist ein " +
				"Entwurf und kein Datum; der Modellkatalog wird je Abruf neu " +
				"geholt",

			Bedingungen: "https://www.anthropic.com/legal/commercial-terms",
			Abgerufen:   "2026-09-12",
			Beleg:       "https://platform.claude.com/docs/en/api/rate-limits",

			// **Nie gegen sein Original gelaufen** (0056, Entscheidung 4).
			// Ein Lauf brauchte ein Konto mit Guthaben, und jeder Ruf
			// kostete Geld; dieses Programm hat keines. Was daraus folgt,
			// steht in `docs/HAERTUNG.md` und nicht in einer Andeutung.
			Nachweis: nil,
		},
		{
			Dienst: DienstOpenAI, Frage: FrageKIVorschlag,
			Wirkung:      WirkungExternSenden,
			Basisadresse: "https://api.openai.com",

			// **Zwei Wege, und beide lesen.** Es gibt hier kein `/v1/files`
			// (das brauchte A12), keine Bilderzeugung und keinen entfernten
			// Werkzeugserver — das Issue schließt den letzten ausdrücklich
			// aus.
			Endpunkte: []string{"/v1/models", "/v1/responses"},

			Zugang: "eine Kopfzeile, Authorization: Bearer <Schlüssel> — ein " +
				"eigener API-Zugang und **kein** ChatGPT-Abonnement; wer " +
				"ChatGPT bezahlt, hat damit kein API-Guthaben",

			Tarif: "kostenpflichtig je Zeichen, und der Preis steht je " +
				"Modell. Die Denkzeichen eines Denkmodells zählen mit und " +
				"erscheinen in keiner Antwort, die jemand liest",

			Grenze: "je Konto und Tarifstufe verschieden; die Dokumentation " +
				"nennt keine Zahl, sie steht in den Kontoeinstellungen. " +
				"Gedrosselt wird auf eine Anfrage je Sekunde. Jede Antwort " +
				"nennt ihr Kontingent in x-ratelimit-*, und ein 429 heißt " +
				"dreierlei: Drossel, leeres Guthaben oder eine erreichte " +
				"Ausgabengrenze — die letzten beiden gehen nicht vorüber",

			Frist: "keine Zwischenspeicherung — ein Vorschlag ist ein " +
				"Entwurf und kein Datum. Und nichts bleibt beim Anbieter " +
				"liegen: `store` steht ausdrücklich auf `false`, obwohl die " +
				"Vorgabe der Schnittstelle `true` ist",

			Bedingungen: "https://openai.com/policies/services-agreement/",
			Abgerufen:   "2026-09-12",
			Beleg:       "https://developers.openai.com/api/docs/guides/rate-limits",

			// **Nie gegen sein Original gelaufen** — derselbe Grund wie bei
			// Anthropic: ein Lauf brauchte ein Konto mit Guthaben.
			Nachweis: nil,
		},
		{
			Dienst: DienstXAI, Frage: FrageKIVorschlag,
			Wirkung:      WirkungExternSenden,
			Basisadresse: "https://api.x.ai",

			// **Zwei Wege, und beide lesen.** Es gibt hier kein
			// `/v1/responses` — das hielte Antworten **30 Tage** beim
			// Anbieter —, keine Bilderzeugung (A12), keine Stapel-API (A13)
			// und keinen entfernten Werkzeugserver.
			Endpunkte: []string{"/v1/language-models", "/v1/chat/completions"},

			Zugang: "eine Kopfzeile, Authorization: Bearer <Schlüssel> — ein " +
				"eigener API-Zugang und **kein** X-Premium-Abonnement; wer " +
				"Grok im Browser benutzt, hat damit kein API-Guthaben",

			Tarif: "kostenpflichtig je Zeichen, und der Preis steht je " +
				"Modell — die Modellliste selbst nennt ihn mit. **Und die " +
				"Antwort nennt ihren eigenen Preis**: cost_in_usd_ticks, " +
				"ein Tick ist ein Zehnmilliardstel Dollar. Das tut sonst " +
				"keiner der drei",

			Grenze: "die Dokumentation nennt **keine** Zahl und **keine** " +
				"Kopfzeilen über das Kontingent — sie sagt, dass ein 429 " +
				"kommt, wer zu schnell fragt, und rät zu wachsendem Warten. " +
				"Gedrosselt wird auf eine Anfrage je Sekunde, weil eine " +
				"geratene Grenze schlimmer wäre als eine vorsichtige. Ein " +
				"429 heißt hier **eines**: zu schnell gefragt — und das " +
				"geht vorüber",

			Frist: "keine Zwischenspeicherung — ein Vorschlag ist ein " +
				"Entwurf und kein Datum. Und nichts bleibt beim Anbieter " +
				"liegen: dieser Weg legt nichts ab",

			Bedingungen: "https://x.ai/legal/terms-of-service",
			Abgerufen:   "2026-09-12",
			Beleg:       "https://docs.x.ai/developers/debugging",

			// **Nie gegen sein Original gelaufen** — derselbe Grund wie bei
			// den beiden darüber: ein Lauf brauchte ein Konto mit Guthaben.
			Nachweis: nil,
		},
	}
}
