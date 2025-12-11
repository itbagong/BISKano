<template>
  <s-form v-model="data.record" :config="data.formCfg" keep-label hide-cancel hide-submit ref="formRef" :mode="readOnly ? 'view': 'edit'">
    <template #input_Status="{ item }">
      <s-input label="Status" keep-label v-model="data.gerenalStatus" read-only kind="text" class="w-full px-2 py-1" />
    </template>
    <template #input_Attendace="{ item }">
      <div class="mb-5">
        <s-input
          label="Journal type ID"
          v-model="data.journalTypeID"
          class="w-1/2"
          :required="true"
          use-list
          :read-only="readOnly"
          :lookup-url="`/hcm/journaltype/find?TransactionType=Talent%20Development%20-%20Promotion%20-%20Tracking%20Assessment`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="handleJounalTypeChange"
        ></s-input>
      </div>
      <div class="title section_title">Attendance</div>
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>Absent Name</th>
              <th
                v-for="(month, index) in item.Months"
                :key="`header-${index}`"
              >
                {{ month }}
              </th>
              <th>Total</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(category, key) in data.otherCategories"
              :key="`row-${key}`"
            >
              <td class="border border-gray-300 px-4 py-2">{{ category }}</td>
              <td
                v-for="(entry, index) in item?.Attendace[key]"
                :key="`entry-${key}-${index}`"
                class="border border-gray-300 px-4"
              >
                <s-input
                  v-model.number="entry.Score"
                  kind="number"
                  class="w-full px-2 py-1"
                  :read-only="readOnly"
                />
              </td>
              <td class="border border-gray-300 px-4 py-2 font-semibold">
                {{ computeTotal(data.record.Attendace[key]) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
    <template #input_ItemTemplateID="{ item }">
      <s-input
        class="w-[300px]"
        label="Item Template ID"
        use-list
        lookup-url="/she/mcuitemtemplate/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.ItemTemplateID"
        :read-only="readOnly"
        @change="(_, v) => renderItemDetails(v, item)"
      />
    </template>
    <template #input_ItemDetails="{ item }">
      <div class="title section_title">Performance</div>
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>Aspect</th>
              <th>Max Score</th>
              <th>Achieved Score</th>
            </tr>
          </thead>
          <tbody v-if="data.loading">
            <tr>
              <td colspan="7">
                <div class="h-[300px] flex items-center justify-center">
                  <Loader kind="circle" />
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="(detail, key) in item.ItemDetails" :key="`row-${key}`">
              <td class="border border-gray-300 px-4 py-1">
                {{ detail.Aspect }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                {{ detail.MaxScore }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="detail.AchievedScore"
                  kind="number"
                  class="px-2"
                  :read-only="readOnly"
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td class="border border-gray-300 px-4 py-1 font-bold">
                Total Score
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ maxScoreTotal }}
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ achievedScoreTotal }}
              </td>
            </tr>
            <tr>
              <td
                class="border border-gray-300 px-4 py-1 font-bold"
                colspan="2"
              >
                Final Score
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ finalScoreTotal }}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </template>
    <template v-if="readOnly" #input_IsProbationEnd="{ item, config }">
      <div class="flex gap-2 items-center mt-1">
        <div>
          <input
            type="checkbox"
            v-model="item.IsProbationEnd"
            ref="control"
            disabled
          />
        </div>
        <div>{{ config.label }}</div>
      </div>
    </template>
    <template v-if="readOnly" #input_IsBecomeEmployee="{ item, config }">
      <div class="flex gap-2 items-center mt-1">
        <div>
          <input
            type="checkbox"
            v-model="item.IsBecomeEmployee"
            ref="control"
            disabled
          />
        </div>
        <div>{{ config.label }}</div>
      </div>
    </template>
    <template v-if="readOnly" #input_IsPromoted="{ item, config }">
      <div class="flex gap-2 items-center mt-1">
        <div>
          <input
            type="checkbox"
            v-model="item.IsPromoted"
            ref="control"
            disabled
          />
        </div>
        <div>{{ config.label }}</div>
      </div>
    </template>
  </s-form>
</template>
<script setup>
import { reactive, onMounted, inject, ref, computed, watch } from "vue";
import { SForm, SInput, loadFormConfig, util } from "suimjs";
import Loader from "@/components/common/Loader.vue";

const axios = inject("axios");
const formRef = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: {} },
  journalTypeID: { type: String, default: "" },
  status: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
});
const emit = defineEmits({
  journalTypeChange: null,
  "update:journalTypeID": null,
});
const data = reactive({
  formCfg: {},
  record: props.modelValue,
  months: ["Month 1", "Month 2", "Month 3", "Month 4", "Month 5", "Month 6"],
  otherCategories: {
    Presences: "Kehadiran",
    Absents: "Alpha/Mangkir",
    Sicks: "Sakit",
    Leaves: "Izin",
    Lates: "Terlambat",
  },
  journalTypeID: props.journalTypeID,
  gerenalStatus: props.status
});
function editRecord(record) {
  if (!record?.Attendace || record?.Attendace?.Presences?.length === 0) {
    const initialData = data.months.map((month) => ({ Name: month, Score: 0 }));
    record.Attendace = {
      Presences: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Absents: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Sicks: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Leaves: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Lates: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
    };
  }
  record.Months = [...data.months];
  data.record = record;
}
function computeTotal(entries) {
  return entries?.reduce((sum, entry) => sum + entry.Score, 0);
}
function renderItemDetails(v, record) {
  data.loading = true;
  axios
    .post("/she/mcuitemtemplate/get", [v])
    .then((r) => {
      record.ItemDetails = r.data.Lines.map((item) => {
        return {
          Aspect: item.Description,
          MaxScore: item.AnswerValue,
          AchievedScore: 0,
        };
      });
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}
const maxScoreTotal = computed(() => {
  const result = data.record?.ItemDetails?.reduce(
    (total, aspect) => total + aspect.MaxScore,
    0
  );
  data.record.MaxScoreTotal = result ?? 0;
  return result;
});
const achievedScoreTotal = computed(() => {
  const result = data.record?.ItemDetails?.reduce(
    (total, aspect) => total + aspect.AchievedScore,
    0
  );
  data.record.AchievedScoreTotal = result ?? 0;
  return result;
});
const finalScoreTotal = computed(() => {
  if (maxScoreTotal.value === 0) {
    return 0;
  }
  const result = parseFloat(
    ((achievedScoreTotal.value / maxScoreTotal.value) * 100).toFixed(2)
  );
  data.record.FinalScore = isNaN(result) ? 0 : result;
  return data.record.FinalScore;
});
onMounted(() => {
  editRecord(data.record)
  loadFormConfig(axios, "/hcm/talentdevelopmentassesment/assesment/formconfig").then(
    (r) => {
      data.formCfg = r;
      util.nextTickN(2, () => { });
    },
    (e) => util.showError(e)
  );
});
function handleJounalTypeChange(_, v1) {
  emit("journalTypeChange", v1);
}
watch(
  () => data.journalTypeID,
  (nv) => {
    emit("update:journalTypeID", nv);
  },
  { deep: true }
);
</script>
<style></style>