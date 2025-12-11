<template>
  <div class="relative w-full mb-2">
    <input
      type="text"
      placeholder="Search for a place..."
      v-model="searchQuery"
      @input="debouncedFetchSuggestions"
      class="input_field"
    />
    <ul
      v-if="suggestions.length > 0 && searchQuery !== ''"
      class="absolute bg-white border rounded w-full z-[2000] max-h-40 overflow-y-auto"
    >
      <li
        v-for="(suggestion, index) in suggestions"
        :key="index"
        @click="selectSuggestion(suggestion)"
        class="p-2 cursor-pointer hover:bg-gray-200"
      >
        {{ suggestion.label }}
      </li>
    </ul>
  </div>
  <div :key="data.key" class="h-[250px] w-full rounded overflow-hidden">
    <l-map
      ref="refMap"
      v-model="data.zoom"
      v-model:zoom="data.zoom"
      :center="data.center"
      @click="addMarker"
      class="cursor-auto"
      :use-global-leaflet="false"
    >
      <l-tile-layer :url="url" />
      <l-marker
        v-if="data.latlng.lat && data.latlng.lng"
        draggable
        :lat-lng="data.latlng"
        refre
      ></l-marker>
    </l-map>
  </div>
</template>
<script setup>
import { reactive, onMounted, ref } from "vue";

import { LMap, LTileLayer, LMarker, LTooltip } from "@vue-leaflet/vue-leaflet";
import "leaflet/dist/leaflet.css";
import { OpenStreetMapProvider } from "leaflet-geosearch";
import "leaflet-geosearch/dist/geosearch.css"; // Optional styling for better appearance

import { util } from "suimjs";

const props = defineProps({
  long: { type: Number, default: 0 },
  lat: { type: Number, default: 0 },
});

const emit = defineEmits({
  onChangeMarker: null,
});
const refMap = ref(null);
const searchQuery = ref("");
const suggestions = ref([]);
const provider = new OpenStreetMapProvider();
let debounceTimeout = null;

const data = reactive({
  zoom: 10,
  latlng: {
    lng: props.long,
    lat: props.lat,
  },
  center: [-7.966743161967913, 112.63401370760941],
  key: Math.random(),
});

const url = "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png";

const addMarker = (e) => {
  console.log(e.latlng);
  data.latlng = e.latlng;
  emit("onChangeMarker", e.latlng);
};
const fetchSuggestions = async () => {
  if (!searchQuery.value) {
    suggestions.value = [];
    return;
  }

  try {
    const results = await provider.search({ query: searchQuery.value });
    suggestions.value = results.map((res) => ({
      label: res.label,
      lat: res.y,
      lng: res.x,
    }));
  } catch (error) {
    console.error("Error fetching suggestions:", error);
  }
};
const debouncedFetchSuggestions = () => {
  clearTimeout(debounceTimeout);
  debounceTimeout = setTimeout(() => {
    fetchSuggestions();
  }, 300); // Adjust the debounce delay as needed
};
const selectSuggestion = (suggestion) => {
  data.latlng = { lat: suggestion.lat, lng: suggestion.lng };
  data.center = [suggestion.lat, suggestion.lng];
  emit("onChangeMarker", data.latlng);
  searchQuery.value = suggestion.label; // Update input with the selected place name
  suggestions.value = []; // Clear suggestions after selection
};
onMounted(() => {
  if (props.lat && props.long) {
    data.center = [props.lat, props.long];
    util.nextTickN(5, () => {
      data.key = Math.random();
    });
  }
});
</script>
<style scoped>
ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
</style>
