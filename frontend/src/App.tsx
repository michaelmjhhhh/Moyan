import { FormEvent, useEffect, useState } from 'react'
import { DictionaryEntry, lookupWord } from './api'

type Dictionary = {
  id: string
  name: string
  language: string
  enabled: boolean
}

type Entry = DictionaryEntry & {
  dictionary: string
  pronunciation?: string
}

const dictionaries: Dictionary[] = [
  { id: 'cc-cedict', name: 'CC-CEDICT', language: '中英', enabled: true },
  { id: 'moyan-sample', name: 'Moyan Sample', language: '英英', enabled: false },
]

export function App() {
  const [query, setQuery] = useState('你好')
  const [submittedQuery, setSubmittedQuery] = useState('你好')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [history, setHistory] = useState(['你好'])
  const [results, setResults] = useState<Entry[]>([])
  const [isSearching, setIsSearching] = useState(false)

  useEffect(() => {
    let cancelled = false
    setIsSearching(true)
    lookupWord(submittedQuery)
      .then((entry) => {
        if (cancelled) return
        setResults(entry ? [{ ...entry, dictionary: 'CC-CEDICT' }] : [])
      })
      .catch(() => {
        if (!cancelled) setResults([])
      })
      .finally(() => {
        if (!cancelled) setIsSearching(false)
      })
    return () => { cancelled = true }
  }, [submittedQuery])

  function submitSearch(event: FormEvent) {
    event.preventDefault()
    const value = query.trim()
    if (!value) return
    setSubmittedQuery(value)
    setHistory((current) => [value, ...current.filter((item) => item !== value)].slice(0, 12))
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <button className="wordmark" type="button" onClick={() => setSidebarOpen((open) => !open)}>
          <span className="wordmark-mark">M</span>
          <span>Moyan</span>
        </button>
        <form className="search-form" onSubmit={submitSearch}>
          <label className="search-label" htmlFor="word-search">查词</label>
          <input
            id="word-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="输入词头"
            autoComplete="off"
            spellCheck={false}
          />
          <button className="search-submit" type="submit" aria-label="搜索">↵</button>
        </form>
        <button className="quiet-button" type="button">设置</button>
      </header>

      <div className="workspace">
        {sidebarOpen && (
          <aside className="sidebar" aria-label="词库与历史">
            <section className="sidebar-section">
              <div className="section-heading">
                <span>词库</span>
                <button className="text-button" type="button">导入</button>
              </div>
              <div className="dictionary-list">
                {dictionaries.map((dictionary) => (
                  <button
                    className={`dictionary-item${dictionary.enabled ? ' is-enabled' : ''}`}
                    key={dictionary.id}
                    type="button"
                  >
                    <span className="status-dot" aria-hidden="true" />
                    <span>
                      <strong>{dictionary.name}</strong>
                      <small>{dictionary.language}</small>
                    </span>
                  </button>
                ))}
              </div>
            </section>

            <section className="sidebar-section history-section">
              <div className="section-heading"><span>近期</span></div>
              <div className="history-list">
                {history.map((item) => (
                  <button
                    className={`history-item${item === submittedQuery ? ' is-current' : ''}`}
                    key={item}
                    type="button"
                    onClick={() => { setQuery(item); setSubmittedQuery(item) }}
                  >
                    {item}
                  </button>
                ))}
              </div>
            </section>
          </aside>
        )}

        <section className="reading-pane" aria-live="polite">
          <div className="reading-header">
            <span className="result-kicker">{isSearching ? '查询中' : results.length ? '词条' : '未找到'}</span>
            <span className="result-query">{submittedQuery}</span>
          </div>
          {results.length ? (
            <div className="entries">
              {results.map((entry) => (
                <article className="entry" key={`${entry.dictionary}-${entry.Headword}`}>
                  <div className="entry-source">{entry.dictionary}</div>
                  <h1>{entry.Headword}</h1>
                  {entry.pronunciation && <div className="pronunciation">{entry.pronunciation}</div>}
                  <DictionaryFrame html={entry.HTML} />
                </article>
              ))}
            </div>
          ) : (
            <div className="empty-state">
              <p>没有找到“{submittedQuery}”</p>
              <span>试试其他词头，或导入一本词典。</span>
            </div>
          )}
        </section>
      </div>
    </main>
  )
}

function DictionaryFrame({ html }: { html: string }) {
  const srcDoc = `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:; media-src data:"></head><body>${html}</body></html>`
  return (
    <iframe
      className="dictionary-frame"
      title="词典正文"
      sandbox=""
      srcDoc={srcDoc}
    />
  )
}
