<template>
  <s-form
    v-model="data.record"
    :config="data.formCfg"
    :mode="readOnly ? 'view' : 'edit'"
    keep-label
    hide-cancel
    hide-submit
    ref="formRef"
  >
    <template #input_ItemTemplateID="{ item }">
      <s-input
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
    <template #input_Details="{ item }">
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border border-gray-300">
              <th class="border border-gray-300 px-4 py- text-center">NO</th>
              <th class="border border-gray-300 px-4 py-1 text-center">
                ASPEK YANG DINILAI
              </th>
              <th class="border border-gray-300 px-4 py-1 text-center">
                BOBOT (%)
              </th>
              <th
                colspan="5"
                class="border border-gray-300 px-4 py-1 text-center"
              >
                BOBOT (%)
              </th>
              <th class="border border-gray-300 px-4 py-1 text-center">
                Total (Bobot * Nilai)
              </th>
              <th class="border border-gray-300 px-4 py-1 text-center">Note</th>
            </tr>
            <tr class="border border-gray-300">
              <th class="px-4 py-1 text-center"></th>
              <th class="px-4 py-1 text-center"></th>
              <th class="px-4 py-1 text-center"></th>
              <th class="border border-gray-300 px-4 py-1 text-center">1</th>
              <th class="border border-gray-300 px-4 py-1 text-center">2</th>
              <th class="border border-gray-300 px-4 py-1 text-center">3</th>
              <th class="border border-gray-300 px-4 py-1 text-center">4</th>
              <th class="border border-gray-300 px-4 py-1 text-center">5</th>
              <th class="border border-gray-300 px-4 py-1 text-center"></th>
              <th class="border border-gray-300 px-4 py-1 text-center"></th>
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
          <tbody v-else-if="data.record?.Details?.length > 0">
            <template
              v-for="(section, key) in item.Details"
              :key="`row-${key}`"
            >
              <tr>
                <td colspan="10" class="border border-gray-300 px-4 py-2">
                  {{ section.Section }}
                </td>
              </tr>
              <tr
                v-for="(detail, detailIndex) in section.Details"
                :key="`${key}-${detailIndex}`"
              >
                <td class="px-4 py-2 border">{{ detailIndex + 1 }}</td>
                <td class="px-4 py-2 border">{{ detail.Subject }}</td>
                <td class="px-4 py-2 border text-center">
                  {{ detail.Weight }}
                </td>
                <td
                  v-for="value in 5"
                  :key="value"
                  class="px-4 py-2 border text-center"
                >
                  <template v-if="readOnly">
                    <span v-if="detail.Value === value">✔️</span>
                  </template>
                  <template v-else>
                    <input
                      type="radio"
                      :name="`value-${key}-${detailIndex}`"
                      :value="value"
                      :disabled="readOnly"
                      v-model="detail.Value"
                      @change="calculateScore(detail)"
                    />
                  </template>
                </td>
                <td class="px-4 py-2 border text-center">{{ detail.Score }}</td>
                <td class="px-4 py-2 border">{{ detail.Note }}</td>
              </tr>
            </template>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold"
              ></td>
              <td
                class="border border-gray-300 px-4 py-1 font-bold text-center"
              >
                {{ totalWeight }}
              </td>
              <td
                colspan="5"
                class="border border-gray-300 px-4 py-1 font-bold"
              >
                Total Nilai
              </td>
              <td
                class="border border-gray-300 px-4 py-1 font-bold text-center"
              >
                {{ totalScore }}
              </td>
              <td class="border border-gray-300 px-4 py-1 font-bold"></td>
            </tr>
          </tfoot>
        </table>
      </div>
    </template>
    <template #input_Start="{ item }">
      <div>
        <div class="mb-2">Kesimpulan / Rekomendasi atasan</div>
        <div>
          <s-input
            v-model="item.Conclusion"
            :read-only="readOnly"
            label="Conclusion"
            keep-label
            use-list
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=InterviewPromotionConclusion"
            lookup-key="_id"
            :lookup-labels="['Name']"
            class="mb-2"
          />
        </div>
        <div class="grid grid-cols-3 gap-2 items-end mb-2">
          <s-input
            v-if="item.Conclusion === 'IPC001'"
            class="w-full"
            label="Start"
            kind="date"
            v-model="item.Start"
            :read-only="readOnly"
          />
          <s-input
            v-if="item.Conclusion === 'IPC002'"
            class="w-full"
            label="Period"
            kind="date"
            :read-only="readOnly"
            v-model="item.PeriodFrom"
          />
          <s-input
            v-if="item.Conclusion === 'IPC002'"
            class="w-full"
            label=""
            kind="date"
            :read-only="readOnly"
            v-model="item.PeriodTo"
          />
        </div>
      </div>
    </template>
  </s-form>
</template>
<script setup>
import { reactive, onMounted, inject, ref, computed } from "vue";
import { SForm, SInput, loadFormConfig, util } from "suimjs";

const axios = inject("axios");
const formRef = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: {} },
  readOnly: { type: Boolean, default: false },
});
const data = reactive({
  loading: false,
  formCfg: {},
  record: props.modelValue,
});

function editRecord(record) {}

function renderItemDetails(v, record) {
  if (v) {
    data.loading = true;
    axios
      .post("/she/mcuitemtemplate/get", [v])
      .then((r) => {
        record.Details = r.data.Lines.filter((o) => o.Parent === "").map(
          (item) => ({
            Section: item.Description,
            Details: r.data.Lines.filter((o) => o.Parent === item.ID).map(
              (child) => ({
                Subject: child.Description.trim(),
                Weight: child.AnswerValue,
                Value: 0,
                Score: 0,
                Note: child.Note.trim(),
              })
            ),
          })
        );
        // record.ItemDetails = r.data.Lines.map((item) => {
        //   return {
        //     Aspect: item.Description,
        //     MaxScore: item.AnswerValue,
        //     AchievedScore: 0,
        //   };
        // });
      })
      .catch((e) => {
        util.showError(e);
      })
      .finally(() => {
        data.loading = false;
      });
  } else {
    record.Details = [];
  }
}
function calculateScore(detail) {
  detail.Score = parseFloat(((detail.Weight / 100) * detail.Value).toFixed(2));
}
const totalScore = computed(() => {
  const result = data.record?.Details?.reduce(
    (sectionTotal, section) =>
      sectionTotal +
      section.Details.reduce(
        (detailTotal, detail) => detailTotal + parseFloat(detail.Score || 0),
        0
      ),
    0
  );
  data.record.Score = result ?? 0;
  return result;
});

const totalWeight = computed(() => {
  const totalWeightSum = data.record?.Details?.reduce(
    (sectionTotal, section) =>
      sectionTotal +
      section.Details.reduce(
        (detailTotal, detail) => detailTotal + detail.Weight,
        0
      ),
    0
  );
  // const totalDetails = data.record?.Details.reduce(
  //   (count, section) => count + section.Details.length,
  //   0
  // );
  // const result = totalDetails > 0 ? parseFloat((totalWeightSum / totalDetails).toFixed(2)) : 0
  data.record.TotalWeight = totalWeightSum;
  return totalWeightSum;
});

onMounted(() => {
  editRecord(data.record);
  loadFormConfig(
    axios,
    "/hcm/talentdevelopmentassesment/interview/formconfig"
  ).then(
    (r) => {
      data.formCfg = r;
      util.nextTickN(2, () => {});
    },
    (e) => util.showError(e)
  );
});
</script>
<style></style>
