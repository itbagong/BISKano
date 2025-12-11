<template>
  <s-grid
    class="grid_Audit_SMKP"
    ref="listControl"
    :config="data.cfg"
    hide-search
    hide-sort
    hide-refresh-button
    hide-select
    hide-new-button
    hide-action
    hide-control
    auto-commit-line
    no-confirm-delete
  >
    <template #item_ElementValue="{ item }">
      <div
        class="w-full td-SMKP-elementValue"
        v-if="item.TemplateLine.Level == 0"
      >
        <s-input
          v-model="item.ElementValue"
          kind="number"
          @change="onChangeElemenValue(item)"
        />
      </div>
      <div v-else>
        <div class="text-right">
          {{ formatNumDecimal(item.ElementValue) }}
        </div>
      </div>
    </template>
    <template #item_ScoreAudit="{ item }">
      <s-input
        v-model="item.ScoreAudit"
        use-list
        :items="[
          { key: 0, text: '0' },
          { key: 1, text: '1' },
          { key: 2, text: '2' },
        ]"
        v-if="item.TemplateLine.Level == 2"
        @change="onChangeAuditScore(item)"
        :showClearButton="false"
      />
    </template>
    <template #item_AuditElementValue="{ item }">
      <div class="text-right">
        {{ formatNumDecimal(item.AuditElementValue) }}
      </div>
    </template>
    <template #item_Note="{ item }">
      <s-input
        v-model="item.Note"
        kind=""
        multi-row="3"
        v-if="item.TemplateLine.Level == 2"
      />
    </template>
    <template #paging="{ item }">
      <div class="grid grid-cols-3 font-semibold text-center border smkp-total">
        <div class="col-span-2 border-b p-1">TOTAL NILAI MAKSIMAL</div>
        <div class="border-b p-1">{{ totalMaxValue }}</div>
        <div class="col-span-2 border-b p-1">TOTAL NILAI AKTUAL</div>
        <div class="border-b p-1">{{ totalActualValue }}</div>
        <div class="col-span-2 border-b p-1">ACHIEVEMENT AUDIT SISTEM SMKP</div>
        <div class="border-b p-1">
          {{ formatNumDecimal(totalPencapainan) }} %
        </div>
      </div>
    </template>
  </s-grid>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import {
  util,
  SInput,
  SButton,
  SGrid,
  SModal,
  SCard,
  loadGridConfig,
} from "suimjs";
import helper from "@/scripts/helper.js";
import Pica from "@/components/common/ItemPica.vue";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  templateId: { type: String, default: "" },
  formatNumDecimal: { type: Function },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const data = reactive({
  cfg: {},
  record: props.modelValue ?? [],
});

