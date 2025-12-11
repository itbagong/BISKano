<template>
  <div class="card w-full">
    <div class="flex mb-4 gap-4">
      <s-button
        icon="rewind"
        class="btn_warning back_btn w-20 mt-3"
        label="back"
        @click="emit('cancel')"
      />
      <s-button
        icon="fileMove"
        class="btn_success w-30 mt-3"
        label="Move Asset"
        @click="moveAsset"
      />
    </div>
    <div class="w-full">
      <s-input
        label="Sites"
        v-model="data.filteredSites"
        useList
        lookup-key="_id"
        :lookup-labels="['_id', 'Name']"
        :lookupSearchs="['_id', 'Name']"
        lookupUrl="/bagong/sitesetup/find"
        multiple
        @change="buildData"
      />
    </div>

    <div class="flex font-semibold">
      <div class="w-40 p-2" v-for="(dt, idx) in data.filteredSites" :key="idx">
        <s-input
          v-model="data.filteredSites[idx]"
          useList
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['_id', 'Name']"
          lookupUrl="/bagong/sitesetup/find"
          read-only
        />
      </div>
    </div>

    <div class="flex overflow-x-auto h-[730px]">
      <div
        class="flex-none p-2 w-40 relative"
        v-for="(dt, idx) in data.record"
        :key="idx"
      >
        <div v-for="(aa, idx2) in dt.assets" :key="idx2">
          <div class="flex gap-2">
            <s-input kind="checkbox" v-model="aa.IsChecked" />
            {{ aa.AssetName }} - {{ aa.CustomerFromName ?? "" }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import helper from "@/scripts/helper.js";

const axios = inject("axios");

const data = reactive({
  record: [],
  loading: false,
  filteredSites: [],
});

const emit = defineEmits({
  cancel: null,
  movedAsset: null,
});

function buildData() {
  const url = "/bagong/asset-movement/get-assets";
  let x = {
    siteID: "",
    assets: [],
  };
  util.nextTickN(3, () => {
    let payload = {
      Where: {
        Op: "$and",
        Items: [
          { Op: "$in", Field: "Dimension.Value", Value: data.filteredSites },
          { Op: "$eq", Field: "Dimension.Key", Value: "Site" },
        ],
      },
    };
    data.record = [];
    axios.post(url, payload).then(
      (r) => {
        const result = Object.groupBy(r.data, ({ SiteFrom }) => SiteFrom);
        for (const ky in result) {
          let obj = helper.cloneObject(x);
          obj.siteID = ky;
          obj.assets = result[ky];
          data.record.push(obj);
        }
        data.loading = false;
      },
      (e) => {
        util.showError(e);
      }
    );
  });
}

function moveAsset() {
  let res = [];
  for (let i in data.record) {
    let o = data.record[i];
    let tmp = o.assets.filter((x) => x.IsChecked == true);
    res = [...res, ...tmp];
  }
  emit("movedAsset", res);
}
</script>
