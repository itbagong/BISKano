<template>
  <div
    class="overflow-x-scroll relative h-[650px]"
    v-if="data.record.length > 0"
  >
    <s-grid
      :class="`trx_RSCA_lines`"
      ref="listControl"
      :config="data.cfg"
      hide-search
      hide-sort
      hide-refresh-button
      hide-select
      hide-footer
      hide-new-button
      hide-action
      editor
      auto-commit-line
      no-confirm-delete
    >
      <template #item_Division="{ item }">
        <s-input
          v-model="item.Division"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRDampak`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_Department="{ item }">
        <s-input
          v-model="item.Department"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRDampak`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_CriticalActivity="{ item }">
        <s-input
          v-model="item.CriticalActivity"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`she/legalregister/find`"
          :lookup-labels="['LegalNo']"
          :lookup-searchs="['LegalNo']"
        />
      </template>
      <template #item_AcceptRisk="{ item }">
        <div class="flex justify-center">
          <s-toggle
            v-model="item.AcceptRisk"
            class="w-[140px] mt-0.5"
            yes-label="Acceptable"
            no-label="Unacceptable"
            yes-class-style="bg-success"
            no-class-style="bg-secondary"
          />
        </div>
      </template>
      <template #item_InherentRiskLevel="{ item }">
        <div :class="[`matrix_div_inherit_${item.ID}`]">
          {{
            calMatrix(
              item,
              item.InherentImpact,
              item.InherentLikelihood,
              `matrix_div_inherit_${item.ID}`,
              "InherentRiskLevel"
            )
          }}
          <div
            class="flex gap-4 font-semibold h-full p-4 h-full justify-center"
            v-if="item.InherentImpact && item.InherentLikelihood"
          >
            <div>{{ riskMatrix[item.InherentRiskLevel].RiskID }}</div>
          </div>
        </div>
      </template>
      <template #item_ResidualRiskLevel="{ item }">
        <div :class="[`matrix_div_residual_${item.ID}`]">
          {{
            calMatrix(
              item,
              item.ResidualImpact,
              item.ResidualLikelihood,
              `matrix_div_residual_${item.ID}`,
              "ResidualRiskLevel"
            )
          }}
          <div
            class="flex gap-4 font-semibold h-full p-4 h-full justify-center"
            v-if="item.ResidualImpact && item.ResidualLikelihood"
          >
            <div>{{ riskMatrix[item.ResidualRiskLevel].RiskID }}</div>
          </div>
        </div>
      </template>
      <template #item_ExpectedRiskLevel="{ item }">
        <div :class="[`matrix_div_expected_${item.ID}`]">
          {{
            calMatrix(
              item,
              item.ExpectedImpact,
              item.ExpectedLikelihood,
              `matrix_div_expected_${item.ID}`,
              "ExpectedRiskLevel"
            )
          }}
          <div
            class="flex gap-4 font-semibold h-full p-4 h-full justify-center"
            v-if="item.ExpectedImpact && item.ExpectedLikelihood"
          >
            <div>{{ riskMatrix[item.ExpectedRiskLevel].RiskID }}</div>
          </div>
        </div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";

import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  SModal,
  loadGridConfig,
} from "suimjs";
import helper from "@/scripts/helper.js";
import SToggle from "@/components/common/SButtonToggle.vue";

const listControl = ref(null);
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  templateId: { type: String, default: () => "" },
  riskMatrix: { type: Object, default: {} },
});

const data = reactive({
  appMode: "grid",
  record: props.modelValue ?? [],
  cfg: {},
});

const emit = defineEmits({
  "update:modelValue": null,
});

function newData(r) {
  data.record = r;
}

function openForm(r) {
  data.record = r;
}

