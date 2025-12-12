import axios from 'axios';
import {API_URL} from '@env';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';
const http = axios.create({
  baseURL: API_URL,
});

http.interceptors.request.use(
  async (config: any) => {
    if (typeof global.token != 'undefined' && global?.token !== '') {
      config.headers.Authorization = `Bearer ${global.token}`;
      // if (global.default) {
      //   config.headers['X-Minerva-RBAC-Domain'] = `${global.rbac_domain}`;
      // }
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  },
);

http.interceptors.response.use(
  (response: any) => {
    if (response.data && response.data?.error) {
      const errorMessage = response.data.error;
      if (errorMessage === 'EOF') {
        return Promise.reject(
          'Sorry, but the data you looking for is not exist',
        );
      } else if (
        errorMessage.indexOf('invalid access token') > 0 ||
        errorMessage.indexOf('invalid token') > 0
      ) {
        Toast.show({
          type: ALERT_TYPE.DANGER,
          title: 'Error',
          textBody: errorMessage,
        });
        return Promise.reject(errorMessage);
      }
      Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody: errorMessage || 'Something went wrong!!',
      });
      return Promise.reject(errorMessage);
    }
    return Promise.resolve(response);
    // return response;
  },
  (error: any) => {
    // console.log(error, typeof error);
    Toast.show({
      type: ALERT_TYPE.DANGER,
      title: 'Error',
      textBody: error?.response?.data || 'Something went wrong!!',
    });
    return Promise.reject(error);
  },
);

export default http;
