package domain

// **Was dieses Programm heute wirklich durchsetzt** — Werkstück G4 von Phase
// 5.5 (#1461).
//
// # Der erste automatische Teilumfang, konkret festgeschrieben
//
// **Das Issue verlangt genau das**: „Ersten automatischen Teilumfang konkret
// festschreiben." Hier steht er — und er ist klein, weil er ehrlich ist.
//
// **Drei von elf Bereichen waren es bei G4**, und zwar die drei, die der
// Spielkern aus G1 wirklich führte — **heute sind es zehn von dreizehn**, und
// welche das sind, steht bei [Grundprofilfassung]:
//
//   - **Zonen** (CR 400): ein Objekt liegt in genau einer Zone, jede Bewegung
//     ist ein Ereignis, und die geordneten Zonen haben eine Reihenfolge, die
//     niemand umgeht ([Zone.Geordnet]).
//   - **Marken** (CR 122): der Kern führt die Zähler eines Objekts und
//     verändert sie nur über einen Befehl.
//   - **Anhängsel** (Ausrüstung CR 301, Auren CR 303): der Kern weiß, was an
//     was hängt, und löst es beim Wegräumen.
//
// **Die anderen acht sind `manuell`** — nicht `unbekannt`: es ist eine
// Entscheidung und kein Versäumnis. Zugstruktur, Stapel, Priorität, Combat,
// zustandsbasierte Aktionen, Spielsteine, Typen und Wechselwirkungen macht,
// wer am Tisch sitzt.
//
// # Was „automatisch" hier heißt und was nicht
//
// **Buchführung, nicht Regelkenntnis.** Der Kern garantiert, dass die
// Zonenzugehörigkeit stimmt — nicht, dass eine Karte dorthin **durfte**. Er
// zählt Marken — er weiß nicht, was eine Marke bewirkt. Genau deshalb trägt
// jeder Eintrag seine Grenzen, und genau deshalb steht [StrengeStreng] nicht
// zur Wahl: acht Bereiche fehlen, und das Issue lässt keine Zwischenstufe zu.
//
// # Und deshalb ist „begleitet" verfügbar und „streng" nicht
//
// **Das ist die Aussage dieses Profils**, und sie ist nachprüfbar statt
// behauptet: [Regelmatrix.TraegtStrenge] rechnet sie aus, und
// [Regelmatrix.Luecken] nennt die acht.

