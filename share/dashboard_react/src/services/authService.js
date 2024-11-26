import { getApi } from './apiHelper'

export const authService = {
  login,
  gitLogin
}

function login(username, password, baseURL) {
  return getApi(baseURL).post('login', { username, password })
}

function gitLogin(username, password, baseURL) {
  return getApi(baseURL).post('login-git', { username, password })
}
