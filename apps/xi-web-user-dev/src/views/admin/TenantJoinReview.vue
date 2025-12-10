<template>
  <div class="w-full">
    <div class="card flex flex-col gap-2" v-if="data.records.length > 0">
      <h1 class="text-primary">Pending Join Request</h1>
      <div class="flex flex-col gap-2">
        <div v-for="record in data.records" class="flex gap-2 even:bg-slate-100 py-1">
          <div class="flex flex-col gap-1 grow">
            {{ record.DisplayName }}
          </div>
          <div class="flex flex-col gap-1 grow">
            <s-button icon='key-change' label='Approve' @click="process('approve',record._id)" class='btn_primary'/>
            <s-button icon='key-change' label='Reject' @click="process('reject',record._id)" class='btn_primary'/>
          </div>
          <div>
            {{ moment(record.Created).fromNow() }}
          </div>
                </div>
            </div>
      </div>
  </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout';
import { SForm, SList, SButton, createFormConfig, formInput, util } from 'suimjs';
import { inject } from 'vue';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import moment from 'moment';

layoutStore().change("tenant");

const router = useRouter();
const joinForm = ref(null);
const axios = inject('axios');
const data = reactive({
    mode: 'register',
    record: {},
    config: {},
    records: []
});

onMounted(() => {
    const cfg = createFormConfig("Reset Password",false);
    const tenantID_input = new(formInput);
    tenantID_input.field = "TenantID";
    tenantID_input.label = "Tenant ID or Code";
    tenantID_input.kind = "string";
    tenantID_input.required = true;

    cfg.addSection("General",false).addRowAuto(1, tenantID_input);
    data.config = cfg.generateConfig();

    loadTenantJoinRequest();
});
function process(status,recordid){
  //console.log(status,recordid)
  var payload = {"RequestID":recordid}
  if (status=="approve"){
    payload.Approve = true
  }else{
    payload.Approve = false
  }
  axios.post("/iam/tenantjoin/approve",payload).then(r=>{
    //alert("Done")
    loadTenantJoinRequest();
  },e=>util.showError(e))
}

function loadTenantJoinRequest() {
    axios.post('/iam/tenantjoin/review?Status=PENDING',{'sort':['-_id']}).then(r => {
        data.records = r.data;
    }, e => util.showError(e))
}

</script>