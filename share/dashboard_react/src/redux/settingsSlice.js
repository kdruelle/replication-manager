import { createSlice, createAsyncThunk, isAnyOf } from '@reduxjs/toolkit'
import { handleError, showErrorBanner, showLoaderBanner, showSuccessBanner } from '../utility/common'
import { settingsService } from '../services/settingsService'

export const switchSetting = createAsyncThunk('settings/switchSetting', async ({ clusterName, setting }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await settingsService.switchSettings(clusterName, setting, baseURL)
    // if (setting === 'monitoring-scheduler') {
    //   await clusterService.getClusterData(clusterName)
    // }
    showSuccessBanner(`Switching ${setting} successful!`, status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner(`Switching ${setting} failed!`, error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const changeTopology = createAsyncThunk(
  'settings/changeTopology',
  async ({ clusterName, topology }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await settingsService.changeTopology(clusterName, topology, baseURL)
      showSuccessBanner(`Topology changed to ${topology} successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Changing topology to ${setting} failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const setSetting = createAsyncThunk('settings/setSetting', async ({ clusterName, setting, value }, thunkAPI) => {
  try {
    // showLoaderBanner(`${setting} `, thunkAPI)
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await settingsService.setSetting(clusterName, setting, value, baseURL)
    showSuccessBanner(`${setting} changed successfully!`, status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner(`Changing ${setting} failed!`, error.toString(), thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const updateGraphiteWhiteList = createAsyncThunk(
  'settings/updateGraphiteWhiteList',
  async ({ clusterName, whiteListValue }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await settingsService.updateGraphiteWhiteList(clusterName, whiteListValue, baseURL)
      showSuccessBanner(`Graphite Whitelist Regexp updated successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Updating Graphite Whitelist Regexp failed!`, error.toString(), thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const updateGraphiteBlackList = createAsyncThunk(
  'settings/updateGraphiteBlackList',
  async ({ clusterName, blackListValue }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await settingsService.updateGraphiteBlackList(clusterName, blackListValue, baseURL)
      showSuccessBanner(`Graphite BlackList Regexp updated successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Updating Graphite BlackList Regexp failed!`, error.toString(), thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

const initialState = {}

export const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    clearSettings: (state, action) => {
      Object.assign(state, initialState)
    }
  }
})

export const { clearSettings } = settingsSlice.actions

// this is for configureStore
export default settingsSlice.reducer
