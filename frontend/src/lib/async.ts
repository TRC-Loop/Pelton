// async.ts defines the small state shape every data-loading store uses so the ui
// can render loading, error and empty states explicitly and consistently instead
// of leaving panes blank.

export type Status = 'idle' | 'loading' | 'ready' | 'error'

export interface AsyncState<T> {
  status: Status
  data: T | null
  error: string
}

// idle is the starting state for an async resource.
export function idle<T>(): AsyncState<T> {
  return { status: 'idle', data: null, error: '' }
}

// loading preserves any existing data so a refresh does not blank the view.
export function loading<T>(prev: AsyncState<T>): AsyncState<T> {
  return { status: 'loading', data: prev.data, error: '' }
}

export function ready<T>(data: T): AsyncState<T> {
  return { status: 'ready', data, error: '' }
}

export function failed<T>(error: string): AsyncState<T> {
  return { status: 'error', data: null, error }
}
