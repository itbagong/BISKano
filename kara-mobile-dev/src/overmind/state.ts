type State = {
  appName: string;
  isSignedIn: boolean;
  isCheckIn: boolean;
  dataStatus: any;
  userInfo: any;
  token: string;
  routineChecklist: any;
  company: any;
  notifications: any[];
};
export const state: State = {
  appName: 'karamobile',
  isSignedIn: false,
  isCheckIn: false,
  dataStatus: {},
  userInfo: {},
  token: '',
  routineChecklist: {},
  company: {},
  notifications: [],
};
