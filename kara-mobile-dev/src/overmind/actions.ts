/* eslint-disable @typescript-eslint/no-unused-vars */
import type {Context} from './index';
import {COMPANY_ID} from '@env';
export const setAppName = ({state}: any, payload: any) => {
  state.appName = payload;
};

export const signIn = async (
  {state, effects, actions}: Context,
  payload: {payload: any; auth: any},
) => {
  let {data} = await effects.api
    .signIn(payload.payload, payload.auth)
    .catch((e: any) => {
      console.log('[ERROR] loginUser : ', e);
      return Promise.reject(e);
    });
  if (data) {
    let res = await effects.api
      .changeCompanyID({
        Data: {CompanyID: COMPANY_ID},
        Scope: 'BOTH',
        Token: data.Token,
      })
      .catch((e: any) => {
        console.log('[ERROR] change data : ', e);
        return Promise.reject(e);
      });
    if (res.data) {
      state.token = res.data.Token;
      state.isSignedIn = true;
      state.userInfo = res.data.Data;
      actions.getDetailCompany(COMPANY_ID);
      return res.data;
    }
    // await actions.getUser({Email: data.Data.Email});
    return data;
  }
  return {};
};
export const getDetailCompany = async (
  {state, effects}: Context,
  payload: {payload: any},
) => {
  let {data} = await effects.api.getDetailCompany(payload).catch((e: any) => {
    console.log('[ERROR] getDetailCompany : ', e);
    return Promise.reject(e);
  });
  if (data) {
    state.company = data;
  }
  return data;
};
export const getRBAC = async ({state}: Context, featureID: string) => {
  const roleIDs = state.userInfo?.RBAC?.RoleIDs;
  if (roleIDs.indexOf('Administrators') >= 0) {
    return {
      canRead: true, // 1
      canCreate: true, // 2
      canUpdate: true, // 4
      canDelete: true, // 8
      canPosting: true, // 16
      canSpecial1: true, // 32
      canSpecial2: true, // 64
      canSpecial3: true, // 128
      Dimension: [],
    };
  }
  const r = {
    canRead: false, // 1
    canCreate: false, // 2
    canUpdate: false, // 4
    canDelete: false, // 8
    canPosting: false, // 16
    canSpecial1: false, // 32
    canSpecial2: false, // 64
    canSpecial3: false, // 128
  } as any;

  const access = state.userInfo?.RBAC?.Access ?? [];
  const idx = access.findIndex((e: any) => e.FeatureID === featureID);
  if (idx === -1) {
    return {...r, Dimension: []};
  }

  const myAcces = access[idx];

  let level = myAcces.Level;

  let i = 0;
  while (level > 0) {
    const bit = level % 2;
    r[Object.keys(r)[i]] = bit > 0;
    level = Math.floor(level / 2);
    i++;
  }

  return {...r, Dimension: myAcces.Dimension};
};
export const signOut = async ({state, effects}: Context) => {
  await effects.api.signOut();
  state.isSignedIn = false;
  state.isCheckIn = false;
  state.token = '';
  global.token = '';
  state.userInfo = {};
};

export const checkIn = async (
  {state, effects, actions}: Context,
  payload: {payload: any},
) => {
  let {data} = await effects.api.Post(payload.payload).catch((e: any) => {
    console.log('[ERROR] Post : ', e);
    return Promise.reject(e);
  });
  if (data) {
    actions.getStatus({});
  }
  return data;
};

