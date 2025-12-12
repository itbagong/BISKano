import http from '@utils/http';
import base64 from 'react-native-base64';

export const api = {
  signIn(payload: any, auth: any) {
    const authHeader =
      'Basic ' + base64.encode(`${auth.username}:${auth.password}`);
    return http.post(
      '/iam/http-auth',
      {...payload},
      {
        headers: {Authorization: authHeader},
      },
    );
  },
  signOut() {
    return http.post('/iam/logout');
  },
  changeCompanyID(payload: any) {
    return http.post('/iam/change-data', payload);
  },
  getDetailCompany(payload: any) {
    return http.post('/tenant/company/get', [payload]);
  },
  //Attendance
  Post(payload: any) {
    return http.post('/kara/me/post', payload);
  },
  getStatus(payload: any) {
    return http.post('/kara/me/status', payload);
  },
  getMasterLocation(payload: any) {
    return http.post('/kara/worklocation/find', payload);
  },
  getNearestLocation(payload: any) {
    return http.post('/kara/admin/worklocation/find-nearest-location', payload);
  },
  getMasterBus(payload: any) {
    return http.post('/tenant/asset/find?GroupID=UNT', payload);
  },
  getListHistory(payload: any) {
    return http.post('/kara/me/history', payload);
  },
  getPhoto(payload: any) {
    return http.post('/kara/photo/get', [payload]);
  },
  getAttendanceSummary(payload: any) {
    return http.post('/kara/me/attendace-summary', payload);
  },

  //routine
  getsRoutine(payload: any) {
    return http.post('/mfg/routine/gets', payload);
  },
  getsSite(payload: any) {
    return http.post('/bagong/sitesetup/find', payload);
  },
  getRoutine(payload: any) {
    return http.post(`/mfg/routine/gets?_id=${payload._id}`, payload);
  },
  createNewRoutine(payload: any) {
    return http.post('/mfg/routine/add-new', payload);
  },
  deleteRoutine(payload: any) {
    return http.post('/mfg/routine/delete', payload);
  },
  getsRoutineDetails(payload: any) {
    return http.post(
      `/mfg/routine/detail/gets?RoutineID=${payload._id}`,
      payload,
    );
  },
  getsRoutineChecklist(payload: any) {
    return http.post('/mfg/routine/checklist/get-checklist', payload);
  },
  saveRoutineChecklist(payload: any) {
    return http.post('/mfg/routine/checklist/save-checklist', payload);
  },
  saveRoutineDetails(payload: any) {
    return http.post('/mfg/routine/detail/save', payload);
  },
  findEmployee(payload: any) {
    return http.post('/tenant/employee/find', payload);
  },
  findDepartment(payload: any) {
    return http.post('tenant/dimension/find?DimensionType=CC', payload);
  },
  findMasterData(payload: any) {
    return http.post(
      `/tenant/masterdata/find?MasterDataTypeID=${payload.MasterDataTypeID}`,
      payload,
    );
  },
  findSite(payload: any) {
    return http.post('/tenant/dimension/find?DimensionType=Site', payload);
  },
  routineRequest(payload: any) {
    return http.post('/mfg/routine/create-for-wo', payload);
  },

  //Approval
  getsDataApproval(payload: any) {
    return http.post('/fico/approvalaggregator/group-by', payload);
  },
  getsDataApprovalJournal(payload: any) {
    return http.post('/fico/approvalaggregator/get-journal', payload);
  },
  postApprovalByGroup(payload: any) {
    return http.post('/mfg/approvalaggregator/post-by-group', payload);
  },
  postApproval(payload: any) {
    return http.post('/fico/approvalaggregator/post', payload);
  },
  postApprovalSCM(payload: any) {
    return http.post('/scm/postingprofile/post', payload);
  },
  postApprovalMFG(payload: any) {
    return http.post('/mfg/postingprofile/post', payload);
  },
  findPreview(payload: any) {
    // return http.get('/tenant/preview/find?', {
    //   params: payload,
    // });
    let url = '';
    switch (payload.SourceType) {
      case "Inventory Receive":
      case "Inventory Issuance":
      case "Movement In":
      case "Movement Out":
      case "Transfer":
      case "Asset Acquisition":
      case "Item Request":
      case "Purchase Request":
      case "INVENTORY":
      case "Purchase Order":
        url = `/scm/preview/get?reload=1&type=${payload.SourceType}&id=${payload.SourceJournalID}&name=${payload.Name}&voucher=${payload.VoucherNo}`;
        break;
      case "Work Request":
      case "Work Order":
      case "Work Order Report Consumption":
      case "Work Order Report Resource":
      case "Work Order Report Output":
        url = `/mfg/preview/get?reload=1&type=${payload.SourceType}&id=${payload.SourceJournalID}&name=${payload.Name}&voucher=${payload.VoucherNo}`;
        break;
      default:
        url = `/fico/preview/get?reload=1&type=${payload.SourceType}&id=${payload.SourceJournalID}&name=${payload.Name}&voucher=${payload.VoucherNo}`;
    }
    return http.get(url);
  },

  //ASSET
  getsAssetByJournal(payload: any) {
    return http.post('/asset/read-by-journal', payload);
  },
  writeBatchWithContent(payload: any) {
    return http.post('/asset/write-batch-with-content', payload);
  },
  deleteAsset(payload: any) {
    return http.post('/asset/delete', payload);
  },
  deleteAttendance(payload: any) {
    return http.post('/kara/admin/trx/delete', payload);
  },
};