function getTemplateID(id) {
  if (!id) {
    data.record = [];
    return;
  }
  const url = `/she/masterrsca/find?_id=${id}`;
  axios.post(url).then(
    (r) => {
      buildLines(r.data[0].Lines);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function refreshGrid() {
  util.nextTickN(3, () => {
    listControl.value.setRecords(data.record);
    util.nextTickN(3, () => {
      formatRowSpan("trx_RSCA_lines");
    });
  });
}

function buildLines(lines) {
  if (data.record.length == 0) {
    data.record = lines;
  }
  refreshGrid();
}

function formatRowSpan(selector) {
  const myTable = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );
  for (let i = 0, row; (row = myTable[0].rows[i]); i++) {
    const firstCell = row.cells[0];
    const secondCell = row.cells[1];
    firstCell.classList.remove("hidden");
    secondCell.classList.remove("hidden");
    if (data.record[i].ParentId == "") {
      let lengthRow = data.record.filter(
        (o) => o.ParentId == data.record[i].ID
      );
      firstCell.rowSpan = lengthRow.length + 1;
      secondCell.rowSpan = lengthRow.length + 1;
    } else {
      firstCell.classList.add("hidden");
      secondCell.classList.add("hidden");
    }
  }
  generateThead();
}

function generateThead() {
  const dtThead = [
    { name: "Category", rowSpan: 2 },
    { name: "Risk No", rowSpan: 2 },
    { name: "Bidang", rowSpan: 2 },
    { name: "Unit Kerja", rowSpan: 2 },
    { name: "Critical Activity/Process/Issues", rowSpan: 2 },
    { name: "Risk Type", rowSpan: 2 },
    {
      name: "Need to be addresed",
      rowSpan: 0,
      colSpan: 2,
      child: ["Risk", "Opportunities"],
    },
    { name: "Cause", rowSpan: 2 },
    { name: "Risk Level", rowSpan: 2 },
    {
      name: "Inherent Risk Level",
      rowSpan: 0,
      colSpan: 3,
      child: ["Impact", "Likelihood", "Risk Level"],
    },
    { name: "Existing Control", rowSpan: 2 },
    {
      name: "Residual Risk Level",
      rowSpan: 0,
      colSpan: 3,
      child: ["Impact", "Likelihood", "Risk Level"],
    },
    { name: "Accept Risk ?", rowSpan: 2 },
    { name: "Treatmen Plan (WAJIB dibuat jika NOT ACCEPTED)", rowSpan: 2 },
    {
      name: "Expected Risk Level",
      rowSpan: 0,
      colSpan: 3,
      child: ["Impact", "Likelihood", "Risk Level"],
    },
    { name: "PIC", rowSpan: 2 },
    { name: "Due Date", rowSpan: 2 },
    { name: "Progress", rowSpan: 2 },
  ];
  const el = document.querySelector(".trx_RSCA_lines table thead");
  let row = el.insertRow(0);
  let row2 = el.insertRow(1);
  for (let i in dtThead) {
    let c = dtThead[i];
    row.innerHTML += `<td ${c.rowSpan > 0 ? "rowspan=" + c.rowSpan : ""}
      ${
        c.colSpan > 0 ? "colspan=" + c.colSpan : ""
      } class='text-center border-r font-semibold'>${c.name}</td>`;
    if (c.child) {
      for (let j in c.child) {
        row2.innerHTML += `<td class='text-center border-r font-semibold'>${c.child[j]}</td>`;
      }
    }
  }
  el.lastElementChild.remove();
}

function calMatrix(dt, saverity, likelihood, selector, field) {
  if (!saverity || !likelihood) {
    setBgMatrix(dt.ID, null, selector);
    return;
  }
  if (props.riskMatrix[`${saverity}#${likelihood}`]) {
    let o = props.riskMatrix[`${saverity}#${likelihood}`];
    dt[field] = o._id;
    setBgMatrix(dt.ID, o.RiskID, selector);
  }
}

function setBgMatrix(id, riskID, selector) {
  util.nextTickN(3, () => {
    let child = document.querySelector(`.${selector}`);
    let parent = child.closest("td");
    parent.classList.remove(
      "BgMatrix-Low",
      "BgMatrix-Medium",
      "BgMatrix-High",
      "BgMatrix-Extreme"
    );
    if (riskID) {
      parent.classList.add(`BgMatrix-${riskID}`);
    }
  });
}

watch(
  () => props.templateId,
  (nv) => {
    getTemplateID(nv);
  }
);

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  }
);

onMounted(() => {
  loadGridConfig(axios, "/she/rscatrx/lines/gridconfig").then(
    (r) => {
      data.cfg = r;
    },
    (e) => {}
  );
  getTemplateID(props.templateId);
});
</script>
<style>
.trx_RSCA_lines tr td {
  min-width: 150px;
}

.trx_RSCA_lines tr th {
  @apply sticky top-0 bg-gray-100;
}

.trx_RSCA_lines .BgMatrix-Low {
  background: #00b050;
}
.trx_RSCA_lines .BgMatrix-Medium {
  background: #ffff00;
}
.trx_RSCA_lines .BgMatrix-High {
  background: #ffc000;
}
.trx_RSCA_lines .BgMatrix-Extreme {
  background: #ff0000;
}
</style>