function getDataByTemplate(id) {
  if (!id) {
    data.record = [];
    refreshGridLines();
    return;
  }

  if (data.record.length > 0) {
    refreshGridLines();
    return;
  }

  const url = "/she/mcuitemtemplate/get";
  axios.post(url, [id]).then(
    (r) => {
      mapFieldLine(r.data.Lines);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function mapFieldLine(dt) {
  let tmp = [];
  for (let i in dt) {
    let o = dt[i];
    let x = {
      No: o.Number,
      ElementAudit: o.Description,
      ElementValue: 0,
      MaxElementValue:
        o.Level == 2
          ? o.AnswerValue
          : sumOfField(dt, o.Level, "Level", "AnswerValue"),
      ScoreAudit: 0,
      AuditElementValue: 0,
      Note: "",
      TemplateLine: o,
    };
    tmp.push(x);
  }
  data.record = tmp;
  refreshGridLines();
}

function sumOfField(dt, level, fieldFilter, fieldsum) {
  let parent = dt.find((o) => o[fieldFilter] == level);
  let child = dt.filter((o) => o.Parent.includes(parent.ID));
  return child.map((o) => o[fieldsum]).reduce((a, b) => a + b, 0);
}

function onChangeAuditScore(item) {
  util.nextTickN(2, () => {
    item.ScoreAudit = parseInt(item.ScoreAudit);

    let splitParent = item.TemplateLine.Parent.split("#");
    let lvl1 = data.record.find((o) => o.TemplateLine.ID == splitParent[1]);
    let lvl0 = data.record.find((o) => o.TemplateLine.ID == splitParent[0]);
    if (lvl0 && lvl1) {
      item.AuditElementValue =
        (item.ScoreAudit / item.MaxElementValue) * item.ElementValue;
    } else {
      item.AuditElementValue = 0;
    }

    for (let i = splitParent.length; i > 0; i--) {
      setParentValue(splitParent[i - 1], i);
    }
  });
}

function onChangeElemenValue(item) {
  util.nextTickN(2, () => {
    item.ElementValue = parseInt(item.ElementValue);
    let child = data.record.filter((o) =>
      o.TemplateLine.Parent.includes(item.TemplateLine.ID)
    );

    for (let i in child) {
      let o = child[i];
      o.ElementValue =
        (o.MaxElementValue / item.MaxElementValue) * item.ElementValue;
    }
  });
}

function formatGridLayout(selector) {
  if (data.record.length == 0 || !props.templateId) return;

  const thead = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_header']`
  );

  thead[0].rows[0].innerHTML = `<tr>
    <th rowspan='2' class='w-[50px]'>No</th>
    <th rowspan='2'>Kriteria</th>
    <th rowspan='2' class='w-[100px]'>Nilai Elemen</th>
    <th rowspan='2' class='w-[100px]'>Nilai Maximum Elemen</th>
    <th colspan='2' >Nilai Audit</th>
    <th rowspan='2' class='w-[350px]'>Catatan</th>
  </tr>`;

  thead[0].insertRow(1);
  thead[0].rows[1].innerHTML = `<tr>
    <th class='w-[100px]' style="border-left:1px solid #D1D5DC">Bobot</th>
    <th class='w-[100px]'>Score</th>
  </tr>`;

  const tbody = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );
  for (var i = 0, row; (row = tbody[0].rows[i]); i++) {
    row.classList.add("tr-smkp-" + data.record[i].TemplateLine.Level);

    for (var j = 0, col; (col = row.cells[j]); j++) {
      if (j == 0) {
        col.classList.add("hidden");
      }
      if (j == 1) {
        col.innerHTML = `<td class="">
        <div class="flex gap-0.5">
          <div class="px-${data.record[i].TemplateLine.Level * 3}"></div>
          <div>
          ${data.record[i].No} ${data.record[i].ElementAudit}
          </div>
        </div>
        </td>`;

        col.colSpan = 2;
      }
    }
  }
}

function refreshGridLines() {
  util.nextTickN(4, () => {
    listControl.value.setRecords(data.record);
    if (data.record.length == 0) return;
    setTimeout(() => {
      formatGridLayout("grid_Audit_SMKP");
    }, 500);
  });
}

function onChangePoint(item) {
  util.nextTickN(2, () => {
    item.Point = parseInt(item.Point);
    let bobot = item.TemplateLine.AnswerValue;

    if (item.Point == 0 || isNaN(item.Point)) {
      item.Score = 0;
      item.ACH = 0;
    } else {
      item.Score = bobot;
      setAch(item.TemplateLine.Parent.split("#")[1], item.Score);
    }

    let splitParent = item.TemplateLine.Parent.split("#");
    for (let i = splitParent.length; i > 0; i--) {
      setParentValue(splitParent[i - 1], i);
    }
  });
}

function setParentValue(id, level) {
  let idx = data.record.findIndex((o) => o.TemplateLine.ID == id);
  let parent = data.record[idx];
  let child = data.record.filter(
    (o) => o.TemplateLine.Parent.includes(id) && o.TemplateLine.Level == level
  );

  let res = {
    ScoreAudit: 0,
    AuditElementValue: 0,
  };

  for (let ky in res) {
    res[ky] = child.map((o) => o[ky]).reduce((a, b) => a + b, 0);
    if (parent.TemplateLine.Level == 0) {
      parent[ky] = res[ky] / child.length;
    } else {
      parent[ky] = res[ky];
    }
  }
}

const totalMaxValue = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.MaxElementValue)
      .reduce((a, b) => a + b, 0);
  },
});

const totalActualValue = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.ScoreAudit)
      .reduce((a, b) => a + b, 0);
  },
});

const totalPencapainan = computed({
  get() {
    let res = (totalActualValue.value / totalMaxValue.value) * 100;
    return isNaN(res) ? 0 : res;
  },
});

watch(
  () => props.templateId,
  (nv) => {
    getDataByTemplate(nv);
  }
);

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

onMounted(() => {
  loadGridConfig(axios, "/she/audit/smkp/gridconfig").then(
    (r) => {
      data.cfg = r;
    },
    (e) => util.showError(e)
  );
  getDataByTemplate(props.templateId);
});
</script>

<style>
.grid_Audit_SMKP .tr-smkp-0 td:not(:has(.td-SMKP-elementValue)),
.grid_Audit_SMKP .smkp-total div {
  background: #bfbfbf;
}

.grid_Audit_SMKP .tr-smkp-1 td {
  background: #ddebf7;
}

.grid_Audit_SMKP .tr-smkp-0,
.grid_Audit_SMKP .tr-smkp-1 {
  @apply font-semibold;
}
</style>
