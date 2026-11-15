// **Eine Adresse, die PixiJS unter `wails://` richtig auflöst** — #2066.
//
// `path.toAbsolute` aus pixi.js 8.20.1 kennt als URL nur `http(s):`
// (`isUrl`). Unter der Basis des nativen Fensters, `wails://localhost/`,
// gibt `rootname` deshalb `wails://`, und `/spielbild/x.png` wurde zu
// `wails://spielbild/x.png` — der Pfad als Host. Eine Adresse **mit** Schema
// lässt `toAbsolute` dagegen unverändert (`isAbsolute`); also bekommt Pixi
// sie schon absolut, aufgelöst vom Browser und nicht von Pixi.

// fuerPixi löst `adresse` gegen `basis` auf — gewöhnlich `document.baseURI`.
export function fuerPixi(adresse: string, basis: string): string {
  return new URL(adresse, basis).href
}