// Grundprofilfassung ist die Fassung, die [Grundprofil] gibt.
//
// **Sie steigt, wenn sich an der Matrix etwas ändert** — und eine laufende
// Partie bleibt bei der, unter der sie angefangen hat.
//
// **Zwei seit dem 2026-09-16**: jeder der drei automatischen Bereiche hat eine
// Grenze dazubekommen, und eine geänderte Zusage ist eine andere Zusage. Was
// sich geändert hat, steht bei [grundmatrix].
//
// **Drei seit dem 2026-09-17**: G9
// (#1592)
// hat `zugstruktur` und `prioritaet` auf `automatisch` gestellt.
//
// **Vier am selben Tag**: `M2`
// (#1604)
// hat `typen` von `manuell` auf `begleitet` gehoben — das Programm liest die
// Typzeile und **warnt**, wo es ein Wort nicht kennt.
//
// **Sechs am selben Tag**: `M5`
// (#1607)
// hat `zustandsbasiert` von `manuell` auf `automatisch` gehoben — dreizehn
// der fünfundzwanzig Aktionen aus CR 704.5 laufen von selbst.
//
// **Fünf am selben Tag**: `M1`
// (#1603)
// hat zwei **Grenzen** von `prioritaet` aufgehoben — der Stapel verrechnet,
// und Abkürzungen gibt es. Eine Grenze, die nicht mehr gilt, ist eine
// falsche Zusage; gesehen bei der ersten Sichtprüfung mit einem echten Deck,
// wo die Regelfreigabe „kein Stapel, also keine Verrechnung" zeigte, während
// daneben ein Stapel verrechnete.
//
// **Sieben am selben Tag**: `M3`
// (#1605)
// bringt den **zwölften Bereich** — `mana` — und stellt ihn auf
// `automatisch`: der Vorrat entsteht, leert sich am Abschnittsende (CR 500.5)
// und ein Grundland tappt für sein Mana (CR 305.6).
//
// **Acht am selben Tag**: `M4`
// (#1606)
// hebt `stapel` von `manuell` auf `automatisch` — ein Zauber wird angekündigt,
// bezahlt und obenauf gelegt (CR 601.2), ein Land kommt ohne Stapel auf den
// Tisch (CR 116.2a), und beim Verrechnen werden die Ziele noch einmal geprüft
// (CR 608.2b).
//
// **Neun am selben Tag**: `M6`
// (#1608)
// hebt `combat` von `manuell` auf `automatisch` — Angreifer und Blocker
// werden erklärt, der Kampfschaden geschieht von selbst, und wer Erstschlag
// hat, teilt in einem eigenen Schritt aus (CR 510.4).
//
// **Zehn am selben Tag**: `M7`
// (#1609)
// hebt `spielsteine` von `manuell` auf `begleitet` — acht der
// neunundsechzig Handlungen mit Schlüsselwort (CR 701) gehen, **wenn jemand
// sie stellt**; von selbst geschieht keine.
//
// **Elf am selben Tag**: `M12`
// (#1614)
// füllt `spielsteine` mit dem, wonach der Bereich heißt — ein Spielstein
// entsteht (CR 111.1), ein Emblem entsteht (CR 114.1), die Kontrolle
// wechselt (CR 110.2 mit CR 613.1b), ein verdecktes Permanent ist eine 2/2
// (CR 708.2a), und eine zweiseitige Karte wandelt sich um (CR 701.27a).
//
// **Und es nimmt zwei Grenzen weg.** „Kein Spielstein wird erzeugt" gilt
// nicht mehr, und bei `zustandsbasiert` fällt die **Legendenregel**
// (CR 704.5j) heraus: `M5` führte sie als Grenze, weil es am Tisch keinen
// Weg gab, jemanden zu fragen. Den Weg gibt es jetzt — eine offene Wahl in
// der Partie und ein Befehl, der sie beantwortet.
//
// **Zwölf am selben Tag**: `M18`
// (#1619)
// nimmt **drei** Grenzen weg — die härteste bei `zonen`, ihre Schwester bei
// `marken` und die dritte bei `anhaengsel`. Ein Zonenwechsel macht seitdem
// ein **neues** Objekt (CR 400.7), an jeder Stelle, an der etwas die Zone
// wechselt, und auch beim Schieben von Hand.
//
// **Dreizehn am selben Tag**: `M13`
// (#1615)
// setzt CR 800.4 durch — eine Partie zu dritt läuft weiter, wenn jemand
// geht, und die Zugreihenfolge wie die Priorität überspringen ihn. Der
// Bereich `prioritaet` bekommt CR 800 als Quelle und eine Grenze über die
// wahlweisen Mehrspielerregeln.
//
// **Vierzehn am selben Tag**: `M20`
// (#1621)
// nimmt der Zugstruktur die Grenze „keine Tag/Nacht-Prüfung": sie läuft im
// Enttappschritt (CR 502.2, CR 731.2a/b), nach der Zahl der Zauber, die der
// vorige aktive Spieler gewirkt hat. **Monarch, Initiative und
// Radioaktivitätsmarken** gehören zu keinem der dreizehn Bereiche — sie
// werden **gestellt** und geschehen nicht, denn was sie vergibt, sind
// eingebaute ausgelöste Fähigkeiten ohne Quelle (CR 725.2, CR 726.2,
// CR 728.1), und die gibt es erst mit `M9`.
//
// **Fünfzehn seit dem 2026-09-18**: `M9`
// (#1611)
// bringt sie. Der Bereich `wechselwirkungen` steht damit auf `automatisch`
// — ausgelöste Fähigkeiten warten (CR 603.2), kommen in APNAP-Reihenfolge
// auf den Stapel (CR 603.3b mit CR 101.4), werden verrechnet und hören auf
// zu existieren (CR 608.2n). **Der Monarch zieht von selbst**, **wer ihm
// Kampfschaden zufügt, wird Monarch**, **wer dem Träger der Initiative
// Kampfschaden zufügt, nimmt sie**, und **Radioaktivitätsmarken mahlen**.
//
// **Und es nimmt zwei Grenzen weg und schreibt zwei um.** „Kein Auslöser,
// wenn etwas gewirkt wird" nennt jetzt den Grund (der Kartentext fehlt, nicht
// der Mechanismus), und die Ausnahme im Aufräumschritt scheitert nicht mehr
// daran, dass nichts wartet, sondern daran, dass der zweite Aufräumschritt
// fehlt (CR 514.3a).
//
// **Sechzehn seit dem 2026-09-18**: `M14`
// (#1616)
// liest **zwölf** Schlüsselwortfähigkeiten statt drei — Flugfähigkeit,
// Reichweite, Bedrohlich, Verteidiger, Wachsamkeit, Trampelschaden,
// Lebensverknüpfung, Todesberührung und Unzerstörbarkeit sind
// dazugekommen. **Welche zwölf, ist gezählt** und nicht gefühlt
// (`docs/SCHLUESSELWOERTER.md`).
//
// **Und die Lesart hat sich geändert**: gelesen wird die Zeile (CR 702.1)
// und nicht mehr das Wort. `CR 704.5h` fällt damit als Grenze weg —
// fünfzehn der fünfundzwanzig zustandsbasierten Aktionen laufen.
//
// **Siebzehn am selben Tag**: `M8`
// (#1610)
// führt aus, was auf einer Karte steht — CR 608.2c, die längste offene
// Zusage dieses Programms. **Eine Karte mit einem `game script` wirkt im
// Spiel**: das Skript läuft in einem Sandkasten mit dreizehn benannten Wegen
// ([Spielwirkung]), ohne Datei, ohne Netz und mit einer Frist.
//
// **Die Grenze bleibt und ändert ihre Gestalt**: kein bezogenes Paket
// bringt ein Spielskript mit — MSE kennt das Feld nicht, weil MSE2 keinen
// Spielkern hat. Wer eine eigene Karte baut, schreibt es hinein.
//
// **Achtzehn am selben Tag**: `M10`
// (#1612)
// bringt die **Schichten** — zwei der zehn wirken: die sechste
// (Fähigkeiten, CR 613.1f) und die siebte (Werte, CR 613.4b/c), und
// **innerhalb** einer Schicht entscheidet der Zeitstempel (CR 613.7). Damit
// fällt die halbe Grenze von `combat` („ein Effekt, der ein Schlüsselwort
// verleiht, wirkt nicht") und die **zweite Hälfte von CR 514.2**, die `G9`
// ausdrücklich offen gelassen hat: im Aufräumschritt endet jetzt auch, was
// „bis zum Ende des Zuges" gilt.
//
// **Neunzehn am selben Tag**: `M11`
// (#1613)
// bringt die **Ersetzung** — CR 614.1a und CR 615.1. Zwei Ereignisse lassen
// sich ändern: der Weg in den Friedhof (ersetzt durch das Exil) und der
// Schaden (verhindert, ganz oder zum Teil nach CR 615.10). **Ein ersetztes
// Ereignis geschieht nie** (CR 614.6), und ein Effekt bekommt **eine**
// Gelegenheit (CR 614.5).
//
// **Damit sind alle neunzehn Werkstücke der Regelmaschine gebaut.**
//
// **Zwanzig am 2026-09-18** — der Nachlauf zum Zweitblick über Phase 5.5
// (§10) hat drei Grenzen angefasst:
//
//   - **CR 613.8 nennt jetzt den richtigen Grund.** Eine Abhängigkeit
//     entsteht nur **innerhalb** einer Schicht (CR 613.8a); der alte Satz
//     begründete ihr Fehlen zwischen zwei Schichten, und dort kann es gar
//     keine geben.
//   - **„neun Wirkungen" waren dreizehn.** `M10` und `M11` hatten vier
//     dazugelegt, und diese Stelle war nicht mitgezogen worden. Sie steht
//     seitdem unter einem Wächter.
//   - **Die Auswertungszeit eines Spielskripts hat ein Issue**
//     (#1625),
//     und die Grenze nennt es.
//
// **Einundzwanzig am 2026-09-19** — die Grenze über die Auswertungszeit
// eines Spielskripts ist **weg**
// (#1625):
// gelesen wird jetzt beim Verrechnen, wie CR 608.2c es verlangt, und eine
// Grenze, die nicht mehr gilt, ist eine falsche Zusage.
//
// **Zweiundzwanzig am selben Tag** — `commander` ist im Grundprofil **keine
// Lücke mehr**
// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
// #1626):
// er gehört einer Variante, die eine Sandkastenpartie nicht spielt, und
// [Regelmatrix.Luecken] zählt ihn dort nicht mehr mit. **Damit trägt das
// Grundprofil immer noch kein `streng`** — aber aus einem Grund, den jemand
// schließen kann.
//
// **Ein geänderter Grund ist eine andere Zusage**, auch wo die Grenze
// selbst dieselbe bleibt: eine Partie friert die **Nummer** ein und nicht
// den Text ([Regelwerk.Fassung]), und zwei Matrizen unter derselben Nummer
// wären nicht mehr auseinanderzuhalten.
const Grundprofilfassung = 22

