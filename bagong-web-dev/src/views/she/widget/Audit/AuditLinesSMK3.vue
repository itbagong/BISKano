<template>
  <s-grid
    class="grid_Audit_SMK3"
    ref="listControl"
    :config="data.cfgSMK3"
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
    <template #item_Point="{ item }">
      <div v-if="item.TemplateLine.Level < 2" class="text-right py-1.5">
        {{ item.Point }}
      </div>
      <s-input
        v-model="item.Point"
        use-list
        :items="[
          { key: 0, text: '0' },
          { key: 1, text: '1' },
        ]"
        v-else
        :show-clear-button="false"
        @change="onChangePoint(item)"
      />
    </template>
    <template #item_Score="{ item }">
      <div class="text-right py-1.5">
        {{ item.Score }}
      </div>
    </template>
    <template #item_ACH="{ item }">
      <div class="text-right py-1.5">
        {{ formatNumDecimal(item.ACH) }}
        %
      </div>
    </template>
    <template #item_Pica="{ item }">
      <div class="w-20 flex justify-center">
        <mdicon
          name="folder"
          size="16"
          @click="item.UsePica = true"
          v-if="item.Point == 0 && item.TemplateLine.Level == 2"
        />
      </div>
      <s-modal
        v-if="item.UsePica"
        display
        hideButtons
        @beforeHide="item.UsePica = false"
      >
        <s-card hide-title class="min-w-[350px]">
          <Pica v-model="item.Pica" />
          <s-button
            class="btn_secondary"
            @click="item.UsePica = false"
            label="Save"
          ></s-button>
        </s-card>
      </s-modal>
    </template>
    <template #paging="{ item }">
      <div class="grid grid-cols-3 font-semibold text-center border smk3-total">
        <div class="col-span-2 border-b p-1">TOTAL POINT</div>
        <div class="border-b p-1">{{ totalPoint }}</div>
        <div class="col-span-2 border-b p-1">TOTAL SCORE</div>
        <div class="border-b p-1">{{ totalScore }}</div>
        <div class="col-span-2 border-b p-1">
          TOTAL PENCAPAIAN SISTEM MANAJEMEN KESELAMATAN, KESEHATAN KERJA (K3)
        </div>
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
  cfgSMK3: {},
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
      Point: 0,
      Score: 0,
      ACH: 0,
      UsePica: false,
      Pica: {},
      TemplateLine: o,
    };
    tmp.push(x);
  }
  data.record = tmp;
  refreshGridLines();
}

function formatGridLayout(selector) {
  if (data.record.length == 0 || !props.templateId) return;

  const thead = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_header']`
  );

  thead[0].insertRow(1);
  thead[0].rows[0].innerHTML = `<tr>
    <th rowspan='2' class='w-[50px]'>No</th>
    <th rowspan='2'>Element AUDIT</th>
    <th colspan='2'>Pemenuhan Kriteria</th>
    <th rowspan='2' class='w-[100px]'>ACH</th>
    <th rowspan='2' class='w-[100px]'>Pica</th>
  </tr>`;

  thead[0].rows[1].innerHTML = `<tr>
    <th class='w-[100px]' style="border-left:1px solid #D1D5DC">Point</th>
    <th class='w-[100px]'>Score</th>
  </tr>`;

  const tbody = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );
  for (var i = 0, row; (row = tbody[0].rows[i]); i++) {
    row.classList.add("tr-smk3-" + data.record[i].TemplateLine.Level);

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
      formatGridLayout("grid_Audit_SMK3");
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

function setAch(parentID, score) {
  if (!parentID) return 0;

  let child = data.record.filter(
    (o) => o.TemplateLine.Parent.includes(parentID) && o.TemplateLine.Level == 2
  );

  let sumScore = child.map((o) => o.Score).reduce((a, b) => a + b, 0);
  for (let i in child) {
    let o = child[i];
    o.ACH = (o.Score / sumScore) * 100;
  }
}

function setParentValue(id, level) {
  let idx = data.record.findIndex((o) => o.TemplateLine.ID == id);
  let parent = data.record[idx];
  let child = data.record.filter(
    (o) => o.TemplateLine.Parent.includes(id) && o.TemplateLine.Level == level
  );

  let res = {
    Point: 0,
    Score: 0,
    ACH: 0,
  };

  for (let ky in res) {
    res[ky] = child.map((o) => o[ky]).reduce((a, b) => a + b, 0);
    if (parent.TemplateLine.Level == 0 && ky == "ACH") {
      parent[ky] = res[ky] / child.length;
    } else {
      parent[ky] = res[ky];
    }
  }
}

const totalPoint = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.Point)
      .reduce((a, b) => a + b, 0);
  },
});

const totalScore = computed({
  get() {
    return data.record
      .filter((o) => o.TemplateLine.Level == 0)
      .map((o) => o.Score)
      .reduce((a, b) => a + b, 0);
  },
});

const totalPencapainan = computed({
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
  loadGridConfig(axios, "/she/audit/smk3/gridconfig").then(
    (r) => {
      data.cfgSMK3 = r;
    },
    (e) => util.showError(e)
  );
  getDataByTemplate(props.templateId);
});
</script>

<style>
.grid_Audit_SMK3 .tr-smk3-0 td,
.grid_Audit_SMK3 .smk3-total div {
  background: #bfbfbf;
}

.grid_Audit_SMK3 .tr-smk3-1 td {
  background: #ddebf7;
}

.grid_Audit_SMK3 .tr-smk3-0,
.grid_Audit_SMK3 .tr-smk3-1 {
  @apply font-semibold;
}
</style>
