# Hüllenprobe: i05geheimnisse

**bestanden** — Commit `7e60af93b`,
2026-10-10 20:02 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-10T20-01-58Z-43818`, Bauabdruck `ebe63794974c`.


**1 — die Marke in den Bund der Probe**

- Scrydex-Schlüssel: gut
- Scrydex-Team: gut
- CardTrader-Marke: gut

**2 — kontrollierte Fehler über den Proxy, und Erfolg**

- Scryfall-Regeln: netz — der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.scryfall.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung) — und es gibt keinen Stand im Speicher (und d [Vorgang 1569f89f]
- Scrydex-Verlauf: netz — Scrydex fragen: der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.scrydex.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung) [Vorgang 0577b3d3]
- CardTrader-Angebote: netz — der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.cardtrader.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung) [Vorgang 4481034a]

**3 — das Diagnosepaket**

- im Paket: uebersicht.json, fassung.txt, bestand.txt, pakete.txt, metriken.json, protokoll.txt, diagnose.json
- ✓ jede Prüfsumme des Manifests gilt, und nichts liegt daneben — diagnose.json
- ✓ im Paket steht die Marke nirgends — weder im Inhalt noch in einem Namen — kein Fund

**4 — der Unified Log seit dem Start**

- ✓ ⌘Q beendet
- Grund im Protokoll, Scryfall-Regeln: time=2026-10-10T22:02:02.876+02:00 level=DEBUG msg="Ruf des Fensters" stelle=app.(*App).ScryfallRegeln kennung=1569f89f ⏎ time=2026-10-10T22:02:02.878+02:00 level=ERROR msg="der verdeckte Grund einer Meldung" vorgang="der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.scryfall.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung) — und es gibt keinen Stand im Speicher (und die Sammeldatei half nicht: nicht gefunden: es liegt keine Sammeldatei bereit (»mse scryfall-sammeldaten 
- Grund im Protokoll, Scrydex-Verlauf: time=2026-10-10T22:02:02.898+02:00 level=DEBUG msg="Ruf des Fensters" stelle=app.(*App).Fremdpreisverlauf kennung=0577b3d3 ⏎ time=2026-10-10T22:02:02.937+02:00 level=ERROR msg="der verdeckte Grund einer Meldung" vorgang="Scrydex fragen: der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.scrydex.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung)" grund="Get \"https://api.scrydex.com/magicthegathering/v1/cards\": proxyconnect tcp: dial tcp 127.0.0.1:9: connect: connection refuse
- Grund im Protokoll, CardTrader-Angebote: time=2026-10-10T22:02:02.953+02:00 level=DEBUG msg="Ruf des Fensters" stelle=app.(*App).Marktangebote kennung=4481034a ⏎ time=2026-10-10T22:02:02.974+02:00 level=ERROR msg="der verdeckte Grund einer Meldung" vorgang="der fremde Dienst ist nicht erreichbar: an seiner Adresse nimmt niemand an: api.cardtrader.com (über den Proxy http://nutzer:xxxxx@127.0.0.1:9 aus umgebung)" grund="Get \"https://api.cardtrader.com/api/v2/expansions\": proxyconnect tcp: dial tcp 127.0.0.1:9: connect: connection refused" kennung=4481034a
- Unified Log: 2067 Einträge der Anwendung, 0 des Systems mit einer Form der Marke
- ✓ die Anwendung hat ihre Einträge geschrieben — der Leser greift nicht daneben — 2067
- ✓ im Protokoll der Anwendung steht die Marke nirgends — kein Fund
- beim System: kein Fund — kein Leck dieses Programms, wenn es dort steht, aber eine Auskunft
- ✓ in der Ausgabe und im Protokoll auf der Platte steht sie nirgends — kein Fund

**5 — die Gegenprobe: ein ungeschwärztes Kontrollstück**

- ✓ der Scanner findet das Kontrollstück, roh und als Basic-Auth — kopf.txt:base64, roh.txt:roh
- ✓ und ein Eintrag der Anwendung mit der Marke zählt als Fund der Anwendung — {"app":[{"marke":"MSEGEHEIM55A3CC024EC0","art":"roh","stelle":14,"prozess":"/Applications/X.app/Contents/MacOS/MagicSetEditor4Mac"}],"system":[]}
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
