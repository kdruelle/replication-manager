import { getApi } from './apiHelper'

export const configService = {
  addDBTag,
  dropDBTag,
  addProxyTag,
  dropProxyTag
}

function addDBTag(clusterName, tag, baseURL) {
  return getApi(baseURL).get(`clusters/${clusterName}/settings/actions/add-db-tag/${tag}`)
}

function dropDBTag(clusterName, tag, baseURL) {
  return getApi(baseURL).get(`clusters/${clusterName}/settings/actions/drop-db-tag/${tag}`)
}

function addProxyTag(clusterName, tag, baseURL) {
  return getApi(baseURL).get(`clusters/${clusterName}/settings/actions/add-proxy-tag/${tag}`)
}

function dropProxyTag(clusterName, tag, baseURL) {
  return getApi(baseURL).get(`clusters/${clusterName}/settings/actions/drop-proxy-tag/${tag}`)
}
