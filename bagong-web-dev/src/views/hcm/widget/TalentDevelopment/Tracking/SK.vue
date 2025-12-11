<template>
  <div v-if="data.loading" class="h-[300px] flex items-center justify-center">
    <Loader kind="circle" />
  </div>
  <s-form
    v-else-if="data.appMode != 'preview'"
    v-model="data.record"
    :config="data.formCfg"
    :mode="readOnly ? 'view' : 'edit'"
    :hide-submit="readOnly"
    keep-label
    hide-cancel
    ref="formControl"
    @fieldChange="onFieldChange"
    @submitForm="save"
  >
    <template #buttons_1="{ item, config, inSubmission }">
      <div class="flex flex-row gap-2">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          moduleid="hcm"
          :auto-post="!waitTrxSubmit"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
        <template v-if="mode !== 'new'">
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </div>
    </template>
    <template #input_DirectorPosition="{ item }">
      <div class="mb-5">
        <s-input
          :key="item.DirectorID"
          label="Director postition"
          v-model="item.DirectorPosition"
          use-list
          :lookup-url="'/tenant/masterdata/find?MasterDataTypeID=PTE'"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          read-only
        ></s-input>
      </div>
    </template>
    <template #input_Notices="{ item }">
      <div class="mb-5">
        <label for="">Notices</label>
        <template v-for="(notice, index) in item.Notices" :key="index">
          <div class="flex flex-row gap-2 mb-2">
            <s-input
              class="w-full"
              v-model="item.Notices[index]"
              hide-label
              :read-only="readOnly"
            >
            </s-input>
            <s-button
              icon="plus"
              class="btn_primary new_btn"
              tooltip="add new"
              @click="item.Notices.push('')"
              :disabled="readOnly"
            />
            <s-button
              icon="delete"
              class="btn_secondary delete_btn"
              tooltip="delete"
              @click="item.Notices.splice(index, 1)"
              :disabled="readOnly || item.Notices.length === 1"
            >
            </s-button>
          </div>
        </template>
      </div>
    </template>
    <template #input_Decides="{ item }">
      <div class="mb-5">
        <label for="">Decides</label>
        <template v-for="(decide, index) in item.Decides" :key="index">
          <div class="flex flex-row gap-2 mb-2">
            <s-input
              class="w-full"
              v-model="item.Decides[index]"
              hide-label
              :read-only="readOnly"
            >
            </s-input>
            <s-button
              icon="plus"
              class="btn_primary new_btn"
              tooltip="add new"
              @click="item.Decides.push('')"
              :disabled="readOnly"
            />
            <s-button
              icon="delete"
              class="btn_secondary delete_btn"
              tooltip="delete"
              @click="item.Decides.splice(index, 1)"
              :disabled="readOnly || item.Decides.length === 1"
            >
            </s-button>
          </div>
        </template>
      </div>
    </template>
    <template #input_JournalTypeID="{ item }">
      <div class="mb-5">
        <s-input
          label="Journal type ID"
          v-model="item.JournalTypeID"
          :required="true"
          use-list
          :read-only="readOnly"
          :lookup-url="journalTypeLookupUrl"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="(_, v1) => getJurnalType(v1)"
        ></s-input>
      </div>
    </template>
    <template #input_Detail="{ item }">
      <div class="suim_area_table">
          <table class="w-full table-auto suim_table">
            <thead name="grid_header">
              <tr class="border-b-[1px] border-slate-500">
                <th class="text-left">Description</th>
                <th class="text-left">Existing</th>
                <th class="text-left">New Propose</th>
              </tr>
            </thead>
            <tbody name="grid_body">
              <template v-for="(value, key) in data.descBenefitDetail">
                <tr
                  class="cursor-pointer border-b-[1px] border-slate-200 last:border-non hover:bg-slate-200 even:bg-slate-100"
                >
                  <td class="py-2">
                    <div class="w-[300px]">
                      {{ key.split(/(?=[A-Z])/).join(" ") }}
                    </div>
                  </td>
                  <td class="py-2">
                    <div class="w-[300px]">
                      <s-input
                        hide-label
                        :label="key"
                        v-if="data.benefitDetail.Existing[key]"
                        v-model="data.benefitDetail.Existing[key]"
                        :kind="value.kind"
                        :use-list="value.useList"
                        :lookup-url="
                          value.useList ? value.lookupUrl : undefined
                        "
                        lookup-key="_id"
                        :lookup-labels="['Name']"
                        read-only
                      />
                      <div v-else>
                        <div v-if="value.kind != 'number'">-</div>
                        <div v-else>0</div>
                      </div>
                    </div>
                  </td>
                  <td class="py-2">
                    <div class="w-[300px]">
                      <s-input
                        hide-label
                        :label="key"
                        v-model="data.benefitDetail.NewPropose[key]"
                        :kind="value.kind"
                        :use-list="value.useList"
                        :lookup-url="
                          value.useList ? value.lookupUrl : undefined
                        "
                        lookup-key="_id"
                        :lookup-labels="['Name']"
                        read-only
                      />
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
    </template>
  </s-form>
  <PreviewReport
    v-else-if="data.appMode == 'preview'"
    class="card w-full"
    title="Preview"
    @close="closePreview"
    :disable-print="helper.isDisablePrintPreview(data.record.Status)"
    :SourceType="data.jType"
    :SourceJournalID="data.record._id"
    :VoucherNo="data.record.LedgerVoucherNo"
    reload="1"
  >
    <template #buttons="props">
      <div class="flex gap-[1px] mr-2">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="data.record.Status"
          :journal-id="data.record._id"
          :posting-profile-id="data.record.PostingProfileID"
          :journal-type-id="data.jType"
          moduleid="hcm"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
      </div>
    </template>
  </PreviewReport>
