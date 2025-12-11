<template>
  <s-card
    class="w-full bg-white suim_datalist"
    hide-footer
    no-gap
    v-if="Object.keys(data.matrixObj).length > 0"
  >
    <div class="mb-2 flex justify-end">
      <s-button
        icon="content-save"
        class="btn_primary"
        @click="preSave"
        :disabled="data.loading"
        v-if="profile.canCreate"
      />
    </div>
    <table
      class="w-full rounded-md relative mb-4 table-risk-matrix"
      v-if="!data.loading"
    >
      <thead>
        <tr
          class="border-[#D8D8D8] bg-[#F8F8F9] [&>*]:px-2 [&>*]:py-1 [&>*]:border-[#D8D8D8] [&>*]:border-[1px]"
        >
          <th
            class="text-center font-semibold"
            :colspan="data.severityList.length + 2"
          >
            {{ type == "RSCA" ? "IMPACT" : "SEVERITY" }}
          </th>
        </tr>
        <tr
          class="text-center border-[#D8D8D8] bg-[#F8F8F9] [&>*]:px-2 [&>*]:py-1 [&>*]:border-[#D8D8D8] [&>*]:border-[1px] text-center"
        >
          <th class="bg-[#F8F8F9]">&nbsp;</th>
          <th class="w-5">Level</th>
          <th v-for="(item, idx) in data.severityList" :key="idx">
            {{ type == "RSCA" ? item.ParameterName : item.Level }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(like, idx) in data.likelihoodList"
          :key="idx"
          class="border-[#D8D8D8] [&>*]:px-2 [&>*]:py-1 [&>*]:border-[#D8D8D8] [&>*]:border-[1px] text-center"
        >
          <td
            :rowspan="data.likelihoodList.length"
            v-if="idx == 0"
            class="text-center w-5 font-bold bg-[#F8F8F9] px-2"
          >
            <span class="vericaltext">Likelihood</span>
          </td>
          <td class="font-semibold bg-[#F8F8F9]">
            {{ type == "RSCA" ? like.ParameterName : like.Level }}
          </td>
          <td
            v-for="(se, index) in data.severityList"
            :key="index"
            :class="`bg-${data.matrixObj[bindMatrix(like._id, se._id)].RiskID}`"
          >
            <div class="flex justify-center items-center gap-2">
              <s-input
                ref="inputMatrix"
                class="rounded-none border-none mr-2 dd-matrix"
                :class="[type == 'RSCA' ? 'w-[90px]' : 'w-[60px]']"
                v-model="data.matrixObj[bindMatrix(like._id, se._id)].RiskID"
                :items="type == 'IBPR' ? IBPRCode : RSCACode"
                :useList="true"
              />
              <div class="text-center font-semibold" v-if="type == 'IBPR'">
                ({{ data.matrixObj[bindMatrix(like._id, se._id)].Value }})
              </div>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="flex justify-center">
      <loader />
    </div>
  </s-card>
  <div v-else class="flex justify-center mb-4 font-semibold">
    --No Data Matrix--
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, ref, nextTick } from "vue";
import { authStore } from "@/stores/auth.js";
import { SCard, SInput, util, SButton } from "suimjs";
import loader from "@/components/common/Loader.vue";

const axios = inject("axios");
const auth = authStore();

const props = defineProps({
  profile: { type: Object, default: () => {} },
  type: { type: String, default: () => "" },
});

const data = reactive({
  severityList: [],
  likelihoodList: [],
  matrixObj: {},
  objLikeliSeverity: {},
  loading: false,
});

const IBPRCode = ["C", "B", "A", "AA"];
const RSCACode = ["Low", "Medium", "High", "Extreme"];

async function getList(kind) {
  const url = `/bagong/${kind}/gets?Type=${props.type}`;
  const param = {
    CompanyID: auth.companyId,
  };
  await axios.post(url, param).then(
    (r) => {
      data[kind + "List"] = r.data.data.sort((a, b) => {
        return a.Level - b.Level;
      });

      for (let i in data[kind + "List"]) {
        let o = data[kind + "List"][i];
        data.objLikeliSeverity[o._id] = o;
      }

      if (kind == "likelihood") buildMatrix();
    },
    (e) => {
      util.showError(e);
    }
  );
}

async function getSources() {
  await getList("severity");
  await getList("likelihood");
}

function getMatrixList() {
  const url = `/bagong/riskmatrix/gets?Type=${props.type}`;
  const param = {
    CompanyID: auth.companyId,
  };
  axios.post(url, param).then(
    (r) => {
      // likeli -> severity
      data.matrixObj = {};
      for (let i in r.data.data) {
        let o = r.data.data[i];
        data.matrixObj[o.LikelihoodID + "#" + o.SeverityID] = o;
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}

function buildMatrix(flag) {
  const emptyObj = {
    _id: "",
    LikelihoodID: "",
    SeverityID: "",
    RiskID: "",
    Value: 0,
    CompanyID: auth.companyId,
    Type: props.type,
  };
  // likeli -> severity
  for (let i in data.likelihoodList) {
    let l = data.likelihoodList[i];
    for (let j in data.severityList) {
      let s = data.severityList[j];
      let ob = data.matrixObj[l._id + "#" + s._id];
      if (ob == undefined) {
        let val = JSON.parse(JSON.stringify(emptyObj));
        val.LikelihoodID = l._id;
        val.SeverityID = s._id;
        data.matrixObj[l._id + "#" + s._id] = val;
      }
      data.matrixObj[l._id + "#" + s._id].Value = l.Value * s.Value;
    }
  }
}

function bindMatrix(l, s) {
  return l + "#" + s;
}

async function preSave() {
  data.loading = true;
  let i = 1;
  for (let ky in data.matrixObj) {
    await onSave(data.matrixObj[ky], i == Object.keys(data.matrixObj).length);
    i++;
  }
}

async function onSave(param, isLast) {
  const url = "/bagong/riskmatrix/save";
  await axios.post(url, param).then(
    (r) => {
      if (isLast) {
        data.loading = false;
        util.showInfo("data has been saved");
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}

defineExpose({
  getMatrixList,
  getSources,
});

onMounted(() => {
  getMatrixList();
  getSources();
});
</script>
<style>
.vericaltext {
  writing-mode: vertical-lr;
  text-orientation: upright;
}

.dd-matrix .vs__actions {
  display: none;
}

.dd-matrix .vs__selected {
  @apply font-semibold;
}

.table-risk-matrix .bg-A,
.table-risk-matrix .bg-High {
  background: #ffc000;
}
.table-risk-matrix .bg-AA,
.table-risk-matrix .bg-Extreme {
  background: #ff0000;
}
.table-risk-matrix .bg-B,
.table-risk-matrix .bg-Medium {
  background: #ffff00;
}
.table-risk-matrix .bg-C,
.table-risk-matrix .bg-Low {
  background: #00b050;
}
</style>
