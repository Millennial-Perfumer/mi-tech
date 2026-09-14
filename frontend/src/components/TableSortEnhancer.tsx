import { useEffect, useRef, type ReactNode } from 'react'

type SortDirection = 'asc' | 'desc'

type SortState = {
  columnIndex: number
  direction: SortDirection
}

type SortableRow = {
  row: HTMLTableRowElement
  index: number
}

function cleanCellValue(value: string) {
  return value.replace(/\s+/g, ' ').trim()
}

function numericValue(value: string) {
  const normalised = value
    .replace(/[₹,%]/g, '')
    .replace(/[×x]$/i, '')
    .replace(/\s*(kg|g)$/i, '')
    .replace(/,/g, '')
    .trim()

  if (!/^-?\d+(\.\d+)?$/.test(normalised)) return undefined
  return Number(normalised)
}

function dateValue(value: string, label: string) {
  const looksLikeDate = /date|time|created|added|published|sent|manufactured|activity/i.test(label)
  if (!looksLikeDate) return undefined

  const parsed = Date.parse(value)
  return Number.isNaN(parsed) ? undefined : parsed
}

function compareCellValues(leftText: string, rightText: string, label: string, direction: SortDirection) {
  const left = cleanCellValue(leftText)
  const right = cleanCellValue(rightText)
  const leftEmpty = !left || left === '—' || left === '-'
  const rightEmpty = !right || right === '—' || right === '-'

  // Keep missing values at the bottom in either direction, which is easier to scan.
  if (leftEmpty || rightEmpty) {
    if (leftEmpty && rightEmpty) return 0
    return leftEmpty ? 1 : -1
  }

  const leftDate = dateValue(left, label)
  const rightDate = dateValue(right, label)
  const leftNumber = numericValue(left)
  const rightNumber = numericValue(right)
  let result: number

  if (leftDate !== undefined && rightDate !== undefined) {
    result = leftDate - rightDate
  } else if (leftNumber !== undefined && rightNumber !== undefined) {
    result = leftNumber - rightNumber
  } else {
    result = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' }).compare(left, right)
  }

  return direction === 'asc' ? result : -result
}

function tableRows(table: HTMLTableElement): SortableRow[] {
  const body = table.tBodies[0]
  if (!body) return []

  return Array.from(body.rows)
    .map((row, index) => ({ row, index }))
    .filter(({ row }) => !row.querySelector('.table-state') && row.cells.length > 0)
}

function updateHeaderState(table: HTMLTableElement, state: SortState) {
  Array.from(table.tHead?.rows[0]?.cells || []).forEach((cell, index) => {
    const header = cell as HTMLTableCellElement
    const button = header.querySelector<HTMLButtonElement>('.sortable-table-button')
    if (!button) return

    const active = index === state.columnIndex
    header.setAttribute('aria-sort', active ? (state.direction === 'asc' ? 'ascending' : 'descending') : 'none')
    const indicator = button.querySelector<HTMLElement>('.sortable-table-indicator')
    if (indicator) indicator.textContent = active ? (state.direction === 'asc' ? '↑' : '↓') : '↕'
    button.setAttribute('aria-label', `${button.dataset.label || button.textContent || 'Column'}; ${active ? `${state.direction === 'asc' ? 'ascending' : 'descending'}, ` : ''}activate to sort`)
  })
}

function sortTable(table: HTMLTableElement, state: SortState) {
  const header = table.tHead?.rows[0]?.cells[state.columnIndex]
  const body = table.tBodies[0]
  if (!header || !body) return

  const label = cleanCellValue(header.textContent || '')
  const rows = tableRows(table)
  rows.sort((left, right) => {
    const result = compareCellValues(
      left.row.cells[state.columnIndex]?.textContent || '',
      right.row.cells[state.columnIndex]?.textContent || '',
      label,
      state.direction,
    )
    return result || left.index - right.index
  })

  rows.forEach(({ row }) => body.appendChild(row))
  updateHeaderState(table, state)
}

function enhanceTable(table: HTMLTableElement, states: WeakMap<HTMLTableElement, SortState>) {
  const headerRow = table.tHead?.rows[0]
  if (!headerRow) return
  // Tables that already provide their own React sort controls (for example GST
  // summaries) should keep their local state and visual treatment.
  if (headerRow.querySelector('button')) return

  Array.from(headerRow.cells).forEach((cell, columnIndex) => {
    const header = cell as HTMLTableCellElement
    if (header.querySelector('button, input, select') || /actions?/i.test(header.textContent || '')) return

    const label = cleanCellValue(header.textContent || '')
    if (!label) return

    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'sortable-table-button'
    button.dataset.label = label
    button.setAttribute('aria-label', `${label}; activate to sort`)

    const text = document.createElement('span')
    text.textContent = label
    const indicator = document.createElement('span')
    indicator.className = 'sortable-table-indicator'
    indicator.setAttribute('aria-hidden', 'true')
    indicator.textContent = '↕'
    button.append(text, indicator)
    header.replaceChildren(button)
    header.setAttribute('aria-sort', 'none')

    button.addEventListener('click', () => {
      const previous = states.get(table)
      const state: SortState = previous?.columnIndex === columnIndex
        ? { columnIndex, direction: previous.direction === 'asc' ? 'desc' : 'asc' }
        : { columnIndex, direction: 'asc' }
      states.set(table, state)
      sortTable(table, state)
    })
  })

  const state = states.get(table)
  if (state) updateHeaderState(table, state)
}

export function TableSortEnhancer({ children }: { children: ReactNode }) {
  const scopeRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const scope = scopeRef.current
    if (!scope) return undefined

    const states = new WeakMap<HTMLTableElement, SortState>()
    let frame = 0
    const enhance = () => {
      frame = 0
      scope.querySelectorAll<HTMLTableElement>('table.orders-table').forEach((table) => enhanceTable(table, states))
    }
    const scheduleEnhance = () => {
      if (frame) return
      frame = window.requestAnimationFrame(enhance)
    }

    scheduleEnhance()
    const observer = new MutationObserver(scheduleEnhance)
    observer.observe(scope, { childList: true, characterData: true, subtree: true })

    return () => {
      observer.disconnect()
      if (frame) window.cancelAnimationFrame(frame)
    }
  }, [])

  return <div ref={scopeRef} className="table-sort-scope">{children}</div>
}
