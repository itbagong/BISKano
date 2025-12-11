<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/fico/loansetup/gridconfig"
      form-config="/fico/loansetup/formconfig"
      grid-read="/fico/loansetup/gets"
      form-read="/fico/loansetup/get"
      grid-mode="grid"
      grid-delete="/fico/loansetup/delete"
      form-keep-label
      form-insert="/fico/loansetup/save"
      form-update="/fico/loansetup/save"
      :grid-fields="['Status']"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterFormConfig="alterFormConfig"
      @alterGridConfig="alterGridConfig"
    >
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, computed } from "vue";
import { DataList, util} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";
import moment from "moment";

layoutStore().name = "tenant";
const featureID = "LoanMaster";

const profile = authStore().getRBAC(featureID);
const auth = authStore();
const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  isPreview: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Loan",
  record: {
    _id: "",
    RequestDate: new Date(),
    Dimension: [],
    Status: "",
  },
  jType: "LOAN",
  journalType: {},
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Master Loan`;
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Master Loan | ${record._id}`;
  data.record = record;
  getDetailEmployee(data.record.EmployeeID, data.record);

  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);


    if(route.query.objname == "/fico/taxsetup"){
      listControl.value.setFormFieldAttr("Modules", "rules", [
        (v) => { 
         
        if (!Array.isArray(v) || v.length == 0)
          return "Required";
        return "";
        },
     ]);
    }
  });
}
function getDetailEmployee(id, record) {
  if (!id) {
    return 
  }
  axios.post("/tenant/employee/get", [id]).then(
    (r) => {
      util.nextTickN(2, () => {
        record.EmployeeName = r.data.Name;
        record.WorkLocation = r.data.Dimension.find(o => o.Key === 'Site').Value;
        record.Dimension = r.data.Dimension;

        const url = "/bagong/employeedetail/find?EmployeeID=" + id;
        axios.post(url).then(
          (rr) => {
            util.nextTickN(2, () => {
              if (rr.data.length > 0) {
                record.Department = rr.data[0].Department;
                record.Position = rr.data[0].Position;
                record.EmployeeStatus = rr.data[0].EmployeeStatus;
                record.NIK = rr.data[0].EmployeeNo;
                record.MobilePhoneNumber = rr.data[0].Phone;
                record.PeriodOfEmployement = getDetailedFromNow(
                  r.data.JoinDate
                );
                record.Salary = util.formatMoney(rr.data[0].BasicSalary);
                console.log(
                  moment(r.data.JoinDate).fromNow(),
                  rr.data[0].BasicSalary
                );
              }
            });
          },
          (e) => {
            util.showError(e);
          }
        );
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
function onFormFieldChange(name, v1, v2, old, record) {
  if (name === "EmployeeID") {
    getDetailEmployee(v1, record);
  }
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Master Loan";
  }
}
function alterFormConfig(cfg) {}
function alterGridConfig(cfg) {
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({ field: "EmployeeNIK", label: "NIK" })
  );
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({ field: "EmployeeName", label: "Employee name" })
  );
  return cfg;
}
</script>
