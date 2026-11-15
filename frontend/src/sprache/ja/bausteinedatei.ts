export const dateiansicht = {
  ueberschrift: 'ファイル',
  laedt: 'ファイルを確認しています…',
  arten: {
    paket: 'MSEパッケージ',
    bild: '画像',
    vektor: 'ベクトル画像',
    pdf: 'PDF',
    text: 'テキスト',
    archiv: 'アーカイブ',
    unbekannt: '不明な種類',
  },
  name: '名前',
  art: '種類',
  groesse: 'サイズ',
  geaendert: '変更日',
  ort: '場所',
  masse: '{breite} × {hoehe} ポイント',
  seiten: '{anzahl} ページ',
  eintraege: '{anzahl} 件の項目',
  gekuerzt: '先頭部分を表示しています。ファイルはさらに続きます。',
  unsicher:
    '{anzahl} 件の項目がファイルの外を指しているか、リンクです。これらは数えるだけで、開きません。',
  unsicherEintrag: '開きません',
  verschluesselt:
    'パスワード付きのバックアップです。中身は復元したときに初めて表示されます。',
  zuGross:
    'プレビューするには大きすぎます。対応するアプリケーションで開いてください。',
  unlesbar:
    '内容を読み取れません。名前、サイズ、日時は正しく表示されています。',
  keineVorschau:
    'この種類にはプレビューがありません。プログラムの可能性があります。必要な場合は Finder で表示してください。',
  imFinder: 'Finder で表示',
  oeffnen: '開く',
  imFinderFuer: '{name} を Finder で表示',
  oeffnenFuer: '{name} を対応するアプリケーションで開く',
  ansehen: '表示',
  ansehenFuer: '{name} を表示',
  zuklappen: '表示を閉じる',
} as const