</template>
<script setup>
import { reactive, onMounted, inject, ref, computed, nextTick } from "vue";
import { SForm, loadFormConfig, util, SInput, SButton } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import helper from "@/scripts/helper.js";
import PreviewReport from "@/components/common/PreviewReport.vue";

const axios = inject("axios");
const formControl = ref(null);

const props = defineProps({
  talentDevelopmentID: { type: String, default: "" },
  employeeID: { type: String, default: "" },
  type: { type: String, default: "" },
  journalTypeLookupUrl: {
    type: String,
    default:
      "/hcm/journaltype/find?TransactionType=Talent%20Development%20-%20Promotion%20-%20Tracking%20SK%20Acting",
  },
  formConfig: {
    type: String,
    default: "hcm/talentdevelopmentsk/formconfig",
  },
});
const data = reactive({
  loading: false,
  appMode: "form",
  talentDevelopmentID: props.talentDevelopmentID,
  formCfg: {},
  record: {
    TalentDevelopmentID: props.talentDevelopmentID,
    Status: "DRAFT",
    Type: props.type,
    Notices: [""],
    Decides: [""],
  },
  journalType: {},
  jType: "SK",
  descBenefitDetail: {
    Department: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=DME",
      kind: "text",
    },
    Position: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PTE",
      kind: "text",
    },
    Grade: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GDE",
    },
    Group: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GME",
      kind: "text",
    },
    SubGroup: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=SGE",
      kind: "text",
    },
    Site: {
      useList: true,
      lookupUrl: "/bagong/sitesetup/find",
      kind: "text",
    },
    PointOfHire: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PME",
      kind: "text",
    },
    BasicSalary: {
      useList: false,
      lookupUrl: null,
      kind: "number",
    },
    Allowance: {
      useList: false,
      lookupUrl: null,
      kind: "text",
    },
  },
  benefitDetail: {
    Existing: {},
    NewPropose: {},
  },
});
function closePreview() {
  data.appMode = "form";
}
function preSubmit(record) {}
function save(record, cb) {
  if (record.Status === "") {
    record.Status = "DRAFT";
  }
  axios.post("/hcm/talentdevelopmentsk/save", record).then(
    (r) => {
      let _record = r.data;
      data.record = _record;
      cb();
    },
    (e) => {
      util.showError(e);
      setLoadingForm(false);
    }
  );
}
const onFieldChange = (name, v1, v2, old, record) => {
  if (name == "DirectorID" && v1) {
    getDirectorDetail(v1, data.record);
  }
};
const getEmpployeeDetail = async (record) => {
  try {
    const response = await axios.post("/bagong/employee/get", [
      props.employeeID,
    ]);
    record.Name = response.data.Name;
    record.NIK = response.data.Detail.IdentityCardNo;
    record.Group = response.data.Detail.IdentityCardNoGroup;
    record.Position = response.data.Detail.Position;
    record.Department = response.data.Detail.Department;
    record.Grade = response.data.Detail.Grade;
    record.Level = response.data.Detail.Level;
    record.IdentityCardNo = response.data.Detail.IdentityCardNo;
    record.PlaceOfBirth = response.data.Detail.PlaceOfBirth;
    record.DateOfBirth = response.data.Detail.DateOfBirth;
    record.Gender = response.data.Detail.Gender;
    record.Religion = response.data.Detail.Religion;
    record.Phone = response.data.Detail.Phone;
    record.Address = response.data.Detail.Address;
    record.POH = response.data.Detail.POH;
    record.JoinedDate = response.data.JoinDate;
  } catch (error) {
    util.showError(error);
  }
};
const getDirectorDetail = async (value, record) => {
  try {
    const response = await axios.post("/bagong/employee/get", [value]);
    record.DirectorPosition = response.data.Detail.Position;
    record.DirectorAddress = response.data.Detail.Address;
  } catch (error) {
    util.showError(error);
  }
};

