// Typdeklaration für Stilimporte als Nebenwirkung.
//
// `import './stil.css'` ist kein Modul im Sinne von TypeScript; ohne diese
// Deklaration bricht `tsc --noEmit` mit TS2882 ab. Vite löst den Import zur
// Bauzeit auf — TypeScript muss nur wissen, dass es ihn geben darf.
declare module '*.css'
