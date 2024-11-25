import { getApi } from './apiHelper'

export const globalClustersService = {
  getClusters,
  getClusterPeers,
  getMonitoredData,
  switchGlobalSetting,
  setGlobalSetting,
  addCluster
}

function getClusterPeers() {
  return getApi().get('clusters/peers')
}

function getClusters() {
  return getApi().get('clusters')
}

function getMonitoredData() {
  return getApi().get('monitor')
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
