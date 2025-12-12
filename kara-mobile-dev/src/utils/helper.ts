/* eslint-disable no-bitwise */
import {API_URL} from '@env';
import moment from 'moment';
import 'moment/locale/id';
moment().locale('id');

export const lastSeen = (timestamp: string) => {
  return `Aktif ${moment(timestamp).fromNow()}`;
};

// export const currencyFormat = (
//   num: number,
//   prefix: string = 'Rp. ',
//   thousandSeparator: string = '.',
// ) => {
//   return (
//     num
//       ? [
//           prefix,
//           num
//             .toString()
//             .replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1' + thousandSeparator),
//         ].filter(d => d)
//       : [prefix, 0]
//   ).join(' ');
// };
const addNumbSep = (number: string, thouSep: string) => {
  return number.replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1' + thouSep);
};
export const currencyFormat = (number: number, options?: any) => {
  let {thouSep, decSep, decimal} = options || {};
  thouSep = thouSep || ',';
  decSep = decSep || '.';
  decimal = isNaN(decimal) ? 0 : Math.abs(decimal);

  let _nSplits = (Number(number) || 0).toFixed(decimal).split('.');
  let _numbStr = _nSplits[0];
  let _decStr = _nSplits.length > 1 ? _nSplits[1] : '';
  _numbStr = addNumbSep(_numbStr, thouSep);

  return _numbStr + (_decStr ? decSep + _decStr : '');
};
export const numberWithSeparator = (x: number | string | undefined) => {
  if (!x) {
    if (x !== 0) {
      return;
    }
  }

  var parts = x.toString().split('.');
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, '.');
  return parts.join(',');
};
export const numberWithDefaultSeparator = (x: number | string | undefined) => {
  if (!x) {
    if (x !== 0) {
      return;
    }
  }

  var parts = x.toString().split('.');
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  return parts.join('.');
};
export const number1Rounding = (x: number) => {
  if (!x || x === 0) {
    return '0';
  }
  var toString = x.toString();
  var match = toString.match(/^-?\d+(?:\.\d{0,1})?/);
  if (match !== null) {
    return match[0];
  }
  return x.toString(1);
};
export const modulus = (x: number) => {
  if (!x || x === 0) {
    return 0;
  }
  var result = x % 1;
  if (result === 0) {
    return x;
  }
  return parseFloat(x.toString(2));
};
export const getImage = (image_id: string) => {
  return API_URL + '/asset/View?id=' + image_id;
};
export const getAssetURL = (asset_id: string) => {
  return API_URL + '/asset/view?id=' + asset_id;
};

// cek isMineral
export const isMineral = (materials: any) => {
  return (
    materials.filter((x: any) => ['ORE', 'WASTE'].includes(x.Key.toUpperCase()))
      .length > 0
  );
};

// get url param
export const getUrlParam = (url: string, param: string) => {
  const include = url.includes(param);
  if (!include) {
    return null;
  }

  const params = url.split(/([?,=,&])/);
  const index = params.indexOf(param);
  const value = params[index + 2];

  return value.toString();
};

export const checkRequired = (data: any, errors: any) => {
  let isOK = true;
  let _errors: any = errors;
  for (const key in errors) {
    if (data[key] === '') {
      isOK = false;
      _errors[key] = {error: true, message: 'Required'};
    }
  }
  return {errors, isOK};
};
export const uniqeValueOfArray = (value: any, index: any, self: any) => {
  return self.indexOf(value) === index;
};
export const urlToBase64 = (url: string, callback: any) => {
  var xhr = new XMLHttpRequest();
  xhr.onload = function () {
    var reader = new FileReader();
    reader.readAsDataURL(xhr.response);
    reader.onloadend = function () {
      // console.log(reader.result);
      callback(reader.result);
    };
  };
  xhr.open('GET', url);
  xhr.responseType = 'blob';
  xhr.send();
};

export const getOrdinalSuffixOf = (numb: number) => {
  var j = numb % 10,
    k = numb % 100;
  if (j === 1 && k !== 11) {
    return numb + 'st';
  }
  if (j === 2 && k !== 12) {
    return numb + 'nd';
  }
  if (j === 3 && k !== 13) {
    return numb + 'rd';
  }
  return numb + 'th';
};

export const setDateFormat = (date: Date) => {
  return moment(date).format('YYYY-MM-DD') + '  00:00:00';
};
export const convertTimeToInt = (time: Date) => {
  let newTime = moment(time, 'DD-MMM-YYYY HH:mm:ss', true).format(
    'YYYY-MM-DD HH:mm:ss',
  );
  return parseInt(newTime.replace(/ |-|:/g, ''), 10);
};
export const uuid = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    var r = (Math.random() * 16) | 0,
      v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};
export const generateRandomColor = () => {
  const randomColor = Math.floor(Math.random() * 16777215)
    .toString(16)
    .padStart(6, '0');
  return `#${randomColor}`;
};
