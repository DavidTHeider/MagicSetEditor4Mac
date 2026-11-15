// **Nutzer, Profile, Freigaben, Token, Zugänge** — eine Gruppe des
// Vertragsrundgangs (B01, #2220).
//
// Sie legt **zwei eigene Dinge** an und räumt sie am Ende wieder weg: einen
// zweiten Menschen (`Vertragsnutzer B`) und ein zweites Profil des arbeitenden
// Menschen (`Vertragsprofil B`). Alles andere hängt daran — Bild, Abzeichen,
// Einstellung, Freigabe, Vorgabe, Token. **Der arbeitende Mensch selbst wird
// nur gelesen und nie verändert**; was ihn berührt (Kennwort, Sitzung,
// Wechsel), steht in `b-sitzung.mjs`.
//
// # Warum zwei Dinge und nicht eines
//
// Ein Mensch schreibt über sich selbst ohne Rückfrage; über einen **anderen**
// nur, wenn dessen zweiter Faktor nicht fragt (`internal/service/ueberfremde.go`).
// Der zweite Mensch hat keinen Faktor — damit gilt für ihn dasselbe wie für
// einen Menschen ohne Schutz — und ist das Ziel von Änderungen, Bild,
// Einstellung und Löschung. Das zweite Profil gehört dem **arbeitenden**
// Menschen und ist damit ein Profil, an dem er Verwalter ist: Token, Vorgaben
// und Freigaben verlangen diese Rolle (`internal/app/zielprofil.go`).
//
// **Nicht gelingbar in einer Einzelanwendung**: `KontoLoesen` — es nimmt die
// Kennung eines Anbieters (GitHub, Google …) von einem Menschen, und eine
// solche gibt es erst nach einer Anmeldung im Browser (`Anmelden`, in
// `scripts/bindungsreichweite.ausnahmen.json`).
//
// # Wie weit das geprüft ist
//
// Die Abfolge ist gegen den App-Bau gefahren worden (`app.NeueBindungen` über
// einem frisch gestarteten `app.NeueApp`, Wegwerfprobe vom 2026-10-09, nicht
// eingecheckt), die Argumentverbünde strikt gegen die Go-Typen und die
// `nimm`-Ausdrücke gegen die serialisierten Antworten. **Nicht geprüft** ist,
// was nur das native Fenster hat: die Wails-Naht selbst, die echten
// Dateidialoge (dort stand ein eingesetzter Dateiwähler) und jeder
// Unterprozess.

import { join } from 'node:path'

import { b, schreibePng } from './hilfen-b.mjs'

// n2 und p2 sind die beiden eigenen Dinge dieser Gruppe.
const n2 = (k) => b(k).n2
const p2 = (k) => b(k).p2
const t1 = (k) => b(k).t1

// nutzerform ist, was die Folgeeinträge vom neuen Menschen brauchen.
const nutzerform = '({ id: a.nutzer.id, name: a.nutzer.name, fassung: a.nutzer.fassung })'
const profilform = '({ id: a.profil.id, name: a.profil.name, fassung: a.profil.fassung })'
const tokenform = '({ id: a.token.id, fassung: a.token.fassung })'

// uebersicht nennt die gespeicherten Fassungen aller Menschen und Profile.
const uebersicht = 'a.nutzer.map((z) => ({ id: z.id, fassung: z.fassung, profile: z.profile.map((p) => ({ id: p.id, fassung: p.fassung })) }))'

// fassungenNachziehen schreibt die gespeicherten Fassungen in die eigenen
// Dinge, soweit es sie schon gibt.
function fassungenNachziehen(liste, k) {
  for (const z of liste) {
    if (n2(k) && z.id === n2(k).id) n2(k).fassung = z.fassung
    for (const p of z.profile) if (p2(k) && p.id === p2(k).id) p2(k).fassung = p.fassung
  }
}

// bildantwort legt ein Bild ab und beantwortet damit den **nächsten**
// Dateidialog — `vertrag.mjs` bricht jeden Dialog sonst ab.
const bildantwort = (name) => async (p, k) => {
  const datei = join(k.ort.ort, 'vertrag', name)
  schreibePng(datei)
  await p.frage({ was: 'dialogantwort', datei })
}

// zielantwort beantwortet einen Speichern-Dialog mit einem Ort im
// Zwischenverzeichnis der Probe.
const zielantwort = (name) => async (p, k) => {
  await p.frage({ was: 'dialogantwort', datei: join(k.ort.ort, 'vertrag', name) })
}

