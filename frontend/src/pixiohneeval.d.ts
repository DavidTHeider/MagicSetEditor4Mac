// **Ein Modul, das nur durch sein Laden wirkt** — SEC-17 (#2164).
//
// `pixi.js` führt `./unsafe-eval` in seiner Exporttabelle ohne `types`, und
// neben `lib/unsafe-eval/init.mjs` liegt nur ein `init.d.ts`, kein
// `init.d.mts` (8.20.1); unter `moduleResolution: bundler` findet TypeScript
// es deshalb nicht. Es gibt nichts zurück (`export {}`), also steht hier
// nichts außer seinem Namen.
declare module 'pixi.js/unsafe-eval'
