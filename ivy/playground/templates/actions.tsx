// ivy imports
import { client } from '../core'
import { generateInputMap } from '../contracts/selectors'

// internal imports
import { INITIAL_ID_LIST } from './constants'
import { getSourceMap, hasSourceChanged } from './selectors'
import { CompiledTemplate } from './types'

export const loadTemplate = (selected: string) => {
  return (dispatch, getState) => {
    if (!selected) {
      selected = INITIAL_ID_LIST[1]
    }
    const state = getState()
    const source = getSourceMap(state)[selected]
    dispatch(setSource(source))
  }
}

export const SHOW_LOCK_INPUT_ERRORS = 'templates/SHOW_LOCK_INPUT_ERRORS'

export const showLockInputErrors = (result: boolean) => {
  return {
    type: SHOW_LOCK_INPUT_ERRORS,
    result
  }
}

export const UPDATE_LOCK_ERROR = 'templates/UPDATE_LOCK_ERROR'

export const updateLockError = (error?) => {
  return {
    type: UPDATE_LOCK_ERROR,
    error
  }
}

export const SET_SOURCE = 'templates/SET_SOURCE'

export const setSource = (source: string) => {
  return (dispatch, getState) => {
    const type = SET_SOURCE
    const sourceChanged = hasSourceChanged(source)(getState())
    dispatch({ type, source, sourceChanged })
    dispatch(fetchCompiled(source))
    dispatch(updateLockError())
  }
}

export const FETCH_COMPILED = 'templates/FETCH_COMPILED'

export const fetchCompiled = (source: string) => {
  return (dispatch, getState) => {
    return client.ivy.compile({ source }).then((result) => {
      const type = FETCH_COMPILED
      const format = (tpl: CompiledTemplate) => {
        if (tpl.error !== '') {
          tpl.clauseInfo = tpl.params = []
        }
        return tpl
      }
      const compiled = format(result)
      const inputMap = generateInputMap(compiled[0])
      dispatch({ type, compiled, inputMap })
    }).catch((e) => {throw e})
  }
}

export const SAVE_TEMPLATE = 'templates/SAVE_TEMPLATE'

export const saveTemplate = () => ({ type: SAVE_TEMPLATE })