// Grundprofilquelle ist die Fassung der Comprehensive Rules, gegen die die
// Matrix geschrieben wurde.
//
// **Abgerufen am 2026-09-15** von `magic.wizards.com/en/rules`; die
// Abschnittsnummern stammen aus dem Inhaltsverzeichnis von
// `MagicCompRules 20260819.txt` und nicht aus dem Gedächtnis (§11).
const Grundprofilquelle = "2026-08-19"

// Grundprofil gibt das Regelprofil, mit dem eine Partie anfängt.
//
// **Eine Funktion und keine Variable**: ein Profil trägt Listen, und eine
// gemeinsame Variable ließe sich von jedem Aufrufer verbiegen.
func Grundprofil(spiel string) Regelprofil {
	return Regelprofil{
		Kennung: "sandkasten",
		Fassung: Grundprofilfassung,
		Spiel:   spiel,
		Wirksam: Grundprofilquelle,
		Quelle: Regelquelle{
			Dokument: DokumentCR,
			Fassung:  Grundprofilquelle,
			Nummern:  []string{"100"},
		},

		// **Zwei bis vier** — das Issue nennt die Spanne für G5 („2–4
		// Spieler"), und ein Tisch für einen allein ist Goldfish (G3).
		//
		// **Ab drei ist es eine Mehrspielerpartie** — CR 800.1, und seit
		// `M13` (#1615)
		// trägt sie: wer geht, wird in Zugreihenfolge und Priorität
		// übersprungen (CR 800.4), und sein Kampfschaden fällt aus
		// (CR 800.4e).
		Spieler: Spielerzahl{Mindestens: 1, Hoechstens: 4},

		// **Zwanzig** — CR 103.4 (Fassung 2026-08-19): jeder Spieler fängt
		// mit zwanzig Lebenspunkten an. Commander weicht ab
		// (CR 103.4c: vierzig), und das ist ein eigenes Profil.
		//
		// **Nachgelesen und nicht erinnert** (§11): der erste Anlauf schrieb
		// hier `103.3` und `903.7` — beide erfunden, beide plausibel, beide
		// falsch. Eine erfundene Nummer ist schlimmer als keine, weil sie
		// wie ein Beleg aussieht.
		Startleben: 20,

		Matrix: grundmatrix(),

		// **Alle sechs sind erlaubt** — im Sandkasten wird nichts
		// durchgesetzt, also gibt es nichts, was eine Abweichung stören
		// könnte. Ein engeres Profil erlaubt weniger.
		ErlaubteHausregeln: Hausregelarten(),
	}
}
