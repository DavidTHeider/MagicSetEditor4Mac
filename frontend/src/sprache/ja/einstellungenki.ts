// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': 'ローカル AI ランタイム',
  'ki.geraetemodell': 'Apple のデバイス上モデル',
  'ki.vorrangmodell': '優先時のOllamaモデル',
  'ki.adresse': 'AI ランタイムのアドレス',
} as const

export const kihinweise = {
  'ki.laufzeit':
    'AI ランタイムに問い合わせるかどうかを指定します。**オフ**の場合は、このコンピュータに対してであっても、呼び出しは一切行われません。ここから何かがインストールされたり、起動されたりすることはありません。',
  'ki.geraetemodell':
    'カード名やフレーバーテキストなどの小さなテキスト作業に macOS のデバイス上モデルを使うかどうか。**自動**：この Mac で有効（Apple Intelligence がオン、`fm` のライセンスに同意済み）なら組み込まれ、最初に使われます。データはこのコンピュータの外に出ません。**オフ**は戻すまで有効です。1 回のセッションだけオフにするには **AI** で。',
  'ki.vorrangmodell':
    '**優先**の提案が、デバイス上のモデルの次に問い合わせるOllamaモデルです。**空欄**は、固定形式で答えられる最初のモデルを意味し、Ollamaは最後に取得したものを先に挙げます。選んだモデルをOllamaが挙げなくなった場合は通知され、黙って別のモデルに置き換えられることはありません。',
  'ki.adresse':
    'ランタイムが待ち受けているアドレスです。空欄の場合は、このコンピュータ上の初期アドレスになります。このコンピュータ上に**ない**アドレスを指定すると、その先のモデルはすべてローカルではないものとして扱われます。ローカルのみの設定では、何も問い合わせません。',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': 'オフ',
  'ki.laufzeit.an': 'オン',
  'ki.geraetemodell.automatisch': '自動',
  'ki.geraetemodell.aus': 'オフ',
} as const
