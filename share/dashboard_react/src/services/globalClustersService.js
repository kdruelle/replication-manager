import { getApi } from './apiHelper'

export const globalClustersService = {
  getClusters,
  getClusterPeers,
  getMonitoredData,
  switchGlobalSetting,
  setGlobalSetting,
  addCluster
}

function getClusterPeers(baseURL) {
  return getApi(baseURL).get('clusters/peers')
}

function getClusters(baseURL) {
  return getApi(baseURL).get('clusters')
}

function getMonitoredData(baseURL) {
  return getApi(baseURL).get('monitor')
}

function switchGlobalSetting(setting) {
  return getApi().get(`clusters/settings/actions/switch/${setting}`)
}

function setGlobalSetting(setting, value) {
  return getApi().get(`clusters/settings/actions/set/${setting}/${value}`)
}

function addCluster(clusterName, formdata) {
  return getApi().post(`clusters/actions/add/${clusterName}`, formdata)
}
