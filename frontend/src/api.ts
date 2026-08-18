export type DictionaryEntry = {
  Headword: string
  HTML: string
  CSS: string
}

const browserEntries: Record<string, DictionaryEntry> = {
  '你好': {
    Headword: '你好',
    HTML: '<p>hello</p><p>hi</p>',
    CSS: '.dictionary-entry { font-family: Georgia, serif; }',
  },
  hello: {
    Headword: 'hello',
    HTML: '<p>a greeting used when meeting someone</p>',
    CSS: '',
  },
}

type WailsApp = {
  LookupWord(word: string): Promise<DictionaryEntry>
  ChooseAndOpenDictionary(): Promise<string>
}

declare global {
  interface Window {
    go?: { main?: { App?: WailsApp } }
  }
}

export async function chooseAndOpenDictionary(): Promise<string | null> {
  const app = window.go?.main?.App
  if (!app) return null
  return app.ChooseAndOpenDictionary()
}

export async function lookupWord(word: string): Promise<DictionaryEntry | null> {
  const app = window.go?.main?.App
  if (app) {
    return app.LookupWord(word)
  }
  return browserEntries[word.trim()] ?? null
}