// tokenwunsch ist der Verbund für Anlegen und Ändern. `profilID` wird beim
// Ändern nicht gelesen, `partie` und `sitz` nur beim Anlegen.
const tokenwunsch = (k, aenderung = {}) => ({
  id: '', profilID: p2(k).id, bezeichnung: 'Vertragstoken', principal: 'Rundgang', scope: 'lesen',
  ablauf: '', partie: '', sitz: '', fassung: '', ...aenderung,
})

// vorgabe ist die Art, die das zweite Profil überhaupt wählen kann.
const vorgabe = (k) => b(k).vorgabe

export const bNutzer = {
  name: 'Nutzer, Profile, Freigaben, Token, Zugänge',
  rufe: [
    // **Wer arbeitet, wird gelesen und nicht geraten** — die Liste der Nutzer
    // ist nach Namen geordnet, ihr erster Eintrag ist nicht verlässlich der
    // arbeitende. `k.nutzer` und `k.profil` lesen auch die Ersatzargumente.
    { name: 'WerArbeitet', nimm: 'a.nutzerID', nach: (w, k) => { b(k).ichID = w; k.nutzer = w } },
    { name: 'GezeigtesProfil', nimm: 'a.profil.id', nach: (w, k) => { b(k).ichProfil = w; k.profil = w } },

    // --- der zweite Mensch ---
    { name: 'NutzerAnlegen', argumente: () => ['Vertragsnutzer B', ''], nimm: nutzerform, nach: (w, k) => { b(k).n2 = w } },
    // **Die Antwort des Anlegens trägt Fassung 0**, der Bestand führt 1 — wer
    // sie zurückgibt, bekommt „der Stand ist überholt" (am App-Bau gefunden,
    // 2026-10-09). Die Übersicht nennt die gespeicherte.
    { name: 'Nutzer', nimm: uebersicht, nach: (w, k) => fassungenNachziehen(w, k) },
    {
      name: 'NutzerAendern',
      argumente: (k) => [{
        id: n2(k).id, name: n2(k).name, bild: '', anrede: '', namenszusatz: '', vorname: 'Vera', weiterenamen: '',
        nachname: 'Vertrag', slogan: 'aus dem Rundgang', kurzebio: '', langebio: '', fassung: n2(k).fassung,
      }],
      nimm: 'a.nutzer.fassung', nach: (w, k) => { n2(k).fassung = w },
    },
    // **Ein Dialog, den die Probe beantwortet**: Bild wählen, Bild entfernen.
    {
      name: 'NutzerbildWaehlen', vorher: bildantwort('b-nutzerbild.png'),
      argumente: (k) => [n2(k).id, n2(k).fassung], nimm: 'a.nutzer.fassung', nach: (w, k) => { n2(k).fassung = w },
    },
    { name: 'NutzerbildEntfernen', argumente: (k) => [n2(k).id, n2(k).fassung], nimm: 'a.nutzer.fassung', nach: (w, k) => { n2(k).fassung = w } },
    // **Auf `system` zurück**, damit nichts davon im Fenster stehen bleibt;
    // `hoch` ist die Übersteuerung, die ein Mensch bewusst wählt.
    { name: 'NutzereinstellungSetzen', argumente: (k) => [n2(k).id, 'barrierefreiheit.kontrast', 'hoch'] },
    { name: 'NutzereinstellungSetzen', argumente: (k) => [n2(k).id, 'barrierefreiheit.kontrast', 'system'] },
    { name: 'PersonendatenSichern', vorher: zielantwort('b-personendaten.json'), argumente: (k) => [n2(k).id] },

    // --- das zweite Profil des arbeitenden Menschen ---
    { name: 'ProfilAnlegen', argumente: (k) => [b(k).ichID, 'Vertragsprofil B'], nimm: profilform, nach: (w, k) => { b(k).p2 = w } },
    { name: 'Nutzer', nimm: uebersicht, nach: (w, k) => fassungenNachziehen(w, k) },
    {
      name: 'ProfilAendern',
      argumente: (k) => [{
        id: p2(k).id, name: p2(k).name, subslogan: 'Rundgang', kurzebeschreibung: 'aus dem Rundgang', langebeschreibung: '',
        badge: '', technisch: false, ablauf: '', fassung: p2(k).fassung,
      }],
      nimm: 'a.profil.fassung', nach: (w, k) => { p2(k).fassung = w },
    },
    {
      name: 'ProfilbadgeWaehlen', vorher: bildantwort('b-profilbadge.png'),
      argumente: (k) => [p2(k).id, p2(k).fassung], nimm: 'a.profil.fassung', nach: (w, k) => { p2(k).fassung = w },
    },
    { name: 'ProfilbadgeEntfernen', argumente: (k) => [p2(k).id, p2(k).fassung], nimm: 'a.profil.fassung', nach: (w, k) => { p2(k).fassung = w } },

    // **Freigaben**: der zweite Mensch bekommt Leserrechte am zweiten Profil
    // und verliert sie wieder. Nur ein Verwalter darf das — der Eigner ist
    // einer, auch ohne Zeile im Bestand.
    { name: 'FreigabeSetzen', argumente: (k) => [{ profilID: p2(k).id, nutzerID: n2(k).id, rolle: 'leser' }] },
    { name: 'FreigabeSetzen', argumente: (k) => [{ profilID: p2(k).id, nutzerID: n2(k).id, rolle: 'mitarbeiter' }] },
    { name: 'FreigabeLoeschen', argumente: (k) => [p2(k).id, n2(k).id] },

    // **Vorgaben**: gesetzt wird ein Wert, den das Angebot **nennt**; gegen das
    // Angebot wird beim Setzen geprüft. Gibt es keines, wird nur abgeräumt
    // (eine leere Liste ist eine Aussage und kein Auslassen).
    {
      name: 'Profilvorgaben', argumente: (k) => [p2(k).id],
      nimm: 'a.vorgaben.filter((v) => v.moeglich.length > 0).map((v) => ({ art: v.art, wert: v.moeglich[0].wert }))',
      nach: (w, k) => { b(k).vorgabe = w[0] ?? null },
    },
    { name: 'ProfilvorgabeSetzen', argumente: (k) => [{ profilID: p2(k).id, art: vorgabe(k)?.art ?? 'lagerort', werte: vorgabe(k) ? [vorgabe(k).wert] : [] }] },
    { name: 'ProfilvorgabeSetzen', argumente: (k) => [{ profilID: p2(k).id, art: vorgabe(k)?.art ?? 'lagerort', werte: [] }] },

    // **Token**: der Klartext steht nur in der Antwort des Anlegens; hier wird
    // er nicht gebraucht. Das zweite wird vom Sammelwiderruf genommen.
    { name: 'TokenAnlegen', argumente: (k) => [tokenwunsch(k)], nimm: tokenform, nach: (w, k) => { b(k).t1 = w } },
    {
      name: 'TokenAendern',
      argumente: (k) => [tokenwunsch(k, { id: t1(k).id, bezeichnung: 'Vertragstoken neu', fassung: t1(k).fassung })],
      nimm: 'a.token.fassung', nach: (w, k) => { t1(k).fassung = w },
    },
    { name: 'TokenRotieren', argumente: (k) => [t1(k).id], nimm: 'a.token.fassung', nach: (w, k) => { t1(k).fassung = w } },
    { name: 'TokenAnlegen', argumente: (k) => [tokenwunsch(k, { bezeichnung: 'Zweites Vertragstoken' })], nimm: tokenform, nach: (w, k) => { b(k).t2 = w } },
    { name: 'TokenLoeschen', argumente: (k) => [t1(k).id] },
    { name: 'TokenAlleWiderrufen', argumente: (k) => [p2(k).id] },

    // **Zugänge** liegen im Schlüsselbund der Probe (`MSE_SCHLUESSELBUND`),
    // nicht im Anmeldebund dieses Rechners. Der Wert ist ein Testwert.
    { name: 'ZugangSetzen', argumente: () => ['mtgjson', 'marke', 'vertrag-testmarke'] },
    { name: 'ZugangLoeschen', argumente: () => ['mtgjson', 'marke'] },

    // **Nicht gelingbar**: siehe den Kopf dieser Datei.
    {
      name: 'KontoLoesen', argumente: (k) => [b(k).ichID, 'github'], erwartet: 'fachfehler',
      grund: 'ohne verknüpftes Anbieterkonto (storage/anbieterkonten.go:130, nur über Anmelden im Browser zu erzeugen)',
    },

    // --- weg, was diese Gruppe angelegt hat: erst das Profil, dann der Mensch ---
    { name: 'ProfilLoeschen', argumente: (k) => [p2(k).id, p2(k).name] },
    { name: 'NutzerLoeschen', argumente: (k) => [n2(k).id, n2(k).name] },
  ],
}
