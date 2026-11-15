// **Welche Einstellungen es gibt — die von Hand aufgeschriebenen** —
// herausgelöst am 2026-09-21, als `einstellung.go` mit dem Feld `Wirkung`
// (UX-04.AK-02,
// #1676)
// über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// In `einstellung.go` steht, **was eine Einstellung ist** — ihre Gestalt,
// ihre Bereiche, ihre Schlüssel. Hier steht, **welche es gibt**, und in
// `einstellungsreihe.go` daneben, welche nicht von Hand geschrieben sind:
// je Dienst, je Anbieter, je Frage.
//
// **Die Reihenfolge ist die Anzeigereihenfolge** und keine Nebensache: sie
// entscheidet, in welcher Folge die Einstellungsseite und `mse
// einstellungen` sie zeigen.

package domain

// **Von oben nach unten nach Tragweite**: die Sprache betrifft jedes Wort, das
// Erscheinungsbild jede Fläche, die Schriftgröße jede Zeile, der Namenskonflikt
// einen Lauf.
var einstellungen = []einstellung{
	// **Die vier Erscheinungseinstellungen erben** — Werkstück 9 (#1351),
	// Punkt 9. `Erben` steht **vorn** in jeder Werteliste, weil es die
	// Vorgabe ist, und die Vorgabe steht in dieser Anwendung oben (§6). Die
	// Regel dahinter steht in `uebersteuerung.go`; die **Vorgabewerte** sind
	// dorthin gewandert, zum Menschen — hier stünde sonst eine zweite.
	{
		Schluessel: EinstellungSprache,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      append([]string{Erben}, fenstersprachen...),
		Vorgabe:    Erben,
	},

	// **Und die eigene Übersetzung darüber** — LOC-02 (#1688). Ohne
	// Werteliste: welche es gibt, sagt der Bestand.
	{
		Schluessel: EinstellungEigeneSprache,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeEigeneSprache,
	},

	// **Darstellung und Zeitzone gleich hinter der Sprache** — LOC-03.AK-03
	// (#1689). Sie beantworten dieselbe Frage wie sie, nur für Zahlen und
	// Uhren statt für Wörter. **Ohne Werteliste und ohne Vorgabe**: eine
	// Locale und eine Zone lassen sich prüfen, aber nicht aufzählen, und leer
	// heißt „wie das System".
	{
		Schluessel: EinstellungDarstellung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeDarstellung,
	},
	{
		Schluessel: EinstellungZeitzone,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeZeitzone,
	},
	// **Gespeicherte Filter** — UX-12.AK-03 (#1684), Vermerk 0135. Leer
	// heißt: keine.
	{
		Schluessel: EinstellungGespeicherteFilter,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeGespeicherteFilter,
	},
	// **Welche Seite ⌘0 bis ⌘9 öffnen** — UX-02.AK-03 (#1674). Leer heißt:
	// die Vorgabe, die Bereiche der Seitenleiste.
	{
		Schluessel: EinstellungZiffernbelegung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeZiffernbelegung,
	},
	{
		Schluessel: EinstellungErscheinungsbild,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{Erben, "system", "hell", "dunkel"},
		Vorgabe:    Erben,
	},
	{
		Schluessel: EinstellungSchriftgroesse,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{Erben, "klein", "normal", "gross", "sehrgross"},
		Vorgabe:    Erben,
	},

	{
		Schluessel: EinstellungSchriftart,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      append([]string{Erben}, schriftarten()...),
		Vorgabe:    Erben,
	},

	// **Die vier aus UX-09** (#1681).
	//
	// **Ohne `erben`**, anders als die drei Erscheinungseinstellungen
	// darüber: sie gehören der Arbeitsumgebung (`einstellungsebene.go`),
	// und die Rücksetzung des Layouts stellt sie damit **ganz** zurück — ein
	// sicherer Stand, auch wenn nichts mehr zu lesen ist (AK-03).
	{
		Schluessel: EinstellungFarbwelt,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte: []string{
			FarbweltPergament, FarbweltKlar, FarbweltSchiefer, FarbweltMoos,
		},
		Vorgabe: FarbweltPergament,
	},
	{
		Schluessel: EinstellungSchriftbereich,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{SchriftbereichUeberschriften, SchriftbereichGanz},
		Vorgabe:    SchriftbereichUeberschriften,
	},
	{
		Schluessel: EinstellungHintergrund,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte: []string{
			HintergrundAus, HintergrundFuellen, HintergrundEinpassen,
			HintergrundKacheln,
		},
		Vorgabe: HintergrundAus,
	},
	{
		// **Ohne Werteliste und mit eigener Regel** — dieselbe Gestalt wie
		// die Darstellung oben: Farben lassen sich prüfen, nicht aufzählen.
		Schluessel: EinstellungGestaltung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeGestaltung,
	},

	// **Die vier Verhaltensweisen bei einem Namenskonflikt**, in MSEs
	// Reihenfolge:
	//
	//	conflicts->Append(_BUTTON_("keep old"));         // 0
	//	conflicts->Append(_BUTTON_("overwrite"));        // 1
	//	conflicts->Append(_BUTTON_("number"));           // 2
	//	conflicts->Append(_BUTTON_("number overwrite")); // 3
	//
	// `src/gui/images_export_window.cpp`, Zeilen 29–32, Fassung f9d9356 vom
	// 2021-07-06.
	//
	// **Deutsch benannt**, weil es eigene Begriffe dieses Programms sind
	// ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)):
	// MSE nennt sie nicht in einer Datei, die wir lesen, sondern in seiner
	// Oberfläche.
	//
	// **Die Vorgabe ist nachgelesen, nicht gewählt:**
	//
	//	, images_export_conflicts(CONFLICT_NUMBER_OVERWRITE)
	//
	// `src/data/settings.cpp`, Zeile 74, Fassung f9d9356 vom 2021-07-06. Damit
	// ist auch der letzte Abnahmekasten aus #491 eingelöst: die Vorgabe ist
	// **nicht** `ueberschreiben`.
	// **Die Aktualisierung, und sie ist aus.**
	//
	// Die Reihenfolge ist die Steigerung: nichts, melden, holen. Wer sie
	// später umstellt, verschiebt die gespeicherte Wahl derer, die schon
	// gewählt haben.
	{
		// **Erst beim nächsten Start** — gemessen am 2026-09-21:
		// `useAktualisierung` fragt in einem Effekt **ohne
		// Abhängigkeiten**, und der Haken hängt im `Fassungsband` des
		// Rahmens. Wer hier umlegt, sieht die Wirkung, wenn das Fenster das
		// nächste Mal aufgeht.
		Schluessel: EinstellungAktualisierung,
		Wirkung:    WirkungNeustart,
		Bereich:    BereichAnwendung,
		Werte: []string{
			AktualisierungAus, AktualisierungMelden, AktualisierungHolen,
		},
		Vorgabe: AktualisierungAus,
	},

	{
		Schluessel: EinstellungErsteSchritte,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{ErsteSchritteAn, ErsteSchritteAus},
		Vorgabe:    ErsteSchritteAn,
	},
	{
		Schluessel: EinstellungSeitenleiste,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{SpalteSichtbar, SpalteVerborgen},
		Vorgabe:    SpalteSichtbar,
	},
	{
		Schluessel: EinstellungLeistenlage,
		Wirkung:    WirkungSofort,
		Bereich:    BereichAnwendung,
		Werte:      []string{LeisteSenkrecht, LeisteWaagerecht},
		Vorgabe:    LeisteSenkrecht,
	},
	{
		// **Erst beim nächsten Start** — gemessen am 2026-09-21:
		// die Hülle (`internal/fenster`) liest `anwendung.Fensterstand` **einmal** und gibt das
		// Ergebnis an `wails.Run`. Danach fragt niemand mehr.
		Schluessel: EinstellungFensterstart,
		Wirkung:    WirkungNeustart,
		Bereich:    BereichAnwendung,
		Werte: []string{
			FensterstartMaximiert, FensterstartNormal, FensterstartZuletzt,
		},
		Vorgabe: FensterstartMaximiert,
	},

	// **Die zwei Zahlen dahinter tragen keine Werteliste** — der Grund steht
	// bei [EinstellungFensterbreite].
	{
		// **Erst beim nächsten Start**, aus demselben Grund wie
		// [EinstellungFensterstart]: die Hülle gibt die Zahl beim Aufgehen über `SetSize` an
		// `wails.Run`.
		Schluessel: EinstellungFensterbreite,
		Wirkung:    WirkungNeustart,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeFenstermass,
	},
	{
		// **Erst beim nächsten Start**, aus demselben Grund wie
		// [EinstellungFensterstart]: die Hülle gibt die Zahl beim Aufgehen über `SetSize` an
		// `wails.Run`.
		Schluessel: EinstellungFensterhoehe,
		Wirkung:    WirkungNeustart,
		Bereich:    BereichAnwendung,
		Pruefung:   PruefeFenstermass,
	},

	// **Die Vorgaben eines Profils** — Werkstück 9 (#1351), Punkte 2, 3 und 8.
	//
	// **Sie stehen im Bereich `vorgaben`** und nicht bei den
	// Anwendungseinstellungen darüber: jene sagen, wie die Anwendung
	// **aussieht**, diese, womit ein neuer Datensatz **anfängt**.
	{
		Schluessel: VorgabeZustand,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Werte:      zustandswerte(),
		Vorgabe:    OhneVorgabe,
	},
	{
		// **Die Währung, mit der eine neue Zeile anfängt** — LOC-03
		// (#1689). Frei und geprüft: ein Code lässt sich prüfen, und 162
		// davon aufzuzählen wäre keine Wahl.
		Schluessel: VorgabeWaehrung,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Pruefung:   PruefeWaehrungscode,
	},
	{
		Schluessel: VorgabeEinzelspiel,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Werte:      []string{ErsteSchritteAus, ErsteSchritteAn},
		Vorgabe:    ErsteSchritteAus,
	},
	{
		// **Ohne Werteliste** — ein Verzeichnis lässt sich nicht aufzählen.
		// **Aber mit Regel** seit OPS-02 (#1719): absolut oder mit `~/`
		// vorn; ob es den Ordner gibt, fragt `app` die Platte.
		Schluessel: VorgabeHochladepfad,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Pruefung:   PruefeOrdnerangabe,
	},
	{
		Schluessel: VorgabeAusfuhrpfad,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Pruefung:   PruefeOrdnerangabe,
	},
	{
		Schluessel: VorgabeSicherungspfad,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Pruefung:   PruefeOrdnerangabe,
	},
	{
		Schluessel: VorgabeDiagnosepfad,
		Wirkung:    WirkungSofort,
		Bereich:    BereichVorgaben,
		Pruefung:   PruefeOrdnerangabe,
	},

	// **Die lokale KI-Laufzeit** — A7 (#1446). Sie steht im Bereich der
	// Dienste, weil sie dieselbe Frage beantwortet: geht von hier aus etwas
	// an einen Dienst, den dieses Programm nicht ist?
	{
		Schluessel: KILaufzeit,
		Wirkung:    WirkungSofort,
		Bereich:    BereichDienste,
		Werte:      []string{KILaufzeitAus, KILaufzeitAn},
		Vorgabe:    KILaufzeitAus,
	},
	{
		// **Apples Gerätemodell** — AI-05 (#2042). Automatisch heißt:
		// eingebunden, wenn es auf diesem Mac aktiviert ist (Vermerk 0166).
		Schluessel: KIGeraetemodell,
		Wirkung:    WirkungSofort,
		Bereich:    BereichDienste,
		Werte:      []string{KIGeraetemodellAutomatisch, KIGeraetemodellAus},
		Vorgabe:    KIGeraetemodellAutomatisch,
	},
	{
		// **Das Modell des Vorrangs bei Ollama** — #2054 (`kivorrangmodell.go`).
		Schluessel: KIVorrangmodell,
		Wirkung:    WirkungSofort,
		Bereich:    BereichDienste,
		Pruefung:   PruefeKIVorrangmodell,
	},
	{
		// **Ohne Werteliste, mit Regel** — eine Adresse lässt sich nicht
		// aufzählen, aber prüfen: http oder https, und ein Rechner dahinter.
		Schluessel: KIAdresse,
		Wirkung:    WirkungSofort,
		Bereich:    BereichDienste,
		Pruefung:   PruefeKIAdresse,
	},

	// **Die Kontoschau** — T11 (#1476). Sie steht bei den Diensten, weil sie
	// dieselbe Frage beantwortet wie die KI-Laufzeit darüber: geht von hier
	// aus etwas hinaus, und wohin.
	//
	// **Und sie fängt aus an**: wer den Marktplatz eingeschaltet hat, hat
	// damit nicht das eigene Konto freigegeben.
	{
		Schluessel: Kontoschau,
		Wirkung:    WirkungSofort,
		Bereich:    BereichDienste,
		Werte:      []string{KontoschauAus, KontoschauAn},
		Vorgabe:    KontoschauAus,
	},

	// **Die sechs Netzeinstellungen** (N5, #1571). Sie stehen **vor** der
	// Bildausfuhr und hinter den Diensten, weil sie zwischen beiden liegen:
	// sie betreffen jeden Dienst und keine Datei.
	//
	// **Keine Werteliste, keine Vorgabe.** Ein Proxy ist eine Adresse und
	// keine Wahl, und eine Vorgabe hier hieße, die Umgebung zu überschreiben,
	// sobald jemand die Anwendung zum ersten Mal öffnet — derselbe Grund, aus
	// dem die Sprache keine hat.
	{Schluessel: NetzProxyHTTP, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeProxyadresse},
	{Schluessel: NetzProxyHTTPS, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeProxyadresse},

	// **`netz.proxy.aus` wird nicht geprüft**, und das ist eine Entscheidung:
	// `httpproxy` nimmt Rechnernamen, Adressbereiche, `.suffix`, Ports und
	// `*` entgegen, und eine Prüfung, die davon die Hälfte kennte, wiese
	// Gültiges ab. Ein unverständlicher Eintrag trifft dort nichts — er ist
	// wirkungslos und nicht gefährlich.
	{
		Schluessel: NetzProxyAus,
		Wirkung:    WirkungSofort,
		Bereich:    BereichNetz,
	},

	{Schluessel: NetzFristVerbindung, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeNetzfrist},
	{Schluessel: NetzFristKopf, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeNetzfrist},
	{Schluessel: NetzFristStillstand, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeNetzfrist},

	// **Die Schnittstelle der Suche** (NET-02.AK-03, #1706) — leer heißt
	// automatisch. Sie gilt ab der nächsten Suche und Ankündigung.
	{Schluessel: NetzSchnittstelle, Wirkung: WirkungSofort, Bereich: BereichNetz, Pruefung: PruefeSchnittstellenname},
	{Schluessel: NetzLauschen, Wirkung: WirkungNeustart, Bereich: BereichNetz, Pruefung: PruefeLauschwahl},

	// **Das eigene Zeugnis der Dienste** (NET-03.AK-01, #1707) — alle leer
	// heißt: das Zeugnis auf der Peeridentität. Sie gelten beim nächsten
	// Start des Dienstes; ein Tausch der Dateien selbst wirkt am Handschlag.
	{Schluessel: DienstTLSZertifikat, Wirkung: WirkungNeustart, Bereich: BereichNetz, Pruefung: PruefeZeugnispfad},
	{Schluessel: DienstTLSSchluessel, Wirkung: WirkungNeustart, Bereich: BereichNetz, Pruefung: PruefeZeugnispfad},
	{Schluessel: DienstTLSKette, Wirkung: WirkungNeustart, Bereich: BereichNetz, Pruefung: PruefeZeugnispfad},
	{Schluessel: DienstTLSAnker, Wirkung: WirkungNeustart, Bereich: BereichNetz, Pruefung: PruefeZeugnispfad},

	// **Und die Kontingente der Dienste** (QA-08.AK-03, #1739) — neben dem
	// Zeugnis, weil sie dieselben zwei Programme betreffen.
	kontingentschalter(),

	{
		Schluessel: EinstellungBildkonflikt,
		Wirkung:    WirkungSofort,
		Bereich:    BereichBildausfuhr,
		Werte: []string{
			BildkonfliktAltbehalten, BildkonfliktUeberschreiben,
			BildkonfliktNummerieren, BildkonfliktNummerierenNeu,
		},
		Vorgabe: BildkonfliktNummerierenNeu,
	},
}
