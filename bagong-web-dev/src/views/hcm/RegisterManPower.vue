<template>
<div class="w-[400px] mt-4">
    <s-form   ref="regForm"
        class="card regMpForm"
        v-model="data.record" :config="data.config" keep-label
        hide-cancel submit-text="Register"  
        buttons-on-bottom :buttons-on-top="false"
        @submit-form="register"
    > 
    </s-form>
</div>
</template>
<script setup>

import { onMounted, reactive, inject, ref, watch } from "vue"
import {
  formInput,
  createFormConfig, 
  SForm, 
  util
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";


layoutStore().name = "tenant";

const axios = inject('axios');

const data = reactive({ 
    record: {},
    config: {},
});
function register(_, cbOk, cbFalse){
   axios.post("/hcm/tracking/apply-applicant",data.record).then(r => {
        data.mode = 'info';
        cbOk();
    }, e => {
        cbFalse();
        util.showError(e);
    })
}
onMounted(() => {
  const cfg = createFormConfig("Register Man Power", true);
  cfg.setting.showTitle = false;
  const emp_input = new formInput();
  emp_input.field = "EmployeeID";
  emp_input.label = "Employee";
  emp_input.kind = "string";
  emp_input.useList = true
  emp_input.multiple = false
  emp_input.required = true;
	emp_input.lookupUrl ="/tenant/employee/find"
  emp_input.lookupKey ="_id"
  emp_input.lookupLabels =['Name']
  emp_input.lookupSearchs=['_id', 'Name']

  const job_input = new formInput();
  job_input.field = "JobID";
  job_input.label = "Job Vacancy";
  job_input.kind = "string";
  job_input.useList = true
  job_input.multiple = true
  job_input.required = true;	
  job_input.lookupUrl ="/hcm/manpowerrequest/find"
  job_input.lookupKey ="_id"
  job_input.lookupLabels =['Name']
  job_input.lookupSearchs=['_id', 'Name']

  cfg
    .addSection("General", false)
    .addRowAuto(1, emp_input, job_input);
  data.config = cfg.generateConfig();
});
</script>
<style>
.regMpForm.loading .loader{
  width: 500px;
}
</style>