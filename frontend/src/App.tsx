import { FormEvent, useEffect, useRef, useState } from 'react'
import { chooseAndOpenDictionary, DictionaryEntry, lookupWords, searchCandidates } from './api'

type Dictionary = {
  id: string
  name: string
  language: string
  enabled: boolean
}

type Entry = DictionaryEntry & {
  pronunciation?: string
}

const initialDictionaries: Dictionary[] = []

export function App() {
  const [query, setQuery] = useState('')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [selectedDictionary, setSelectedDictionary] = useState<string | null>(null)
  const [history, setHistory] = useState<string[]>([])
  const [dictionaries, setDictionaries] = useState(initialDictionaries)
  const [results, setResults] = useState<Entry[]>([])
  const [candidates, setCandidates] = useState<string[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importStatus, setImportStatus] = useState('')
  const searchInputRef = useRef<HTMLInputElement>(null)
  const enabledNames = new Set(dictionaries.filter((dictionary) => dictionary.enabled).map((dictionary) => dictionary.name))
  const visibleResults = results.filter((entry) => enabledNames.has(entry.Dictionary) && (!selectedDictionary || entry.Dictionary === selectedDictionary))

  useEffect(() => {
    if (!submittedQuery) {
      setResults([])
      setCandidates([])
      setIsSearching(false)
      return
    }
    let cancelled = false
    setIsSearching(true)
    setCandidates([])
    async function search() {
      try {
        const entries = await lookupWords(submittedQuery)
        if (cancelled) return
        setResults(entries)
        if (entries.length === 0) {
          setCandidates(await searchCandidates(submittedQuery))
        }
      } catch {
        if (!cancelled) setResults([])
      } finally {
        if (!cancelled) setIsSearching(false)
      }
    }
    void search()
    return () => { cancelled = true }
  }, [submittedQuery])

  useEffect(() => {
    function handleShortcut(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchInputRef.current?.focus()
        searchInputRef.current?.select()
      }
    }
    window.addEventListener('keydown', handleShortcut)
    return () => window.removeEventListener('keydown', handleShortcut)
  }, [])

  function toggleDictionary(id: string) {
    const dictionary = dictionaries.find((item) => item.id === id)
    if (!dictionary) return
    setDictionaries((current) => current.map((item) => item.id === id ? { ...item, enabled: !item.enabled } : item))
    if (dictionary.enabled && selectedDictionary === dictionary.name) {
      setSelectedDictionary(null)
    }
  }

  async function importDictionary() {
    setIsImporting(true)
    setImportStatus('正在导入词典…')
    try {
      const name = await chooseAndOpenDictionary()
      if (!name) {
        setImportStatus('已取消导入')
        return
      }
      setDictionaries((current) => current.some((dictionary) => dictionary.name === name)
        ? current.map((dictionary) => dictionary.name === name ? { ...dictionary, enabled: true } : dictionary)
        : [...current, { id: name, name, language: 'MDX', enabled: true }])
      setSelectedDictionary(name)
      setImportStatus(`已导入：${name}`)
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      setImportStatus(`导入失败：${message}`)
    } finally {
      setIsImporting(false)
    }
  }

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
        <form className="search-form" role="search" onSubmit={submitSearch}>
          <label className="search-label" htmlFor="word-search">查词</label>
          <input
            id="word-search"
            ref={searchInputRef}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="输入词头"
            autoComplete="off"
            spellCheck={false}
          />
          <button className="search-submit" type="submit" aria-label="搜索" disabled={isSearching}>↵</button>
        </form>
        <button className="quiet-button" type="button">设置</button>
      </header>

      <div className="workspace">
        {sidebarOpen && (
          <aside className="sidebar" aria-label="词库与历史">
            <section className="sidebar-section">
              <div className="section-heading">
                <span>词库</span>
                <span className="section-actions">
                  <button className={`text-button${selectedDictionary === null ? ' is-active' : ''}`} type="button" onClick={() => setSelectedDictionary(null)} aria-pressed={selectedDictionary === null}>全部</button>
                  <button className="text-button" type="button" onClick={importDictionary} disabled={isImporting}>
                    {isImporting ? '导入中…' : '导入'}
                  </button>
                </span>
              </div>
              <div className="dictionary-list">
                {dictionaries.length > 0 ? dictionaries.map((dictionary) => (
                  <div className="dictionary-row" key={dictionary.id}>
                    <button
                      className={`dictionary-item${dictionary.enabled ? ' is-enabled' : ''}${selectedDictionary === dictionary.name ? ' is-selected' : ''}`}
                      type="button"
                      aria-pressed={selectedDictionary === dictionary.name}
                      onClick={() => setSelectedDictionary((current) => current === dictionary.name ? null : dictionary.name)}
                    >
                      <span className="status-dot" aria-hidden="true" />
                      <span className="dictionary-label">
                        <strong>{dictionary.name}</strong>
                        <small>{dictionary.language}</small>
                      </span>
                    </button>
                    <button
                      className={`dictionary-toggle${dictionary.enabled ? ' is-enabled' : ''}`}
                      type="button"
                      aria-pressed={dictionary.enabled}
                      aria-label={`${dictionary.name}${dictionary.enabled ? '：停用' : '：启用'}`}
                      onClick={() => toggleDictionary(dictionary.id)}
                    >
                      {dictionary.enabled ? '停用' : '启用'}
                    </button>
                  </div>
                )) : (
                  <div className="dictionary-empty">还没有词典<br /><span>点击右上角“导入”开始</span></div>
                )}
              </div>
            </section>

            {importStatus && <p className="status-message" role="status" aria-live="polite">{importStatus}</p>}

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
            <span className="result-kicker">{isSearching ? '查询中' : visibleResults.length ? '词条' : submittedQuery ? '未找到' : '准备就绪'}</span>
            <span className="result-query">{submittedQuery || '导入词典后开始查词'}</span>
          </div>
          {visibleResults.length ? (
            <div className="entries">
              {visibleResults.map((entry) => (
                <article className="entry" key={`${entry.Dictionary}-${entry.Headword}`}>
                  <div className="entry-source">{entry.Dictionary}</div>
                  <h1>{entry.Headword}</h1>
                  {entry.pronunciation && <div className="pronunciation">{entry.pronunciation}</div>}
                  <DictionaryFrame html={entry.HTML} css={entry.CSS} />
                </article>
              ))}
            </div>
          ) : (
            <div className="empty-state">
              <p>{!submittedQuery ? '从一本词典开始' : selectedDictionary && results.length ? `“${selectedDictionary}”中没有找到“${submittedQuery}”` : results.length && !visibleResults.length ? '当前词典均已停用' : `没有找到“${submittedQuery}”`}</p>
              {candidates.length > 0 ? (
                <>
                  <span>你是不是想查：</span>
                  <div className="candidate-list">
                    {candidates.map((candidate) => (
                      <button key={candidate} type="button" onClick={() => { setQuery(candidate); setSubmittedQuery(candidate) }}>
                        {candidate}
                      </button>
                    ))}
                  </div>
                </>
              ) : <span>{!submittedQuery ? '点击左侧“导入”，选择一个 MDX 词典。' : '试试其他词头，或导入一本词典。'}</span>}
            </div>
          )}
        </section>
      </div>
    </main>
  )
}

function DictionaryFrame({ html, css }: { html: string; css: string }) {
  const safeCss = css.replace(/<\/style/gi, '<\\/style')
  const srcDoc = `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:; media-src data:; font-src data:"><style>${safeCss}</style></head><body>${html}</body></html>`
  return (
    <iframe
      className="dictionary-frame"
      title="词典正文"
      sandbox=""
      srcDoc={srcDoc}
    />
  )
}
