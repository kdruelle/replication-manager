import { meetApi } from './apiHelper'

export const meetService = {
  loginMeet
}

function loginMeet() {
  return meetApi.post('users/login', null, 1, true)
}