const onLoadFormConfig = async (axiosInstance, formConfig) => {
  try {
    return await loadFormConfig(axiosInstance, formConfig);
  } catch (error) {
    util.showError(error);
    return null;
  }
};

const getsData = async () => {
  data.loading = true;
  try {
    const response = await axios.post(
      `/hcm/talentdevelopmentsk/find?TalentDevelopmentID=${data.talentDevelopmentID}&Type=${props.type}`,
      { Take: 1 }
    );

    if (response.data.length > 0) {
      data.record = response.data[0];
      if (data.record.Notices?.length === 0 || data.record.Notices === null) {
        data.record.Notices = [""];
      }
      if (
        (data.record.Decides?.length === 0) |
        (data.record.Decides === null)
      ) {
        data.record.Decides = [""];
      }
      await getEmpployeeDetail(data.record);
      if (data.record.DirectorID)
        await getDirectorDetail(data.record.DirectorID, data.record);

      util.nextTickN(2, async () => {
        data.formCfg = await onLoadFormConfig(axios, props.formConfig);
      });
    } else {
      await getEmpployeeDetail(data.record);
      util.nextTickN(2, async () => {
        data.formCfg = await onLoadFormConfig(axios, props.formConfig);
      });
    }
  } catch (error) {
    util.showError(error);
  } finally {
    data.loading = false;
  }
};
function getBenefitDetail(ID, EmployeeID) {
  data.benefitDetail = {
    Existing: {},
    NewPropose: {},
  };
  if (ID && EmployeeID) {
    const url = "/hcm/talentdevelopment/get-detail";
    axios.post(url, { ID, EmployeeID }).then(
      (r) => {
        data.benefitDetail = r.data;
      },
      (e) => util.showError(e)
    );
  }
}
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function setLoadingForm(loading) {
  formControl.value.setLoading(loading);
}

// function closePreview() {
//   data.appMode = "grid";
// }

function getJurnalType(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  formControl.value.setLoading(true);
  axios
    .post("/hcm/journaltype/get", [id])
    .then(
      (r) => {
        data.journalType = r.data;
        data.record.PostingProfileID = r.data.PostingProfileID;
        data.checklistTemplateID = r.data.ChecklistTemplate;
      },
      (e) => {
        data.journalType = {};
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      formControl.value.setLoading(false);
    });
}
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});

function setFormRequired(required) {
  formControl.value.getAllField().forEach((e) => {
    formControl.value.setFieldAttr(e.field, "required", required);
  });
}
function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    formControl.value.setCurrentTab(0);
    trxSubmit(doSubmit);
  }
}
function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = formControl.value.validate();
    if (valid) {
      setLoadingForm(true);
      save(data.record, () => {
        doSubmit();
      });
    }
    setFormRequired(false);
  });
}
function trxPostSubmit(record, action) {
  setLoadingForm(false);
  // closePreview();
  getsData();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
onMounted(() => {
  getsData();
  if (props.type === 'PERMANENT') {
    getBenefitDetail(props.talentDevelopmentID, props.employeeID);
  }
});
</script>
<style></style>
