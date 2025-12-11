<template>
  <s-grid
    class="grid_Audit_SMKPAU"
    ref="listControl"
    :config="data.cfg"
    hide-search
    hide-sort
    hide-refresh-button
    hide-select
    hide-new-button
    hide-action
    hide-control
    hide-footer
    auto-commit-line
    no-confirm-delete
    total-url="manual"
  >
    <template #item_Point="{ item }">
      <s-toggle
        v-model="item.Point"
        class="w-[100px] mt-0.5"
        v-if="item.TemplateLine.Level == 1"
        @change="onChangeActualValue(item)"
      />
      <div v-else></div>
    </template>
    <template #item_ACH="{ item }">
      <div class="text-right py-1.5">
        {{ formatNumDecimal(item.ACH) }}
        %
      </div>
    </template>
    <template #item_Note="{ item }">
      <s-input
        v-model="item.Note"
        multi-row="3"
        v-if="item.TemplateLine.Level == 1"
      />
      <div v-else></div>
    </template>
    <template #item_ActualValue="{ item }">
      <s-input
        v-model="item.ActualValue"
        kind="number"
        v-if="item.TemplateLine.Level == 1"
        @change="onChangeActualValue(item)"
      />
      <div v-else class="text-right">{{ item.ActualValue }}</div>
    </template>
    <template #item_SupportingDocument="{ item }">
      <div class="grid justify-center" v-if="item.TemplateLine.Level == 1">
        <uploader
          :journalId="jurnalId"
          :journalType="data.attchKind"
          :config="{ label: '' }"
          is-single-upload
          single-save
          :tags="[`${data.attchKind}_No_${item.TemplateLine.ID}`]"
          :key="1"
          bytag
        />
      </div>
    </template>
    <template #item_AttachedDocument="{ item }">
      <s-input
        v-model="item.AttachedDocument"
        multi-row="3"
        v-if="item.TemplateLine.Level == 1"
      />
    </template>

    <template #grid_total="{ item }">
      <tr class="font-semibold tr-smkpau-0">
        <td colspan="5" class="p-2 text-center">
          Hasil Akhir Implementasi Sistem Manajemen Keselamatan Perusahaan
          Angkutan Umum (SMKPAU)
        </td>
        <td class="text-right">{{ totalTarget }}</td>
        <td class="text-right">{{ totalActual }}</td>
        <td class="text-right">{{ formatNumDecimal(totalAch) }} %</td>
        <td class="text-right"></td>
      </tr>
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
import SToggle from "@/components/common/SButtonToggle.vue";
import Uploader from "@/components/common/Uploader.vue";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  templateId: { type: String, default: "" },
  formatNumDecimal: { type: Function },
  jurnalId: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const data = reactive({
  cfg: {},
  record: props.modelValue ?? [],
  attchKind: "AUDIT_LINES_SMKPOU",
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
      SupportingDocument: "",
      AttachedDocument: "",
      Point: true,
      Target: o.AnswerValue,
      ActualValue: 0,
      ACH: 0,
      Note: "",
      TemplateLine: o,
    };
    tmp.push(x);
  }
  let parent = tmp.filter((o) => o.TemplateLine.Level == 0);
  for (let i in parent) {
    let o = parent[i];
    o.Target = sumOfField(tmp, o.TemplateLine.ID, "Target");
  }
  data.record = tmp;

  refreshGridLines();
}

function sumOfField(dt, parentId, fieldsum) {
  let child = dt.filter((o) => o.TemplateLine.Parent.includes(parentId));
  return child.map((o) => o[fieldsum]).reduce((a, b) => a + b, 0);
}

function onChangeActualValue(item) {
  util.nextTickN(2, () => {
    item.ActualValue = parseInt(item.ActualValue);
    let lvl0 = data.record.find(
      (o) => o.TemplateLine.ID == item.TemplateLine.Parent
    );

    if (lvl0) {
      item.ACH = (item.ActualValue / lvl0.Target) * 100;
    } else {
      item.ACH = 0;
    }

    if (!item.Point) {
      item.ACH = 0;
      item.ActualValue = 0;
    }

    let splitParent = item.TemplateLine.Parent.split("#");
    for (let i = splitParent.length; i > 0; i--) {
      setParentValue(splitParent[i - 1], i);
    }
  });
}

function formatGridLayout(selector) {
  if (data.record.length == 0) return;

  const thead = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_header']`
  );

  thead[0].insertRow(1);
  thead[0].rows[0].innerHTML = `<tr>
    <th rowspan='2' class='w-[50px]'>No</th>
    <th rowspan='2'>Uraian</th>
    <th rowspan='2' class='w-[150px]'>Dokument/Bukti lain pendukung jawaban</th>
    <th rowspan='2' class='w-[200px]'>Dokument dilampirkan</th>
    <th colspan='5'>Aspek Pemenuhan Tiap Elemen</th>
  </tr>`;

  thead[0].rows[1].innerHTML = `<tr>
    <th class='w-[80px]' style="border-left:1px solid #D1D5DC">Point</th>
    <th class='w-[80px]'>Target</th>
    <th class='w-[80px]'>Actual</th>
    <th class='w-[80px]'>ACH</th>
    <th class='w-[200px]'>Temuan Ketidaksesuaian</th>
  </tr>`;

  const tbody = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );
  for (var i = 0, row; (row = tbody[0].rows[i]); i++) {
    row.classList.add("tr-smkpau-" + data.record[i].TemplateLine.Level);

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
      formatGridLayout("grid_Audit_SMKPAU");
    }, 500);
  });
}

function setParentValue(id, level) {
  let idx = data.record.findIndex((o) => o.TemplateLine.ID == id);
  let parent = data.record[idx];
  let child = data.record.filter(
    (o) =>
      o.TemplateLine.Parent.includes(id) &&
      o.TemplateLine.Level == level &&
      o.Point
  );

  let res = {
    ActualValue: 0,
    ACH: 0,
  };

  for (let ky in res) {
    res[ky] = child.map((o) => o[ky]).reduce((a, b) => a + b, 0);
    parent[ky] = isNaN(res[ky]) ? 0 : res[ky];
  }
}

const totalTarget = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.Target)
      .reduce((a, b) => a + b, 0);
  },
});

const totalActual = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.ActualValue)
      .reduce((a, b) => a + b, 0);
  },
});

const totalAch = computed({
  get() {
    let div = data.record.filter((o) => o.TemplateLine.Level == 0);
    let sumOfAch = data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.ACH)
      .reduce((a, b) => a + b, 0);
    let res = sumOfAch / div.length;
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
  loadGridConfig(axios, "/she/audit/smkpau/gridconfig").then(
    (r) => {
      data.cfg = r;
    },
    (e) => util.showError(e)
  );
  getDataByTemplate(props.templateId);
});
</script>

<style>
.grid_Audit_SMKPAU .tr-smkpau-0 td,
.grid_Audit_SMKPAU .smkpou-total div {
  background: #ddebf7;
}

.grid_Audit_SMKPAU .tr-smkpau-0 {
  @apply font-semibold;
}
</style>