export const checkOut = async (
  {effects, actions}: Context,
  payload: {payload: any},
) => {
  let {data} = await effects.api.Post(payload.payload).catch((e: any) => {
    console.log('[ERROR] Post : ', e);
    return Promise.reject(e);
  });
  if (data) {
    actions.getStatus({});
  }
  return data;
};
export const getStatus = async ({effects, state}: Context, payload: {}) => {
  let {data} = await effects.api.getStatus(payload).catch((e: any) => {
    console.log('[ERROR] getStatus : ', e);
    return Promise.reject(e);
  });
  if (data) {
    state.dataStatus = data;
    state.isCheckIn = data.CurrentOp === 'Checkin';
  }
  return data;
};
export const getMasterLocation = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getMasterLocation(payload).catch((e: any) => {
    console.log('[ERROR] getMasterLocation : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getNearestLocation = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getNearestLocation(payload).catch((e: any) => {
    console.log('[ERROR] getNearestLocation : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getMasterBus = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getMasterBus(payload).catch((e: any) => {
    console.log('[ERROR] getMasterBus : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getListHistory = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getListHistory(payload).catch((e: any) => {
    console.log('[ERROR] getListHistory : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getPhoto = async ({effects}: Context, payload: string) => {
  let {data} = await effects.api.getPhoto(payload).catch((e: any) => {
    console.log('[ERROR] getPhoto : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getAttendanceSummary = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api
    .getAttendanceSummary(payload)
    .catch((e: any) => {
      console.log('[ERROR] getAttendanceSummary : ', e);
      return Promise.reject(e);
    });
  return data;
};

//routine
export const getsRoutine = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getsRoutine(payload).catch((e: any) => {
    console.log('[ERROR] getsRoutine : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getsSite = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getsSite(payload).catch((e: any) => {
    console.log('[ERROR] getsSite : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getRoutine = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getRoutine(payload).catch((e: any) => {
    console.log('[ERROR] getRoutine : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const createNewRoutine = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.createNewRoutine(payload).catch((e: any) => {
    console.log('[ERROR] createNewRoutine : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const deleteRoutine = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.deleteRoutine(payload).catch((e: any) => {
    console.log('[ERROR] deleteRoutine : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getsRoutineDetails = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getsRoutineDetails(payload).catch((e: any) => {
    console.log('[ERROR] getsRoutineDetails : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getsRoutineChecklist = async (
  {effects, state}: Context,
  payload: {},
) => {
  let {data} = await effects.api
    .getsRoutineChecklist(payload)
    .catch((e: any) => {
      console.log('[ERROR] getsRoutineChecklist : ', e);
      return Promise.reject(e);
    });
  state.routineChecklist = data;
  return data;
};
export const changeDataRoutineChecklist = async (
  {state}: Context,
  payload: {},
) => {
  console.log(payload);
  state.routineChecklist = payload;
};

export const saveRoutineChecklist = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api
    .saveRoutineChecklist(payload)
    .catch((e: any) => {
      console.log('[ERROR] saveRoutineChecklist : ', e);
      return Promise.reject(e);
    });
  return data;
};
export const saveRoutineDetails = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.saveRoutineDetails(payload).catch((e: any) => {
    console.log('[ERROR] saveRoutineDetails : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const findEmployee = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.findEmployee(payload).catch((e: any) => {
    console.log('[ERROR] findEmployee : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const findDepartment = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.findDepartment(payload).catch((e: any) => {
    console.log('[ERROR] findDepartment : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const findMasterData = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.findMasterData(payload).catch((e: any) => {
    console.log('[ERROR] findMasterData : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const findSite = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.findSite(payload).catch((e: any) => {
    console.log('[ERROR] findSite : ', e);
    return Promise.reject(e);
  });
  return data;
};

export const routineRequest = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.routineRequest(payload).catch((e: any) => {
    console.log('[ERROR] routineRequest : ', e);
    return Promise.reject(e);
  });
  return data;
};

//approval
export const getsDataApproval = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getsDataApproval(payload).catch((e: any) => {
    console.log('[ERROR] getsDataApproval : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const getsDataApprovalJournal = async (
  {effects}: Context,
  payload: {},
) => {
  let {data} = await effects.api
    .getsDataApprovalJournal(payload)
    .catch((e: any) => {
      console.log('[ERROR] getsDataApprovalJournal : ', e);
      return Promise.reject(e);
    });
  return data;
};
export const postApprovalByGroup = async (
  {effects}: Context,
  payload: any[],
) => {
  let {data} = await effects.api
    .postApprovalByGroup(payload)
    .catch((e: any) => {
      console.log('[ERROR] postApprovalByGroup : ', e);
      return Promise.reject(e);
    });
  return data;
};
export const postApproval = async (
  {effects}: Context,
  payload: {
    payload: any[];
    type: String;
  },
) => {
  if (payload.type === 'fico') {
    let {data} = await effects.api
      .postApproval(payload.payload)
      .catch((e: any) => {
        console.log('[ERROR] postApproval : ', e);
        return Promise.reject(e);
      });
    return data;
  } else if (payload.type === 'scm') {
    let {data} = await effects.api
      .postApprovalSCM(payload.payload)
      .catch((e: any) => {
        console.log('[ERROR] postApproval scm : ', e);
        return Promise.reject(e);
      });
    return data;
  } else {
    let {data} = await effects.api
      .postApprovalMFG(payload.payload)
      .catch((e: any) => {
        console.log('[ERROR] postApproval mfg : ', e);
        return Promise.reject(e);
      });
    return data;
  }
};
export const findPreview = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.findPreview(payload).catch((e: any) => {
    console.log('[ERROR] findPreview : ', e);
    return Promise.reject(e);
  });
  return data;
};

// ASSET
export const getsAssetByJournal = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.getsAssetByJournal(payload).catch((e: any) => {
    console.log('[ERROR] getsAssetByJournal : ', e);
    return Promise.reject(e);
  });
  return data;
};
export const writeBatchWithContent = async (
  {effects}: Context,
  payload: any[],
) => {
  let {data} = await effects.api
    .writeBatchWithContent(payload)
    .catch((e: any) => {
      console.log('[ERROR] writeBatchWithContent : ', e);
      return Promise.reject(e);
    });
  return data;
};
export const deleteAsset = async ({effects}: Context, payload: string) => {
  let {data} = await effects.api.deleteAsset(payload).catch((e: any) => {
    console.log('[ERROR] deleteAsset : ', e);
    return Promise.reject(e);
  });
  return data;
};

export const deleteAttendance = async ({effects}: Context, payload: {}) => {
  let {data} = await effects.api.deleteAttendance(payload).catch((e: any) => {
    console.log('[ERROR] deleteAttendance : ', e);
    return Promise.reject(e);
  });
  return data;
};